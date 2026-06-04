package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/logstash/domain"
	"github.com/utmstack/utmstack/backend/modules/logstash/dto"
)

type FilterGroupRepository interface {
	Create(ctx context.Context, group *domain.UtmLogstashFilterGroup) error
	Update(ctx context.Context, group *domain.UtmLogstashFilterGroup) error
	GetByID(ctx context.Context, id int64) (*domain.UtmLogstashFilterGroup, error)
	List(ctx context.Context, filters dto.FilterGroupListFilters) ([]domain.UtmLogstashFilterGroup, int64, error)
	Count(ctx context.Context, filters dto.FilterGroupCountFilters) (int64, error)
	Delete(ctx context.Context, id int64) error
}

type FilterRepository interface {
	Create(ctx context.Context, filter *domain.UtmLogstashFilter) error
	Update(ctx context.Context, filter *domain.UtmLogstashFilter) error
	GetByID(ctx context.Context, id int64) (*domain.UtmLogstashFilter, error)
	List(ctx context.Context, filters dto.FilterFilters) ([]domain.UtmLogstashFilter, int64, error)
	FiltersByPipelineID(ctx context.Context, pipelineID int64) ([]domain.UtmLogstashFilter, error)
	Delete(ctx context.Context, id int64) error
	FindAllBySystemOwner(ctx context.Context) ([]domain.UtmLogstashFilter, error)
	FindByLogstashFilterAndSystemOwner(ctx context.Context, content string) (*domain.UtmLogstashFilter, error)
	FindByFilterName(ctx context.Context, name string) (*domain.UtmLogstashFilter, error)
	FindByModuleName(ctx context.Context, moduleName string) ([]domain.UtmLogstashFilter, error)
	// FindDataTypeByID resolves a data_type string from utm_data_types by id.
	// Used by the scheduler to resolve filter.DataTypeID → dataType string.
	FindDataTypeByID(ctx context.Context, dataTypeID int64) (string, error)
}

type PipelineFilterRepository interface {
	GetFilters(ctx context.Context, pipelineID int32) ([]domain.UtmGroupLogstashPipelineFilters, error)
	Save(ctx context.Context, rel *domain.UtmGroupLogstashPipelineFilters) error
	DeleteByFilterID(ctx context.Context, filterID int32) error
	DeleteRelationsByPipelineID(ctx context.Context, pipelineID int32) error
}

type PipelineRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.UtmLogstashPipeline, error)
	AllActivePipelinesByServer(ctx context.Context) ([]domain.UtmLogstashPipeline, error)
	SaveAll(ctx context.Context, pipelines []domain.UtmLogstashPipeline) error
	GetNextID(ctx context.Context) (int64, error)
	DeletePipeline(ctx context.Context, id int64) error
}

type StatisticsRepository interface {
	GetLatestStatistic(ctx context.Context, dataType string) (*dto.StatisticDocument, error)
}
