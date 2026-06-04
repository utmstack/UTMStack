package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/asset_metrics/domain"
)

type AssetMetricRepository interface {
	Save(ctx context.Context, m domain.UtmAssetMetric) error
	Update(ctx context.Context, m domain.UtmAssetMetric) error
	FindAll(ctx context.Context) ([]domain.UtmAssetMetric, error)
	FindByID(ctx context.Context, id string) (*domain.UtmAssetMetric, error)
	Delete(ctx context.Context, id string) error
	FindAllByAssetName(ctx context.Context, assetName string) ([]domain.UtmAssetMetric, error)
	FindAllByAssetNameIn(ctx context.Context, assetNames []string) ([]domain.UtmAssetMetric, error)
}

type AssetMetricUsecase interface {
	Save(ctx context.Context, m domain.UtmAssetMetric) error
	Update(ctx context.Context, m domain.UtmAssetMetric) error
	FindAll(ctx context.Context) ([]domain.UtmAssetMetric, error)
	FindByID(ctx context.Context, id string) (*domain.UtmAssetMetric, error)
	Delete(ctx context.Context, id string) error
	FindAllByAssetName(ctx context.Context, assetName string) ([]domain.UtmAssetMetric, error)
	FindAllByAssetNameIn(ctx context.Context, assetNames []string) ([]domain.UtmAssetMetric, error)
}
