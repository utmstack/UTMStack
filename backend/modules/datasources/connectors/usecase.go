package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"github.com/utmstack/utmstack/backend/modules/datasources/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type DatasourceUsecase interface {
	GetByID(ctx context.Context, id uint64) (*dto.DatasourceDTO, error)
	List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[dto.DatasourceDTO], error)
	UpdateGroup(ctx context.Context, req dto.UpdateGroupRequest) error
	UpdateLabels(ctx context.Context, req dto.UpdateLabelsRequest) error
	Delete(ctx context.Context, id uint64) error
	Ping(ctx context.Context, req dto.PingRequest) error
	Register(ctx context.Context, req dto.RegisterRequest) error
}

type AssetGroupUsecase interface {
	Create(ctx context.Context, g *domain.UtmAssetGroup) (*domain.UtmAssetGroup, error)
	Update(ctx context.Context, g *domain.UtmAssetGroup) (*domain.UtmAssetGroup, error)
	GetByID(ctx context.Context, id uint64) (*domain.UtmAssetGroup, error)
	List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[dto.AssetGroupDTO], error)
	Delete(ctx context.Context, id uint64) error
}
