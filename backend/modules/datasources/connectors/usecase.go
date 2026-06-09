package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"github.com/utmstack/utmstack/backend/modules/datasources/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

// LicenseCapProvider reports the effective datasource limit for the current
// license, used only to surface the usage banner. Satisfied by the billing
// module's license service.
type LicenseCapProvider interface {
	DatasourceCap() (limit int64, unlimited bool)
}

type DatasourceUsecase interface {
	GetByID(ctx context.Context, id uint64) (*dto.DatasourceDTO, error)
	List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[dto.DatasourceDTO], error)
	Count(ctx context.Context) (int64, error)
	UpdateGroup(ctx context.Context, req dto.UpdateGroupRequest) error
	UpdateLabels(ctx context.Context, req dto.UpdateLabelsRequest) error
	Delete(ctx context.Context, id uint64) error
	Ping(ctx context.Context, req dto.PingRequest) error
	Register(ctx context.Context, req dto.RegisterRequest) error
	Enrichment(ctx context.Context) ([]dto.DatasourceEnrichment, error)
}

type AssetGroupUsecase interface {
	Create(ctx context.Context, g *domain.UtmAssetGroup) (*domain.UtmAssetGroup, error)
	Update(ctx context.Context, g *domain.UtmAssetGroup) (*domain.UtmAssetGroup, error)
	GetByID(ctx context.Context, id uint64) (*domain.UtmAssetGroup, error)
	List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[dto.AssetGroupDTO], error)
	Delete(ctx context.Context, id uint64) error
}
