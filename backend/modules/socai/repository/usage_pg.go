package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/utmstack/utmstack/backend/modules/socai/domain"
)

type UsageRepo struct{ db *gorm.DB }

func NewUsageRepo(db *gorm.DB) *UsageRepo { return &UsageRepo{db: db} }

func (r *UsageRepo) Consume(ctx context.Context, tenantID string) (int64, error) {
	day := time.Now().UTC().Truncate(24 * time.Hour)

	var count int64
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO socai_ai_usage (tenant_id, day, count)
		VALUES (?, ?, 1)
		ON CONFLICT (tenant_id, day)
		DO UPDATE SET count = socai_ai_usage.count + 1
		RETURNING count`, tenantID, day).Scan(&count).Error

	return count, err
}

// UsedToday reports the running total without spending anything.
func (r *UsageRepo) UsedToday(ctx context.Context, tenantID string) (int64, error) {
	day := time.Now().UTC().Truncate(24 * time.Hour)

	var row domain.AIUsage
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND day = ?", tenantID, day).
		Take(&row).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	return row.Count, err
}
