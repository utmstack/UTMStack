package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/logstash/dto"
)

type FilterGroupUsecase interface {
	Create(ctx context.Context, req dto.CreateFilterGroupRequest) (*dto.FilterGroupResponse, error)
	Update(ctx context.Context, req dto.UpdateFilterGroupRequest) (*dto.FilterGroupResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.FilterGroupResponse, error)
	List(ctx context.Context, filters dto.FilterGroupListFilters) ([]dto.FilterGroupResponse, int64, error)
	Count(ctx context.Context, filters dto.FilterGroupCountFilters) (int64, error)
	Delete(ctx context.Context, id int64) error
}

type FilterUsecase interface {
	Create(ctx context.Context, req dto.CreateFilterRequest, pipelineID *int64) (*dto.FilterResponse, error)
	Update(ctx context.Context, req dto.UpdateFilterRequest) (*dto.FilterResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.FilterResponse, error)
	List(ctx context.Context, filters dto.FilterFilters) ([]dto.FilterResponse, int64, error)
	FiltersByPipelineID(ctx context.Context, pipelineID int64) ([]dto.FilterResponse, error)
	Delete(ctx context.Context, id int64) error
}

type PipelineUsecase interface {
	List(ctx context.Context, page, size int) ([]dto.UtmLogstashPipelineDTO, int64, error)
	GetStats(ctx context.Context) (*dto.ApiStatsResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.UtmLogstashPipelineVM, error)
	Validate(ctx context.Context, vm dto.UtmLogstashPipelineVM, mode string) (*dto.PipelineErrors, error)
	Delete(ctx context.Context, id int64) error
}
