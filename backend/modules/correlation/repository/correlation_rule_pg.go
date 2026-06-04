package repository

import (
	"context"
	"errors"
	"time"

	"github.com/utmstack/utmstack/backend/modules/correlation/connectors"
	"github.com/utmstack/utmstack/backend/modules/correlation/domain"
	correrrors "github.com/utmstack/utmstack/backend/modules/correlation/errors"
	"gorm.io/gorm"
)

type pgCorrelationRuleRepository struct {
	db *gorm.DB
}

func NewCorrelationRuleRepository(db *gorm.DB) connectors.CorrelationRuleRepository {
	return &pgCorrelationRuleRepository{db: db}
}

func (r *pgCorrelationRuleRepository) Create(ctx context.Context, rule *domain.UtmCorrelationRules) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		rule.RuleLastUpdate = &now

		dataTypes := rule.DataTypes
		rule.DataTypes = nil

		if err := tx.Create(rule).Error; err != nil {
			return err
		}

		return r.SyncDataTypes(ctx, tx, rule.ID, dataTypes)
	})
}

func (r *pgCorrelationRuleRepository) Update(ctx context.Context, rule *domain.UtmCorrelationRules) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		rule.RuleLastUpdate = &now

		dataTypes := rule.DataTypes
		rule.DataTypes = nil

		if err := tx.Save(rule).Error; err != nil {
			return err
		}

		return r.SyncDataTypes(ctx, tx, rule.ID, dataTypes)
	})
}

func (r *pgCorrelationRuleRepository) GetByID(ctx context.Context, id int64) (*domain.UtmCorrelationRules, error) {
	var rule domain.UtmCorrelationRules
	err := r.db.WithContext(ctx).
		Preload("DataTypes").
		First(&rule, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rule, nil
}

func (r *pgCorrelationRuleRepository) List(ctx context.Context, f connectors.CorrelationRuleFilters) ([]domain.UtmCorrelationRules, int64, error) {
	page, size := normPage(f.Page, f.Size)

	q := r.db.WithContext(ctx).Model(&domain.UtmCorrelationRules{})

	// When filtering by data type we need a join; use a sub-query to keep
	// the count and the fetch consistent with the DISTINCT requirement.
	if len(f.DataTypes) > 0 {
		q = q.Where(`id IN (
			SELECT DISTINCT rule_id FROM utm_group_rules_data_type grdt
			JOIN utm_data_types dt ON dt.id = grdt.data_type_id
			WHERE dt.data_type IN ?
		)`, f.DataTypes)
	}

	// ruleName.contains (case-insensitive)
	if f.RuleName != "" {
		q = q.Where("rule_name ILIKE ?", "%"+f.RuleName+"%")
	}

	// search (general, same column as ruleName in Java)
	if f.Search != "" {
		q = q.Where("rule_name ILIKE ?", "%"+f.Search+"%")
	}

	// ruleActive.equals
	if f.RuleActive != nil {
		q = q.Where("rule_active = ?", *f.RuleActive)
	}

	// ruleCategory.in
	if len(f.RuleCategory) > 0 {
		q = q.Where("rule_category IN ?", f.RuleCategory)
	}

	// ruleAdversary.in
	if len(f.RuleAdversary) > 0 {
		q = q.Where("rule_adversary IN ?", f.RuleAdversary)
	}

	// ruleTechnique.in
	if len(f.RuleTechnique) > 0 {
		q = q.Where("rule_technique IN ?", f.RuleTechnique)
	}

	// ruleConfidentiality.in
	if len(f.RuleConfidentiality) > 0 {
		q = q.Where("rule_confidentiality IN ?", f.RuleConfidentiality)
	}

	// ruleIntegrity.in
	if len(f.RuleIntegrity) > 0 {
		q = q.Where("rule_integrity IN ?", f.RuleIntegrity)
	}

	// ruleAvailability.in
	if len(f.RuleAvailability) > 0 {
		q = q.Where("rule_availability IN ?", f.RuleAvailability)
	}

	// systemOwner.equals
	if f.SystemOwner != nil {
		q = q.Where("system_owner = ?", *f.SystemOwner)
	}

	// date range on rule_last_update (inclusive)
	if f.InitDate != "" && f.EndDate != "" {
		q = q.Where("rule_last_update BETWEEN ? AND ?", f.InitDate, f.EndDate)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rules []domain.UtmCorrelationRules
	if err := q.Preload("DataTypes").
		Order("id ASC").
		Offset(page * size).
		Limit(size).
		Find(&rules).Error; err != nil {
		return nil, 0, err
	}
	return rules, total, nil
}

func (r *pgCorrelationRuleRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Clear M2M association first.
		rule := &domain.UtmCorrelationRules{ID: id}
		if err := tx.Model(rule).Association("DataTypes").Clear(); err != nil {
			return err
		}

		result := tx.Delete(&domain.UtmCorrelationRules{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return correrrors.ErrCorrelationRuleNotFound
		}
		return nil
	})
}

func (r *pgCorrelationRuleRepository) ActivateDeactivate(ctx context.Context, id int64, active bool) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&domain.UtmCorrelationRules{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"rule_active":      active,
			"rule_last_update": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return correrrors.ErrCorrelationRuleNotFound
	}
	return nil
}

func (r *pgCorrelationRuleRepository) FindDistinctPropertyValues(ctx context.Context, prop, value string) ([]string, error) {
	query := `SELECT DISTINCT ` + prop + ` FROM utm_correlation_rules WHERE ` + prop + ` IS NOT NULL`
	args := []any{}
	if value != "" {
		query += ` AND LOWER(` + prop + `) LIKE ?`
		args = append(args, "%"+value+"%")
	}
	query += ` ORDER BY ` + prop + ` ASC`

	var results []string
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&results).Error; err != nil {
		return nil, err
	}
	if results == nil {
		results = []string{}
	}
	return results, nil
}

func (r *pgCorrelationRuleRepository) FindByRuleName(ctx context.Context, ruleName string) (*domain.UtmCorrelationRules, error) {
	var rule domain.UtmCorrelationRules
	err := r.db.WithContext(ctx).
		Preload("DataTypes").
		Where("rule_name = ?", ruleName).
		First(&rule).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rule, nil
}

func (r *pgCorrelationRuleRepository) FindAllBySystemOwner(ctx context.Context, systemOwner bool) ([]domain.UtmCorrelationRules, error) {
	var rules []domain.UtmCorrelationRules
	if err := r.db.WithContext(ctx).
		Where("system_owner = ?", systemOwner).
		Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *pgCorrelationRuleRepository) SyncDataTypes(
	ctx context.Context,
	tx *gorm.DB,
	ruleID int64,
	dataTypes []domain.UtmDataTypes,
) error {
	resolved := make([]domain.UtmDataTypes, 0, len(dataTypes))

	for _, dt := range dataTypes {
		if dt.ID > 0 {
			// Check if the record actually exists.
			var existing domain.UtmDataTypes
			err := tx.WithContext(ctx).First(&existing, dt.ID).Error
			if err == nil {
				// Reuse.
				resolved = append(resolved, existing)
				continue
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			// Not found → fall through to create.
		}

		// Create new data type with systemOwner=false.
		now := time.Now().UTC()
		newDT := domain.UtmDataTypes{
			DataType:            dt.DataType,
			DataTypeName:        dt.DataTypeName,
			DataTypeDescription: dt.DataTypeDescription,
			Included:            dt.Included,
			LastUpdate:          &now,
			SystemOwner:         false,
		}
		if err := tx.WithContext(ctx).Create(&newDT).Error; err != nil {
			return err
		}
		resolved = append(resolved, newDT)
	}

	// Replace the full M2M association atomically.
	rule := &domain.UtmCorrelationRules{ID: ruleID}
	return tx.WithContext(ctx).Model(rule).Association("DataTypes").Replace(resolved)
}
