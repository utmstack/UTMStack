package usecase

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"github.com/utmstack/utmstack/backend/modules/datasources/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type assetGroupUsecase struct {
	repo        connectors.AssetGroupRepository
	datasources connectors.DatasourceRepository // for clearing dangling group_id on delete
}

func NewAssetGroupUsecase(repo connectors.AssetGroupRepository, datasources connectors.DatasourceRepository) connectors.AssetGroupUsecase {
	return &assetGroupUsecase{repo: repo, datasources: datasources}
}

func (u *assetGroupUsecase) Create(ctx context.Context, g *domain.UtmAssetGroup) (*domain.UtmAssetGroup, error) {
	if g.CreatedDate.IsZero() {
		g.CreatedDate = time.Now().UTC()
	}
	if err := u.repo.Save(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (u *assetGroupUsecase) Update(ctx context.Context, g *domain.UtmAssetGroup) (*domain.UtmAssetGroup, error) {
	if err := u.repo.Save(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (u *assetGroupUsecase) GetByID(ctx context.Context, id uint64) (*domain.UtmAssetGroup, error) {
	g, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, domain.ErrNotFound
	}
	return g, nil
}

func (u *assetGroupUsecase) List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[dto.AssetGroupDTO], error) {
	res, err := u.repo.List(ctx, req)
	if err != nil {
		return common_models.ListResponse[dto.AssetGroupDTO]{}, err
	}
	return common_models.MapList(res, func(row connectors.AssetGroupRow) dto.AssetGroupDTO {
		return dto.ToAssetGroupDTO(&row.Group, row.DatasourcesCount)
	}), nil
}

func (u *assetGroupUsecase) Delete(ctx context.Context, id uint64) error {
	// Clear the FK first: there is no ON DELETE SET NULL, so a group with
	// datasources attached would leave them pointing nowhere. The clear is
	// tenant-scoped by the callback, so another tenant's rows are untouched.
	if u.datasources != nil {
		if err := u.datasources.ClearGroup(ctx, id); err != nil {
			return err
		}
	}
	return u.repo.Delete(ctx, id)
}
