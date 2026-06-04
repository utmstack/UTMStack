package domain

import "time"

type UtmDataInputStatusCheckpoint struct {
	ID                     int64     `gorm:"column:id;primaryKey;autoIncrement"`
	LastProcessedTimestamp time.Time `gorm:"column:last_processed_timestamp;not null"`
}

func (UtmDataInputStatusCheckpoint) TableName() string { return "utm_data_input_status_checkpoint" }
