package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
	"gorm.io/gorm"
)

type pgAlertTagRepository struct {
	db *gorm.DB
}

func NewAlertTagRepository(db *gorm.DB) connectors.AlertTagRepository {
	return &pgAlertTagRepository{db: db}
}

func (r *pgAlertTagRepository) Create(ctx context.Context, tag domain.AlertTag) (*domain.AlertTag, error) {
	if err := r.db.WithContext(ctx).Create(&tag).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrTagNameTaken
		}
		return nil, err
	}
	return &tag, nil
}

func (r *pgAlertTagRepository) Update(ctx context.Context, tag domain.AlertTag) (*domain.AlertTag, error) {
	if err := r.db.WithContext(ctx).Save(&tag).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrTagNameTaken
		}
		return nil, err
	}
	return &tag, nil
}

func (r *pgAlertTagRepository) List(ctx context.Context, f dto.AlertTagFilters) ([]domain.AlertTag, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Size < 1 || f.Size > 200 {
		f.Size = 20
	}

	q := r.db.WithContext(ctx).Model(&domain.AlertTag{})

	if f.TagName != nil && *f.TagName != "" {
		q = q.Where("tag_name ILIKE ?", "%"+*f.TagName+"%")
	}
	if f.SystemOwner != nil {
		q = q.Where("system_owner = ?", *f.SystemOwner)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []domain.AlertTag
	if err := q.
		Order("id ASC").
		Offset((f.Page - 1) * f.Size).
		Limit(f.Size).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *pgAlertTagRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.AlertTag, error) {
	var row domain.AlertTag
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgAlertTagRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.AlertTag, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []domain.AlertTag
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *pgAlertTagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.AlertTag{}).Error
}

// ---------------------------------------------------------------------------
// Tagging rules
// ---------------------------------------------------------------------------

type pgAlertTagRuleRepository struct {
	db *gorm.DB
}

func NewAlertTagRuleRepository(db *gorm.DB) connectors.AlertTagRuleRepository {
	return &pgAlertTagRuleRepository{db: db}
}

func (r *pgAlertTagRuleRepository) Create(ctx context.Context, rule domain.AlertTagRule) (*domain.AlertTagRule, error) {
	if err := r.db.WithContext(ctx).Create(&rule).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrRuleNameTaken
		}
		return nil, err
	}
	return &rule, nil
}

func (r *pgAlertTagRuleRepository) Update(ctx context.Context, rule domain.AlertTagRule) (*domain.AlertTagRule, error) {
	if err := r.db.WithContext(ctx).Save(&rule).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrRuleNameTaken
		}
		return nil, err
	}
	return &rule, nil
}

func (r *pgAlertTagRuleRepository) List(ctx context.Context, f dto.AlertTagRuleFilters) ([]domain.AlertTagRule, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Size < 1 || f.Size > 200 {
		f.Size = 20
	}

	where, args := buildTagRuleWhere(f)

	table := domain.AlertTagRule{}.TableName()

	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s`, table, where)
	var total int64
	if err := r.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.Size
	dataSQL := fmt.Sprintf(
		`SELECT * FROM %s WHERE %s ORDER BY id ASC LIMIT %d OFFSET %d`,
		table, where, f.Size, offset,
	)
	var rows []domain.AlertTagRule
	if err := r.db.WithContext(ctx).Raw(dataSQL, args...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *pgAlertTagRuleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.AlertTagRule, error) {
	var row domain.AlertTagRule
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgAlertTagRuleRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.AlertTagRule, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []domain.AlertTagRule
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *pgAlertTagRuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Where("id = ?", id).Delete(&domain.AlertTagRule{}).Error
}

func (r *pgAlertTagRuleRepository) FindAllActive(ctx context.Context) ([]domain.AlertTagRule, error) {
	var rows []domain.AlertTagRule
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
		tagCSV := idSliceToCSV(f.TagIDs)
		// Both sides stay text[]: the column holds a comma-separated list of
		// UUIDs, and casting either of them to a narrower type is what breaks
		// the moment an id stops being a number.
		parts = append(parts,
			"(string_to_array(cast(? as varchar), ',') && string_to_array(rule_applied_tags, ','))")
		args = append(args, tagCSV)
	}

	return strings.Join(parts, " AND "), args
}

func idSliceToCSV(ids []uuid.UUID) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = id.String()
	}
	return strings.Join(parts, ",")
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}
