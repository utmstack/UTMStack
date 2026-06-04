package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

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
