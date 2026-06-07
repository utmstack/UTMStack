package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"gorm.io/gorm"
)

type pgRuleRepository struct {
	db *gorm.DB
}

func NewRuleRepository(db *gorm.DB) connectors.RuleRepository {
	return &pgRuleRepository{db: db}
}

func (r *pgRuleRepository) Create(ctx context.Context, rule *domain.AlertResponseRule) (*domain.AlertResponseRule, error) {
	// Wrap rule insert + template sync in a single transaction
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(rule).Error; err != nil {
			if isUniqueViolation(err) {
				return domain.ErrRuleNameTaken
			}
			return err
		}
		return tx.Model(rule).Association("Templates").Replace(rule.Templates)
	})
	if err != nil {
		return nil, err
	}
	return rule, nil
}

func (r *pgRuleRepository) Update(ctx context.Context, rule *domain.AlertResponseRule) (*domain.AlertResponseRule, error) {
	// Wrap rule save + template sync in a single transaction
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(rule).Error; err != nil {
			if isUniqueViolation(err) {
				return domain.ErrRuleNameTaken
			}
			return err
		}
		return tx.Model(rule).Association("Templates").Replace(rule.Templates)
	})
	if err != nil {
		return nil, err
	}
	return rule, nil
}

func (r *pgRuleRepository) GetByID(ctx context.Context, id int64) (*domain.AlertResponseRule, error) {
	var rule domain.AlertResponseRule
	err := r.db.WithContext(ctx).
		Preload("Templates").
		First(&rule, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrRuleNotFound
		}
		return nil, err
	}
	return &rule, nil
}

func (r *pgRuleRepository) List(ctx context.Context, f connectors.RuleFilters) ([]domain.AlertResponseRule, int64, error) {

	q := r.db.WithContext(ctx).Model(&domain.AlertResponseRule{}).Preload("Templates")

	// id.equals
	if f.ID != 0 {
		q = q.Where("id = ?", f.ID)
	}
	// name.contains
	if f.RuleName != "" {
		q = q.Where("rule_name ILIKE ?", "%"+f.RuleName+"%")
	}
	// active.equals
	if f.RuleActive != nil {
		q = q.Where("rule_active = ?", *f.RuleActive)
	}
	// agentPlatform.equals
	if f.AgentPlatform != "" {
		q = q.Where("agent_platform = ?", f.AgentPlatform)
	}
	// createdBy.equals
	if f.CreatedBy != "" {
		q = q.Where("created_by = ?", f.CreatedBy)
	}
	// lastModifiedBy.equals
	if f.LastModifiedBy != "" {
		q = q.Where("last_modified_by = ?", f.LastModifiedBy)
	}
	// createdDate.greaterThanOrEqual
	if f.CreatedDateGTE != "" {
		q = q.Where("created_date >= ?", f.CreatedDateGTE)
	}
	// createdDate.lessThanOrEqual
	if f.CreatedDateLTE != "" {
		q = q.Where("created_date <= ?", f.CreatedDateLTE)
	}
	// lastModifiedDate.greaterThanOrEqual
	if f.LastModifiedDateGTE != "" {
		q = q.Where("last_modified_date >= ?", f.LastModifiedDateGTE)
	}
	// lastModifiedDate.lessThanOrEqual
	if f.LastModifiedDateLTE != "" {
		q = q.Where("last_modified_date <= ?", f.LastModifiedDateLTE)
	}
	// systemOwner.equals
	if f.SystemOwner != nil {
		q = q.Where("system_owner = ?", *f.SystemOwner)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rules []domain.AlertResponseRule
	if err := q.Order("id ASC").
		Offset(f.Offset()).
		Limit(f.Limit()).
		Find(&rules).Error; err != nil {
		return nil, 0, err
	}
	return rules, total, nil
}

func (r *pgRuleRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&domain.AlertResponseRule{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrRuleNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "23505") ||
		strings.Contains(err.Error(), "duplicate key")
}
