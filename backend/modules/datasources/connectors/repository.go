package connectors

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type StatSource struct {
	// TenantID is whose source this is. One reconciler serves every tenant, so
	// a batch spans them and the tenant cannot be inferred from the caller.
	TenantID   string
	DataSource string
	DataType   string
	LastSeen   time.Time
}

type StatsReader interface {
	DistinctSources(ctx context.Context, from, to time.Time) ([]StatSource, error)
}

type DatasourceRepository interface {
	FindByID(ctx context.Context, id uint64) (*domain.Datasource, error)
	List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[domain.Datasource], error)
	Count(ctx context.Context) (int64, error)
	UpsertBatch(ctx context.Context, items []domain.Datasource) error
	RegisterBatch(ctx context.Context, items []domain.Datasource) error
	UpsertLivenessBatch(ctx context.Context, items []domain.Datasource) error
	EnrichmentRows(ctx context.Context) ([]domain.Datasource, error)
	UpdateGroup(ctx context.Context, ids []uint64, groupID *uint64) error
	UpdateLabels(ctx context.Context, id uint64, labels string) error
	UpdateSensitivity(ctx context.Context, id uint64, conf, integ, avail int) error
	// ClearGroup nulls group_id on every datasource pointing at groupID so an
	// asset-group delete does not leave dangling references. The tenancy
	// callback scopes the write to the caller's tenant.
	ClearGroup(ctx context.Context, groupID uint64) error
	ListSensitive(ctx context.Context) ([]domain.Datasource, error)
	Delete(ctx context.Context, id uint64) error
}

type AssetGroupRepository interface {
	Save(ctx context.Context, g *domain.UtmAssetGroup) error
	FindByID(ctx context.Context, id uint64) (*domain.UtmAssetGroup, error)
	List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[AssetGroupRow], error)
	Delete(ctx context.Context, id uint64) error
}

type AssetGroupRow struct {
	Group            domain.UtmAssetGroup
	DatasourcesCount int64
}
