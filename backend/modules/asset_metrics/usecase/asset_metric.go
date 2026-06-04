package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/asset_metrics/connectors"
	"github.com/utmstack/utmstack/backend/modules/asset_metrics/domain"
)

type assetMetricUsecase struct {
	repo connectors.AssetMetricRepository
}

func NewAssetMetricUsecase(repo connectors.AssetMetricRepository) connectors.AssetMetricUsecase {
	return &assetMetricUsecase{repo: repo}
}

func (u *assetMetricUsecase) Save(ctx context.Context, m domain.UtmAssetMetric) error {
	return u.repo.Save(ctx, m)
}

func (u *assetMetricUsecase) Update(ctx context.Context, m domain.UtmAssetMetric) error {
	return u.repo.Update(ctx, m)
}

func (u *assetMetricUsecase) FindAll(ctx context.Context) ([]domain.UtmAssetMetric, error) {
	return u.repo.FindAll(ctx)
}

func (u *assetMetricUsecase) FindByID(ctx context.Context, id string) (*domain.UtmAssetMetric, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *assetMetricUsecase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}

func (u *assetMetricUsecase) FindAllByAssetName(ctx context.Context, assetName string) ([]domain.UtmAssetMetric, error) {
	return u.repo.FindAllByAssetName(ctx, assetName)
}

func (u *assetMetricUsecase) FindAllByAssetNameIn(ctx context.Context, assetNames []string) ([]domain.UtmAssetMetric, error) {
	return u.repo.FindAllByAssetNameIn(ctx, assetNames)
}
