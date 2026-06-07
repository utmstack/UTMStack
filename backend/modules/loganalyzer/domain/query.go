package domain

import "time"

type UtmLogAnalyzerQuery struct {
	ID               uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name             string    `gorm:"column:la_name;size:100;not null" json:"name"`
	Description      string    `gorm:"column:la_description" json:"description"`
	Owner            string    `gorm:"column:la_owner" json:"owner"`
	CreationDate     time.Time `gorm:"column:la_creation_date" json:"creationDate"`
	ModificationDate time.Time `gorm:"column:la_modification_date" json:"modificationDate"`
	Columns          string    `gorm:"column:la_columns" json:"columns"`
	Filters          string    `gorm:"column:la_filters" json:"filters"`
	DataOrigin       string    `gorm:"column:la_data_origin" json:"dataOrigin"`
	IDPattern        *uint64   `gorm:"column:id_pattern" json:"idPattern,omitempty"`
}

func (UtmLogAnalyzerQuery) TableName() string { return "utm_log_analyzer_query" }
