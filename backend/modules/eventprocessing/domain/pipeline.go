package domain

type UtmLogstashPipeline struct {
	ID                  int64   `gorm:"column:id;->;<-:create"            json:"id"`
	PipelineID          *string `gorm:"column:pipeline_id"                json:"pipelineId"`
	PipelineName        *string `gorm:"column:pipeline_name"              json:"pipelineName"`
	PipelineStatus      *string `gorm:"column:pipeline_status"            json:"pipelineStatus"`
	ModuleName          *string `gorm:"column:module_name"                json:"moduleName"`
	SystemOwner         bool    `gorm:"column:system_owner"               json:"systemOwner"`
	PipelineDescription *string `gorm:"column:pipeline_description"       json:"pipelineDescription"`
	PipelineInternal    bool    `gorm:"column:pipeline_internal;default:false" json:"pipelineInternal"`
	EventsOut           *int64  `gorm:"column:events_out"                 json:"eventsOut"`
}

func (UtmLogstashPipeline) TableName() string {
	return "utm_logstash_pipeline"
}
