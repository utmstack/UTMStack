package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/opensearch/domain"
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
	DiskUsage(ctx context.Context) (float64, error)
	OldestIndices(ctx context.Context, pattern string) ([]string, error)
}

type IndexPatternRepository interface {
	Create(ctx context.Context, p *domain.UtmIndexPattern) error
	Update(ctx context.Context, p *domain.UtmIndexPattern) error
	GetByID(ctx context.Context, id int64) (*domain.UtmIndexPattern, error)
	List(ctx context.Context, f dto.IndexPatternFilters) ([]domain.UtmIndexPattern, int64, error)
	Count(ctx context.Context, f dto.IndexPatternFilters) (int64, error)
	Delete(ctx context.Context, id int64) error
	FindAll(ctx context.Context) ([]domain.UtmIndexPattern, error)
}

type ISMRepository interface {
	GetPolicy(ctx context.Context) (*domain.IndexPolicy, error)
	CreatePolicy(ctx context.Context, policy domain.IndexPolicy) error
	UpdatePolicy(ctx context.Context, policy domain.IndexPolicy, seqNo, primaryTerm int64) error
	AddPolicyToIndex(ctx context.Context, indexPattern string) error
	ChangePolicyForIndex(ctx context.Context, pattern string, req domain.UpdateManagedIndexPolicyConfiguration) (*domain.UpdateManagedIndexPolicyResponse, error)
	GetSnapshotRepo(ctx context.Context) (bool, error)
	RegisterSnapshotRepo(ctx context.Context) error
	IsIndexRemovable(ctx context.Context, indexName string) (bool, error)
	GetIndexProperties(ctx context.Context, pattern string) ([]dto.IndexField, error)
}
