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
	Update(ctx context.Context, req dto.UpdateCorrelationRuleRequest) error
	Delete(ctx context.Context, relPath string) error
	GetByRelPath(ctx context.Context, relPath string) (*dto.CorrelationRuleResponse, error)
	List(ctx context.Context, filters dto.CorrelationRuleFilters) (*ListResult[dto.CorrelationRuleResponse], error)
	SetActive(ctx context.Context, relPath string, active bool) (bool, error)
	FindDistinctPropertyValues(ctx context.Context, prop, value string) ([]string, error)
	ImportRules(ctx context.Context, files []dto.ImportRuleFile) []dto.ImportRuleResult
	ExportRules(ctx context.Context, relPaths []string) ([]dto.ExportedRuleFile, error)
}

type PipelineUsecase interface {
	Create(ctx context.Context, req dto.CreatePipelineRequest) (*dto.PipelineResponse, error)
	Update(ctx context.Context, req dto.UpdatePipelineRequest) (*dto.PipelineResponse, error)
	Delete(ctx context.Context, relPath string) error
	GetByRelPath(ctx context.Context, relPath string) (*dto.PipelineResponse, error)
	List(ctx context.Context, f dto.PipelineFilters) (*ListResult[dto.PipelineResponse], error)
	SetActive(ctx context.Context, relPath string, active bool) error
	SetOrder(ctx context.Context, order []string) error
	DataTypes(ctx context.Context) []string
}

type IngestionStatsUsecase interface {
	Totals(ctx context.Context, groupBy, status, from, to string, top int) (*dto.IngestionStatsResponse, error)
	Timeline(ctx context.Context, groupBy, status, interval, from, to string, top int, dataSource string) (*dto.IngestionTimelineResponse, error)
}

type PlaygroundUsecase interface {
	TestPipeline(ctx context.Context, req dto.TestPipelineRequest) (*dto.PlaygroundResponse, error)
	TestRule(ctx context.Context, req dto.TestRuleRequest) (*dto.PlaygroundResponse, error)
}
