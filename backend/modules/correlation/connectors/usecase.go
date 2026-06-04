package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/correlation/dto"
)

type ListResult[T any] struct {
	Items []T
	Total int64
}

type RegexPatternUsecase interface {
	Create(ctx context.Context, req dto.CreateRegexPatternRequest) (*dto.RegexPatternResponse, error)
	Update(ctx context.Context, req dto.UpdateRegexPatternRequest) (*dto.RegexPatternResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.RegexPatternResponse, error)
	List(ctx context.Context, f dto.RegexPatternFilters) (*ListResult[dto.RegexPatternResponse], error)
	Delete(ctx context.Context, id int64) error
}

type TenantConfigUsecase interface {
	Create(ctx context.Context, req dto.CreateTenantConfigRequest) (*dto.TenantConfigResponse, error)
	Update(ctx context.Context, req dto.UpdateTenantConfigRequest) (*dto.TenantConfigResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.TenantConfigResponse, error)
	List(ctx context.Context, f dto.TenantConfigFilters) (*ListResult[dto.TenantConfigResponse], error)
	Delete(ctx context.Context, id int64) error
}

type DataTypeUsecase interface {
	Create(ctx context.Context, req dto.CreateDataTypeRequest) (*dto.DataTypeResponse, error)
	Update(ctx context.Context, req dto.UpdateDataTypeRequest) (*dto.DataTypeResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.DataTypeResponse, error)
	List(ctx context.Context, f dto.DataTypeFilters) (*ListResult[dto.DataTypeResponse], error)
	Delete(ctx context.Context, id int64) error
	// UpdateIncludeExcludeList updates only the included flag for each item.
	// Items with a nil ID are silently filtered out.
	// Steps 2+3 (asset deletion and synchronization) are no-op stubs until
	// modules #33 and #13 are ported.
	UpdateIncludeExcludeList(ctx context.Context, items []dto.UpdateIncludeExcludeItem) error
}

type CorrelationRuleUsecase interface {
	Create(ctx context.Context, req dto.CreateCorrelationRuleRequest) error

	Update(ctx context.Context, req dto.UpdateCorrelationRuleRequest) error

	GetByID(ctx context.Context, id int64) (*dto.CorrelationRuleResponse, error)

	List(ctx context.Context, filters dto.CorrelationRuleFilters) (*ListResult[dto.CorrelationRuleResponse], error)

	Delete(ctx context.Context, id int64) error

	ActivateDeactivate(ctx context.Context, id int64, active bool) error

	FindDistinctPropertyValues(ctx context.Context, prop, value string) ([]string, error)
}
