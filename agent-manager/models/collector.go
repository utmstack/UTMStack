package models

import "gorm.io/gorm"

type CollectorModule string

const (
	AS_400 CollectorModule = "AS_400"
)

type Collector struct {
	gorm.Model
	CollectorKey  string               `json:"collectorKey" gorm:"uniqueIndex;type:varchar(255)"`
	Ip            string               `json:"ip" gorm:"type:varchar(100)"`
	Hostname      string               `json:"hostname" gorm:"type:varchar(255)"`
	Version       string               `json:"version" gorm:"type:varchar(100)"`
	Module        CollectorModule      `json:"module" gorm:"type:varchar(100)"`
	Configuration CollectorConfigGroup `json:"configuration" gorm:"foreignKey:CollectorID"`
	DeletedBy     string               `json:"deletedBy" gorm:"type:varchar(255)"`
}

type CollectorConfigGroup struct {
	gorm.Model
	ModuleID              uint                           `json:"moduleId"`
	GroupName             string                         `json:"groupName" gorm:"type:varchar(255)"`
	GroupDescription      string                         `json:"groupDescription" gorm:"type:text"`
	CollectorGroupConfigs []CollectorGroupConfigurations `json:"collectorGroupConfigs" gorm:"foreignKey:ConfigGroupID"`
	CollectorID           uint                           // Foreign key for Collector
}

type CollectorGroupConfigurations struct {
	gorm.Model
	ConfigGroupID   uint   `json:"groupId"` // This acts as the foreign key for CollectorConfigGroup
	ConfKey         string `json:"confKey" gorm:"type:varchar(255)"`
	ConfValue       string `json:"confValue" gorm:"type:text"`
	ConfName        string `json:"confName" gorm:"type:varchar(255)"`
	ConfDescription string `json:"confDescription" gorm:"type:text"`
	ConfDataType    string `json:"confDataType" gorm:"type:varchar(100)"`
	ConfRequired    bool   `json:"confRequired" gorm:"type:boolean"`
}
