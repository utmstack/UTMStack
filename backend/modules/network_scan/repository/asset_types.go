package repository

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
	"gorm.io/gorm"
)

type pgAssetTypesRepository struct {
	db *gorm.DB
}

func NewAssetTypesRepository(db *gorm.DB) connectors.AssetTypesRepository {
	return &pgAssetTypesRepository{db: db}
}

func (r *pgAssetTypesRepository) List(ctx context.Context, p domain.Pagination) ([]domain.UtmAssetTypes, int64, error) {
	page, size := normalizePage(p)
	q := r.db.WithContext(ctx).Model(&domain.UtmAssetTypes{})

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []domain.UtmAssetTypes
	if err := q.Order("type_name ASC").Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
