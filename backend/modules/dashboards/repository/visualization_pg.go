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

func (r *pgVisualizationRepository) Save(ctx context.Context, v *domain.UtmVisualization) error {
	return r.db.WithContext(ctx).Save(v).Error
}

func (r *pgVisualizationRepository) FindByID(ctx context.Context, id uint64) (*domain.UtmVisualization, error) {
	var v domain.UtmVisualization
	err := r.db.WithContext(ctx).First(&v, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *pgVisualizationRepository) List(ctx context.Context, f dto.VisualizationFilter) ([]domain.UtmVisualization, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.UtmVisualization{})
	if f.Name != "" {
		q = q.Where("name ILIKE ?", "%"+f.Name+"%")
	}
	if f.ChartType != "" {
		q = q.Where("chart_type = ?", f.ChartType)
	}
	if f.IDPattern != nil {
		q = q.Where("id_pattern = ?", *f.IDPattern)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domain.UtmVisualization
	if err := q.Order("name ASC").Offset(f.Offset()).Limit(f.Limit()).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgVisualizationRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.UtmVisualization{}, id).Error
}
