package dto

import "github.com/utmstack/utmstack/backend/modules/logstash/domain"

type UtmLogstashPipelineDTO struct {
	ID                  int64   `json:"id"`
	PipelineID          *string `json:"pipelineId"`
	PipelineName        *string `json:"pipelineName"`
	PipelineStatus      *string `json:"pipelineStatus"`
	ModuleName          *string `json:"moduleName"`
	SystemOwner         bool    `json:"systemOwner"`
	PipelineDescription *string `json:"pipelineDescription"`
	PipelineInternal    bool    `json:"pipelineInternal"`
}

func PipelineDTOFromDomain(p domain.UtmLogstashPipeline) UtmLogstashPipelineDTO {
	return UtmLogstashPipelineDTO{
		ID:                  p.ID,
		PipelineID:          p.PipelineID,
		PipelineName:        p.PipelineName,
		PipelineStatus:      p.PipelineStatus,
		ModuleName:          p.ModuleName,
		SystemOwner:         p.SystemOwner,
		PipelineDescription: p.PipelineDescription,
		PipelineInternal:    p.PipelineInternal,
	}
}

type UtmLogstashPipelineVM struct {
	Pipeline UtmLogstashPipelineDTO                   `json:"pipeline"`
	Filters  []domain.UtmGroupLogstashPipelineFilters `json:"filters"`
}

type ApiStatsResponse struct {
	General   *ApiEngineResponse `json:"general"`
	Pipelines []PipelineStats    `json:"pipelines"`
}

type ApiEngineResponse struct {
	Version string `json:"version"`
	Status  string `json:"status"`
}

type PipelineEvents struct {
	Out *int64 `json:"out"`
}

type PipelineStats struct {
	ID                  int64          `json:"id"`
	PipelineID          *string        `json:"pipelineId"`
	PipelineName        string         `json:"pipelineName"`
	PipelineStatus      *string        `json:"pipelineStatus"`
	ModuleName          *string        `json:"moduleName"`
	SystemOwner         bool           `json:"systemOwner"`
	PipelineDescription *string        `json:"pipelineDescription"`
	Events              PipelineEvents `json:"events"`
	Errors              int64          `json:"errors"`
}

type StatisticDocument struct {
	DataSource string `json:"dataSource"`
	DataType   string `json:"dataType"`
	Count      int64  `json:"count"`
	Type       string `json:"type"`
	Cause      string `json:"cause"`
	Timestamp  string `json:"@timestamp"`
}

type PipelineErrors struct {
	ValidationErrors []Validation `json:"validationErrors"`
}

type Validation struct {
	Entity string `json:"entity"`
	Field  string `json:"field"`
	Msg    string `json:"msg"`
}

const (
	PipelineValidationModeInsert = "INSERT"
	PipelineValidationModeUpdate = "UPDATE"
)
