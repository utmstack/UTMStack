package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
)

type ListResult[T any] struct {
	Items []T
	Total int64
}

type RegexPatternUsecase interface {
	Create(ctx context.Context, req dto.CreateRegexPatternRequest) (*dto.RegexPatternResponse, error)
	Update(ctx context.Context, req dto.UpdateRegexPatternRequest) (*dto.RegexPatternResponse, error)
	GetByID(ctx context.Context, patternID string) (*dto.RegexPatternResponse, error)
	List(ctx context.Context, f dto.RegexPatternFilters) (*ListResult[dto.RegexPatternResponse], error)
	Delete(ctx context.Context, patternID string) error
}

type TenantConfigUsecase interface {
	Create(ctx context.Context, req dto.CreateTenantConfigRequest) (*dto.TenantConfigResponse, error)
	Update(ctx context.Context, req dto.UpdateTenantConfigRequest) (*dto.TenantConfigResponse, error)
	GetByID(ctx context.Context, assetName string) (*dto.TenantConfigResponse, error)
	List(ctx context.Context, f dto.TenantConfigFilters) (*ListResult[dto.TenantConfigResponse], error)
	Delete(ctx context.Context, assetName string) error
}

type CorrelationRuleUsecase interface {
	Create(ctx context.Context, req dto.CreateCorrelationRuleRequest) error
	Update(ctx context.Context, req dto.UpdateCorrelationRuleRequest) error
	GetByRelPath(ctx context.Context, relPath string) (*dto.CorrelationRuleResponse, error)
	List(ctx context.Context, filters dto.CorrelationRuleFilters) (*ListResult[dto.CorrelationRuleResponse], error)
	Delete(ctx context.Context, relPath string) error
	SetActive(ctx context.Context, relPath string, active bool) error
	FindDistinctPropertyValues(ctx context.Context, prop, value string) ([]string, error)
}

type FilterUsecase interface {
	Create(ctx context.Context, req dto.CreateFilterRequest) (*dto.FilterResponse, error)
	Update(ctx context.Context, req dto.UpdateFilterRequest) (*dto.FilterResponse, error)
	GetByRelPath(ctx context.Context, relPath string) (*dto.FilterResponse, error)
	List(ctx context.Context, f dto.FilterFilters) ([]dto.FilterResponse, int64, error)
	Delete(ctx context.Context, relPath string) error
	SetActive(ctx context.Context, relPath string, active bool) error
}

type IngestionStatsUsecase interface {
	Totals(ctx context.Context, groupBy, status, from, to string, top int) (*dto.IngestionStatsResponse, error)
	Timeline(ctx context.Context, groupBy, status, interval, from, to string, top int) (*dto.IngestionTimelineResponse, error)
}
