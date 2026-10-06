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

// TenantUsage is one tenant's successfully-ingested volume for one UTC day —
// the same "enqueue_success" count the rest of this file already reads, just
// summed per tenant instead of per data source. Feeds the installer's daily
// report to Customer Manager.
type TenantUsage struct {
	TenantID   uuid.UUID
	EventCount int64
	Bytes      int64
}

type StatsReader interface {
	DistinctSources(ctx context.Context, from, to time.Time) ([]StatSource, error)
	TenantUsageByDay(ctx context.Context, day time.Time) ([]TenantUsage, error)
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
