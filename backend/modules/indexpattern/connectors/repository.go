package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/indexpattern/domain"
	"github.com/utmstack/utmstack/backend/modules/indexpattern/dto"
)

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
