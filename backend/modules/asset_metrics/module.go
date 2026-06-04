package asset_metrics

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/asset_metrics/domain"
	"github.com/utmstack/utmstack/backend/modules/asset_metrics/handler"
	"github.com/utmstack/utmstack/backend/modules/asset_metrics/repository"
	"github.com/utmstack/utmstack/backend/modules/asset_metrics/usecase"
	"gorm.io/gorm"
)

type Module struct {
	handler *handler.AssetMetricHandler
	repo    interface {
		FindAllByAssetName(ctx context.Context, assetName string) ([]domain.UtmAssetMetric, error)
		FindAllByAssetNameIn(ctx context.Context, assetNames []string) ([]domain.UtmAssetMetric, error)
	}
}

func NewModule(db *gorm.DB) *Module {
	repo := repository.NewAssetMetricRepository(db)
	uc := usecase.NewAssetMetricUsecase(repo)
	h := handler.NewAssetMetricHandler(uc)

	return &Module{
		handler: h,
		repo:    repo,
	}
}

func (m *Module) FindAllByAssetName(ctx context.Context, assetName string) ([]domain.UtmAssetMetric, error) {
	return m.repo.FindAllByAssetName(ctx, assetName)
}

func (m *Module) FindAllByAssetNameIn(ctx context.Context, assetNames []string) ([]domain.UtmAssetMetric, error) {
	return m.repo.FindAllByAssetNameIn(ctx, assetNames)
}
