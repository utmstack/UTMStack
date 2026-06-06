package domain

const RelationUserCustomFilter = "USER_CUSTOM_FILTER"

type UtmGroupLogstashPipelineFilters struct {
	ID         int64  `gorm:"primaryKey;autoIncrement"          json:"id"`
	FilterID   int32  `gorm:"column:filter_id;not null"         json:"filterId"`
	PipelineID int32  `gorm:"column:pipeline_id;not null"       json:"pipelineId"`
	Relation   string `gorm:"column:relation;size:50"           json:"relation"`
}

func (UtmGroupLogstashPipelineFilters) TableName() string {
	return "utm_group_logstash_pipeline_filters"
}
