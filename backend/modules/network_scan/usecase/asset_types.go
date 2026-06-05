package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
)

type assetTypesUsecase struct {
	repo connectors.AssetTypesRepository
}

func NewAssetTypesUsecase(repo connectors.AssetTypesRepository) connectors.AssetTypesUsecase {
	return &assetTypesUsecase{repo: repo}
}

func (u *assetTypesUsecase) List(ctx context.Context, p domain.Pagination) ([]domain.UtmAssetTypes, int64, error) {
	return u.repo.List(ctx, p)
}
