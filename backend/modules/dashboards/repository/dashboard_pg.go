package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"
	"gorm.io/gorm"
)

type pgDashboardRepository struct{ db *gorm.DB }

func NewDashboardRepository(db *gorm.DB) connectors.DashboardRepository {
	return &pgDashboardRepository{db: db}
}

func (r *pgDashboardRepository) Save(ctx context.Context, d *domain.Dashboard) error {
	return r.db.WithContext(ctx).Save(d).Error
}

func (r *pgDashboardRepository) FindByID(ctx context.Context, id uint64) (*domain.Dashboard, error) {
	var d domain.Dashboard
	err := r.db.WithContext(ctx).First(&d, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *pgDashboardRepository) List(ctx context.Context, f dto.DashboardFilter) ([]domain.Dashboard, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.Dashboard{})
	if f.Name != "" {
		q = q.Where("name ILIKE ?", "%"+f.Name+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domain.Dashboard
	if err := q.Order("name ASC").Offset(f.Offset()).Limit(f.Limit()).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// Delete removes the dashboard and its visualizations together — a
// visualization can't outlive (or be reused outside of) its dashboard.
func (r *pgDashboardRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("dashboard_id = ?", id).Delete(&domain.Visualization{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.Dashboard{}, id).Error
	})
}
