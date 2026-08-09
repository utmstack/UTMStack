package connectors

import (
	"context"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/datasources/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type AssetProjector interface {
	ProjectAssets(assets []common_models.AssetSensitivity) error
}

type DatasourceUsecase interface {
	GetByID(ctx context.Context, id uuid.UUID) (*dto.DatasourceDTO, error)
	List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[dto.DatasourceDTO], error)
	Count(ctx context.Context) (int64, error)
	UpdateLabels(ctx context.Context, req dto.UpdateLabelsRequest) error
	UpdateSensitivity(ctx context.Context, req dto.UpdateSensitivityRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	Ping(ctx context.Context, req dto.PingRequest) error
	Register(ctx context.Context, req dto.RegisterRequest) error
	ProjectAssets(ctx context.Context) error
}
