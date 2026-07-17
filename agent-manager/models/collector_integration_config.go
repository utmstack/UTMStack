package models

import "gorm.io/gorm"

type CollectorIntegrationConfig struct {
	gorm.Model
	CollectorID      uint      `gorm:"not null;index;uniqueIndex:idx_collector_data_type"`
	Collector        Collector `gorm:"foreignKey:CollectorID;references:ID;constraint:OnDelete:CASCADE"`
	DataType         string    `gorm:"column:data_type;type:varchar(255);not null;uniqueIndex:idx_collector_data_type"`
	DesiredStateJSON string    `gorm:"column:desired_state_json;type:text"`
	ConfigStatus     string    `gorm:"column:config_status;type:varchar(20)"`
	RequestID        string    `gorm:"column:request_id;type:varchar(64);index"`
	LastError        string    `gorm:"column:last_error;type:text"`
}
