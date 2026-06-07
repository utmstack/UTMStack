package domain

import "time"

type UtmVisualization struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"column:name;size:100;not null;uniqueIndex" json:"name"`
	Description  string    `gorm:"column:description;size:255" json:"description"`
	EventType    string    `gorm:"column:event_type;size:50;not null" json:"eventType"`
	CreatedDate  time.Time `gorm:"column:created_date" json:"createdDate"`
	ModifiedDate time.Time `gorm:"column:modified_date" json:"modifiedDate"`
	UserCreated  string    `gorm:"column:user_created;size:50" json:"userCreated"`
	UserModified string    `gorm:"column:user_modified;size:50" json:"userModified"`
	ChartConfig  string    `gorm:"column:chart_config" json:"chartConfig"`
	ChartAction  string    `gorm:"column:chart_action" json:"chartAction"`
	SystemOwner  bool      `gorm:"column:system_owner" json:"systemOwner"`
	IDPattern    *uint64   `gorm:"column:id_pattern" json:"idPattern,omitempty"`
	ChartType    string    `gorm:"column:chart_type" json:"chartType"`
	Query        string    `gorm:"column:query" json:"query"`
	Filters      string    `gorm:"column:filters" json:"filters"`
	Aggregation  string    `gorm:"column:aggregation" json:"aggregation"`
	SQLQuery     string    `gorm:"column:sql_query" json:"sqlQuery"`
}

func (UtmVisualization) TableName() string { return "utm_visualization" }
