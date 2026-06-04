package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/datainput/dto"
)

type DataInputStatusUsecase interface {
	Create(ctx context.Context, req dto.CreateDataInputStatusRequest) (*dto.DataInputStatusResponse, error)
	Update(ctx context.Context, req dto.UpdateDataInputStatusRequest) (*dto.DataInputStatusResponse, error)
	GetByID(ctx context.Context, id string) (*dto.DataInputStatusResponse, error)
	List(ctx context.Context, filters dto.DataInputStatusListFilters) ([]dto.DataInputStatusResponse, int64, error)
	Count(ctx context.Context, filters dto.DataInputStatusCountFilters) (int64, error)
	Delete(ctx context.Context, id string) error
}
