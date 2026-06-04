package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/opensearch/dto"
)

type GatewayRepository interface {
	PropertyValues(ctx context.Context, index, field string) ([]string, error)
	PropertyValuesWithCount(ctx context.Context, index, field string, filters map[string]any, top int, orderByCount, sortAsc bool) (map[string]int64, error)
	IndexProperties(ctx context.Context, pattern string) ([]dto.IndexPropertyType, error)
	Indices(ctx context.Context, pattern string) ([]dto.IndexInfo, error)
	DeleteIndices(ctx context.Context, indices []string) error
	Search(ctx context.Context, index string, query map[string]any, from, size int, sortField, sortOrder string) ([]map[string]any, int64, error)
	Count(ctx context.Context, index string, query map[string]any) (int64, error)
	CountExists(ctx context.Context, index string, query map[string]any) (bool, error)
	SearchSQL(ctx context.Context, sql string) ([]map[string]any, error)
	CountSQL(ctx context.Context, wrappedSQL string) (int64, error)
	ClusterHealth(ctx context.Context) (*dto.ClusterHealth, error)
}
