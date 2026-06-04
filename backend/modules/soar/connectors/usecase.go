package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

// ListResult is the return type for all paginated list operations.
// The handler writes the items as a bare JSON array and the total as
// X-Total-Count header (matching Java's Spring Page response contract).
type ListResult[T any] struct {
	Items []T
	Total int64
}

type RuleUsecase interface {
	Create(ctx context.Context, req dto.CreateRuleRequest, createdBy string) (*dto.RuleResponse, error)
	Update(ctx context.Context, req dto.UpdateRuleRequest, modifiedBy string) (*dto.RuleResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.RuleResponse, error)
	List(ctx context.Context, f dto.RuleFilters) (*ListResult[dto.RuleResponse], error)
	ResolveFilterValues(ctx context.Context) (*dto.ResolveFilterValuesResponse, error)
}

type TemplateUsecase interface {
	List(ctx context.Context, f dto.TemplateFilters) (*ListResult[dto.TemplateResponse], error)
}

type ExecutionUsecase interface {
	List(ctx context.Context, f dto.ExecutionFilters) (*ListResult[dto.ExecutionResponse], error)
}
