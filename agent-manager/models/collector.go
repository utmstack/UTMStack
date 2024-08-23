package models

import "gorm.io/gorm"

type CollectorModule string

const (
	AS_400 CollectorModule = "AS_400"
)

type Collector struct {
	gorm.Model
	CollectorKey string          `json:"collectorKey" gorm:"uniqueIndex;type:varchar(255)"`
	Ip           string          `json:"ip" gorm:"type:varchar(100)"`
	Hostname     string          `json:"hostname" gorm:"type:varchar(255)"`
	Version      string          `json:"version" gorm:"type:varchar(100)"`
	Module       CollectorModule `json:"module" gorm:"type:varchar(100)"`
	DeletedBy    string          `json:"deletedBy" gorm:"type:varchar(255)"`
}
