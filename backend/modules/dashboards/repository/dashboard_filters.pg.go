package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"
	"gorm.io/gorm"

	"github.com/google/uuid"
)

type pgDashboardFilterRepository struct{ db *gorm.DB }

func NewDashboardFilterRepository(db *gorm.DB) connectors.DashboardFilterRepository {
	return &pgDashboardFilterRepository{db: db}
}

func (r *pgDashboardFilterRepository) Save(ctx context.Context, v *domain.DashboardFilter) error {
	return r.db.WithContext(ctx).Save(v).Error
}

func (r *pgDashboardFilterRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.DashboardFilter, error) {
	var v domain.DashboardFilter
	err := r.db.WithContext(ctx).First(&v, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *pgDashboardFilterRepository) List(ctx context.Context, f dto.DashboardFilterFilter) ([]domain.DashboardFilter, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.DashboardFilter{})
	if f.DashboardID != nil {
		q = q.Where("dashboard_id = ?", *f.DashboardID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domain.DashboardFilter
	if err := q.Order("id ASC").Offset(f.Offset()).Limit(f.Limit()).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgDashboardFilterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.DashboardFilter{}, id).Error
}
