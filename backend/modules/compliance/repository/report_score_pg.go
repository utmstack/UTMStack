package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

type pgReportScoreStore struct{ db *gorm.DB }

func NewReportScoreStore(db *gorm.DB) connectors.ReportScoreStore {
	return &pgReportScoreStore{db: db}
}

func (r *pgReportScoreStore) Upsert(ctx context.Context, p *domain.ReportScore) error {
	tid := tenantFromCtx(ctx)
	if tid == uuid.Nil && tenancy.Enabled() {
		return ErrNoTenant
	}
	p.TenantID = tid
	p.Day = dayOf(p.Day)

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "tenant_id"}, {Name: "framework_key"}, {Name: "day"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"generated_at", "score", "total", "evaluated", "compliant", "body",
		}),
	}).Create(p).Error
}

func (r *pgReportScoreStore) History(ctx context.Context, frameworkKey string, from, to time.Time) ([]domain.ReportScore, error) {
	var rows []domain.ReportScore
	err := scopeTenant(ctx, r.db.WithContext(ctx).Model(&domain.ReportScore{})).
		Select("tenant_id", "framework_key", "day", "generated_at",
			"score", "total", "evaluated", "compliant",
			"(body IS NOT NULL) AS has_body").
		Where("framework_key = ? AND day >= ? AND day <= ?", frameworkKey, dayOf(from), dayOf(to)).
		Order("day ASC").
		Find(&rows).Error
	return rows, err
}

func (r *pgReportScoreStore) Body(ctx context.Context, frameworkKey string, day time.Time) ([]byte, error) {
	var row domain.ReportScore
	err := scopeTenant(ctx, r.db.WithContext(ctx).Model(&domain.ReportScore{})).
		Select("body").
		Where("framework_key = ? AND day = ?", frameworkKey, dayOf(day)).
		Take(&row).Error
	if err != nil {
		return nil, err
	}
	return row.Body, nil
}

func dayOf(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func (r *pgReportScoreStore) PruneBodies(ctx context.Context, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Model(&domain.ReportScore{}).
		Where("day < ? AND body IS NOT NULL", dayOf(before)).
		Update("body", nil)
	return res.RowsAffected, res.Error
}
