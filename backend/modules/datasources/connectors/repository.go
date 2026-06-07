package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type DatasourceRepository interface {
	FindByID(ctx context.Context, id uint64) (*domain.Datasource, error)
	FindByName(ctx context.Context, name string) (*domain.Datasource, error)
	List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[domain.Datasource], error)
	// UpsertBatch registers-or-updates datasources by name in one statement, refreshing
	// the ping-owned fields and clearing deregistered_at, while preserving user curation
	// (group_id, labels) and discovered_at on existing rows.
	UpsertBatch(ctx context.Context, items []domain.Datasource) error
	UpdateGroup(ctx context.Context, ids []uint64, groupID *uint64) error
	UpdateLabels(ctx context.Context, id uint64, labels string) error
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
