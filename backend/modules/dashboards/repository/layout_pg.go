package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"
	"gorm.io/gorm"
)

type pgLayoutRepository struct{ db *gorm.DB }

func NewLayoutRepository(db *gorm.DB) connectors.LayoutRepository {
	return &pgLayoutRepository{db: db}
}

func (r *pgLayoutRepository) Save(ctx context.Context, l *domain.UtmDashboardVisualization) error {
	return r.db.WithContext(ctx).Save(l).Error
}

func (r *pgLayoutRepository) FindByID(ctx context.Context, id uint64) (*domain.UtmDashboardVisualization, error) {
	var l domain.UtmDashboardVisualization
	err := r.db.WithContext(ctx).First(&l, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *pgLayoutRepository) List(ctx context.Context, f dto.LayoutFilter) ([]domain.UtmDashboardVisualization, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.UtmDashboardVisualization{})
	if f.IDDashboard != nil {
		q = q.Where("id_dashboard = ?", *f.IDDashboard)
	}
	if f.IDVisualization != nil {
		q = q.Where("id_visualization = ?", *f.IDVisualization)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domain.UtmDashboardVisualization
	if err := q.Order("dv_order ASC").Offset(f.Offset()).Limit(f.Limit()).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgLayoutRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.UtmDashboardVisualization{}, id).Error
}
