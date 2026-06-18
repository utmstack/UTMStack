package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
	"gorm.io/gorm"
)

type pgAlertTagRuleRepository struct {
	db *gorm.DB
}

func NewAlertTagRuleRepository(db *gorm.DB) connectors.AlertTagRuleRepository {
	return &pgAlertTagRuleRepository{db: db}
}

func (r *pgAlertTagRuleRepository) Create(ctx context.Context, rule domain.UtmAlertTagRule) (*domain.UtmAlertTagRule, error) {
	if err := r.db.WithContext(ctx).Create(&rule).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrRuleNameTaken
		}
		return nil, err
	}
	return &rule, nil
}

func (r *pgAlertTagRuleRepository) Update(ctx context.Context, rule domain.UtmAlertTagRule) (*domain.UtmAlertTagRule, error) {
	if err := r.db.WithContext(ctx).Save(&rule).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrRuleNameTaken
		}
		return nil, err
	}
	return &rule, nil
}

func (r *pgAlertTagRuleRepository) List(ctx context.Context, f dto.AlertTagRuleFilters) ([]domain.UtmAlertTagRule, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Size < 1 || f.Size > 200 {
		f.Size = 20
	}

	where, args := buildTagRuleWhere(f)

	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM utm_alert_tag_rule WHERE %s`, where)
	var total int64
	if err := r.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.Size
	dataSQL := fmt.Sprintf(
		`SELECT * FROM utm_alert_tag_rule WHERE %s ORDER BY id ASC LIMIT %d OFFSET %d`,
		where, f.Size, offset,
	)
	var rows []domain.UtmAlertTagRule
	if err := r.db.WithContext(ctx).Raw(dataSQL, args...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *pgAlertTagRuleRepository) GetByID(ctx context.Context, id uint64) (*domain.UtmAlertTagRule, error) {
	var row domain.UtmAlertTagRule
	if err := r.db.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgAlertTagRuleRepository) FindByIDs(ctx context.Context, ids []int64) ([]domain.UtmAlertTagRule, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []domain.UtmAlertTagRule
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *pgAlertTagRuleRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&domain.UtmAlertTagRule{}, id).Error
}

func (r *pgAlertTagRuleRepository) FindAllActive(ctx context.Context) ([]domain.UtmAlertTagRule, error) {
	var rows []domain.UtmAlertTagRule
	if err := r.db.WithContext(ctx).
		Where("rule_active = true AND rule_deleted = false").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func buildTagRuleWhere(f dto.AlertTagRuleFilters) (string, []any) {
	parts := []string{"1=1"}
	var args []any

	if f.ID != nil {
		parts = append(parts, "id = ?")
		args = append(args, *f.ID)
	}
	if f.Name != nil && *f.Name != "" {
		parts = append(parts, "lower(rule_name) LIKE lower(?)")
		args = append(args, "%"+*f.Name+"%")
	}
	if f.ConditionField != nil && *f.ConditionField != "" {
		parts = append(parts, "lower(rule_conditions) LIKE lower(?)")
		args = append(args, "%"+*f.ConditionField+"%")
	}
	if f.ConditionValue != nil && *f.ConditionValue != "" {
		parts = append(parts, "lower(rule_conditions) LIKE lower(?)")
		args = append(args, "%"+*f.ConditionValue+"%")
	}
	if f.RuleActive != nil {
		parts = append(parts, "rule_active = ?")
		args = append(args, *f.RuleActive)
	}
	if f.RuleDeleted != nil {
		parts = append(parts, "rule_deleted = ?")
		args = append(args, *f.RuleDeleted)
	}
	if len(f.TagIDs) > 0 {
		tagCSV := int64SliceToCSV(f.TagIDs)
		parts = append(parts,
			"(cast(string_to_array(cast(? as varchar), ',') as int[]) && cast(string_to_array(rule_applied_tags, ',') as int[]))")
		args = append(args, tagCSV)
	}

	return strings.Join(parts, " AND "), args
}

func int64SliceToCSV(ids []int64) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = fmt.Sprintf("%d", id)
	}
	return strings.Join(parts, ",")
}
