package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type ListResult[T any] struct {
	Items []T
	Total int64
}

type RegexPatternUsecase interface {
	GetByID(ctx context.Context, patternID string) (*dto.RegexPatternResponse, error)
	List(ctx context.Context, f dto.RegexPatternFilters) (*ListResult[dto.RegexPatternResponse], error)
}

type AssetProjectionUsecase interface {
	ProjectAssets(assets []common_models.AssetSensitivity) error
}

type CorrelationRuleUsecase interface {
	Create(ctx context.Context, req dto.CreateCorrelationRuleRequest) error
	ImportRules(ctx context.Context, files []dto.ImportRuleFile) []dto.ImportRuleResult
	Update(ctx context.Context, req dto.UpdateCorrelationRuleRequest) error
	GetByRelPath(ctx context.Context, relPath string) (*dto.CorrelationRuleResponse, error)
	List(ctx context.Context, filters dto.CorrelationRuleFilters) (*ListResult[dto.CorrelationRuleResponse], error)
	Delete(ctx context.Context, relPath string) error
	SetActive(ctx context.Context, relPath string, active bool) (bool, error)
	FindDistinctPropertyValues(ctx context.Context, prop, value string) ([]string, error)
	ExportRules(ctx context.Context, relPaths []string) ([]dto.ExportedRuleFile, error)
}

type FilterUsecase interface {
	Create(ctx context.Context, req dto.CreateFilterRequest) (*dto.FilterResponse, error)
	Update(ctx context.Context, req dto.UpdateFilterRequest) (*dto.FilterResponse, error)
	GetByRelPath(ctx context.Context, relPath string) (*dto.FilterResponse, error)
	List(ctx context.Context, f dto.FilterFilters) ([]dto.FilterResponse, int64, error)
	DataTypes(ctx context.Context) []string
	Delete(ctx context.Context, relPath string) error
	SetActive(ctx context.Context, relPath string, active bool) error
	SetOrder(ctx context.Context, relPath string, order int32) (*dto.FilterResponse, error)
}

type IngestionStatsUsecase interface {
	Totals(ctx context.Context, groupBy, status, from, to string, top int) (*dto.IngestionStatsResponse, error)
	Timeline(ctx context.Context, groupBy, status, interval, from, to string, top int, dataSource string) (*dto.IngestionTimelineResponse, error)
}

type PlaygroundUsecase interface {
	TestFilter(ctx context.Context, req dto.TestFilterRequest) (*dto.PlaygroundResponse, error)
	TestRule(ctx context.Context, req dto.TestRuleRequest) (*dto.PlaygroundResponse, error)
}
