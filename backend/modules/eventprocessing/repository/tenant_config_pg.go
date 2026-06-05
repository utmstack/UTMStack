package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/correlation/connectors"
	"github.com/utmstack/utmstack/backend/modules/correlation/domain"
	correrrors "github.com/utmstack/utmstack/backend/modules/correlation/errors"
	"gorm.io/gorm"
)

type pgTenantConfigRepository struct {
	db *gorm.DB
}

func NewTenantConfigRepository(db *gorm.DB) connectors.TenantConfigRepository {
	return &pgTenantConfigRepository{db: db}
}

func (r *pgTenantConfigRepository) Create(ctx context.Context, t *domain.UtmTenantConfig) (*domain.UtmTenantConfig, error) {
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return nil, err
	}
	return t, nil
}

func (r *pgTenantConfigRepository) Update(ctx context.Context, t *domain.UtmTenantConfig) (*domain.UtmTenantConfig, error) {
	if err := r.db.WithContext(ctx).Save(t).Error; err != nil {
		return nil, err
	}
	return t, nil
}

func (r *pgTenantConfigRepository) GetByID(ctx context.Context, id int64) (*domain.UtmTenantConfig, error) {
	var t domain.UtmTenantConfig
	err := r.db.WithContext(ctx).First(&t, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *pgTenantConfigRepository) List(ctx context.Context, f connectors.TenantConfigFilters) ([]domain.UtmTenantConfig, int64, error) {
	page, size := normPage(f.Page, f.Size)

	q := r.db.WithContext(ctx).Model(&domain.UtmTenantConfig{})

	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("asset_name ILIKE ?", like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []domain.UtmTenantConfig
	if err := q.Order("id ASC").
		Offset(page * size).
		Limit(size).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgTenantConfigRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&domain.UtmTenantConfig{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return correrrors.ErrTenantConfigNotFound
	}
	return nil
}
