package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"
	"gorm.io/gorm"
)

type pgVisualizationRepository struct{ db *gorm.DB }

func NewVisualizationRepository(db *gorm.DB) connectors.VisualizationRepository {
	return &pgVisualizationRepository{db: db}
}

func (r *pgVisualizationRepository) Save(ctx context.Context, v *domain.Visualization) error {
	return r.db.WithContext(ctx).Save(v).Error
}

func (r *pgVisualizationRepository) FindByID(ctx context.Context, id uint64) (*domain.Visualization, error) {
	var v domain.Visualization
	err := r.db.WithContext(ctx).First(&v, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *pgVisualizationRepository) List(ctx context.Context, f dto.VisualizationFilter) ([]domain.Visualization, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.Visualization{})
	if f.DashboardID != nil {
		q = q.Where("dashboard_id = ?", *f.DashboardID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domain.Visualization
	if err := q.Order("id ASC").Offset(f.Offset()).Limit(f.Limit()).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgVisualizationRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Visualization{}, id).Error
}
