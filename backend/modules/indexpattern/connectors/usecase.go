package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/indexpattern/domain"
	"github.com/utmstack/utmstack/backend/modules/indexpattern/dto"
)

type IndexPatternUsecase interface {
	Create(ctx context.Context, req dto.CreateIndexPatternRequest) (*dto.IndexPatternResponse, error)
	Update(ctx context.Context, req dto.UpdateIndexPatternRequest) (*dto.IndexPatternResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.IndexPatternResponse, error)
	List(ctx context.Context, f dto.IndexPatternFilters) ([]dto.IndexPatternResponse, int64, error)
	Fields(ctx context.Context, f dto.IndexPatternFilters) ([]dto.IndexPatternFieldsResponse, int64, error)
	Delete(ctx context.Context, id int64) error
}

type PolicyUsecase interface {
	GetPolicy(ctx context.Context) (*domain.PolicySettings, error)
	UpdatePolicy(ctx context.Context, settings domain.PolicySettings) (*domain.UpdateManagedIndexPolicyResponse, error)
	IsIndexRemovable(ctx context.Context, indexName string) bool
}
