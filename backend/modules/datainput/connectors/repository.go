package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/datainput/domain"
	"github.com/utmstack/utmstack/backend/modules/datainput/dto"
)

type DataInputStatusRepository interface {
	Create(ctx context.Context, status *domain.UtmDataInputStatus) error
	Update(ctx context.Context, status *domain.UtmDataInputStatus) error
	GetByID(ctx context.Context, id string) (*domain.UtmDataInputStatus, error)
	FindImportantDatasource(ctx context.Context, page, size int) ([]domain.UtmDataInputStatus, int64, error)
	Count(ctx context.Context, filters dto.DataInputStatusCountFilters) (int64, error)
	Delete(ctx context.Context, id string) error
	FindDataSourcesToConfigure(ctx context.Context, excludeType string) ([]string, error)
	GetCheckpoint(ctx context.Context) (*domain.UtmDataInputStatusCheckpoint, error)
	SaveCheckpoint(ctx context.Context, checkpoint *domain.UtmDataInputStatusCheckpoint) error
}
