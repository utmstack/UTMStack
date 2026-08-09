package connectors

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type StatSource struct {
	TenantID   uuid.UUID
	DataSource string
	DataType   string
	LastSeen   time.Time
}

type StatsReader interface {
	DistinctSources(ctx context.Context, from, to time.Time) ([]StatSource, error)
}

type DatasourceRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Datasource, error)
	List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[domain.Datasource], error)
	Count(ctx context.Context) (int64, error)
	UpsertBatch(ctx context.Context, items []domain.Datasource) error
	RegisterBatch(ctx context.Context, items []domain.Datasource) error
	UpsertLivenessBatch(ctx context.Context, items []domain.Datasource) error
	UpdateLabels(ctx context.Context, id uuid.UUID, labels string) error
	UpdateSensitivity(ctx context.Context, id uuid.UUID, conf, integ, avail int) error
	ListSensitive(ctx context.Context) ([]domain.Datasource, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
