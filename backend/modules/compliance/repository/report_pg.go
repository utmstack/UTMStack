package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

type pgReportStore struct{ db *gorm.DB }

func NewReportStore(db *gorm.DB) connectors.ReportStore { return &pgReportStore{db: db} }

func (r *pgReportStore) Get(ctx context.Context, frameworkKey string) (*domain.Report, error) {
	var row domain.Report
	err := scopeTenant(ctx, r.db.WithContext(ctx)).
		First(&row, "framework_key = ?", frameworkKey).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrReportNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgReportStore) Save(ctx context.Context, rep *domain.Report) error {
	tid := tenantFromCtx(ctx)
	if tid == uuid.Nil && tenancy.Enabled() {
		return ErrNoTenant
	}
	rep.TenantID = tid

	expected := rep.Version
	rep.Version = expected + 1

	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "tenant_id"}, {Name: "framework_key"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"framework_name", "framework_source", "generated_at",
			"window_from", "window_to", "score", "version", "body",
		}),
		Where: clause.Where{Exprs: []clause.Expression{
			clause.Eq{
				Column: clause.Column{Table: "compliance_report", Name: "version"},
				Value:  expected,
			},
		}},
	}).Create(rep)

	if res.Error != nil {
		rep.Version = expected
		return res.Error
	}
	if res.RowsAffected == 0 {
		rep.Version = expected
		return domain.ErrReportConflict
	}
	return nil
}

func (r *pgReportStore) List(ctx context.Context) ([]domain.Report, error) {
	var rows []domain.Report
	err := scopeTenant(ctx, r.db.WithContext(ctx).Model(&domain.Report{})).
		Select("id", "tenant_id", "framework_key", "framework_name", "generated_at", "score").
		Order("framework_name ASC").
		Find(&rows).Error
	return rows, err
}

func (r *pgReportStore) Delete(ctx context.Context, frameworkKey string) error {
	res := scopeTenant(ctx, r.db.WithContext(ctx)).
		Delete(&domain.Report{}, "framework_key = ?", frameworkKey)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrReportNotFound
	}
	return nil
}
