package domain

type UtmModule struct {
	ID                int64  `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	DataType          string `gorm:"column:data_type;size:250;not null;uniqueIndex" json:"dataType"`
	PrettyName        string `gorm:"column:pretty_name"                                                  json:"prettyName"`
	ModuleName        string `gorm:"column:module_name;size:128;not null;uniqueIndex:idx_utm_module_name" json:"moduleName"`
	ModuleDescription string `gorm:"column:module_description"                                           json:"moduleDescription"`
	ModuleActive      bool   `gorm:"column:module_active;default:false"                                  json:"moduleActive"`
	ModuleIcon        string `gorm:"column:module_icon"                                                  json:"moduleIcon"`
	ModuleCategory    string `gorm:"column:module_category;size:128"                                     json:"moduleCategory"`
	IngestType        string `gorm:"column:ingest_type;size:32" json:"ingestType"`
	IsSystem          bool   `gorm:"column:is_system;not null;default:false" json:"isSystem"`
}

func (UtmModule) TableName() string { return "utm_module" }
