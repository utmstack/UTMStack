package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/asset_metrics/connectors"
	"github.com/utmstack/utmstack/backend/modules/asset_metrics/domain"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("asset metric not found")

type pgAssetMetricRepository struct {
	db *gorm.DB
}

func NewAssetMetricRepository(db *gorm.DB) connectors.AssetMetricRepository {
	return &pgAssetMetricRepository{db: db}
}

func (r *pgAssetMetricRepository) Save(ctx context.Context, m domain.UtmAssetMetric) error {
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *pgAssetMetricRepository) Update(ctx context.Context, m domain.UtmAssetMetric) error {
	return r.db.WithContext(ctx).Save(&m).Error
}

func (r *pgAssetMetricRepository) FindAll(ctx context.Context) ([]domain.UtmAssetMetric, error) {
	var items []domain.UtmAssetMetric
	if err := r.db.WithContext(ctx).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *pgAssetMetricRepository) FindByID(ctx context.Context, id string) (*domain.UtmAssetMetric, error) {
	var m domain.UtmAssetMetric
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *pgAssetMetricRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.UtmAssetMetric{}, "id = ?", id).Error
}

func (r *pgAssetMetricRepository) FindAllByAssetName(ctx context.Context, assetName string) ([]domain.UtmAssetMetric, error) {
	var items []domain.UtmAssetMetric
	if err := r.db.WithContext(ctx).Where("asset_name = ?", assetName).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *pgAssetMetricRepository) FindAllByAssetNameIn(ctx context.Context, assetNames []string) ([]domain.UtmAssetMetric, error) {
	var items []domain.UtmAssetMetric
	if len(assetNames) == 0 {
		return items, nil
	}
	if err := r.db.WithContext(ctx).Where("asset_name IN ?", assetNames).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
