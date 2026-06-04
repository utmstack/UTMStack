package domain

type UtmModuleGroup struct {
	ID               int64   `gorm:"primaryKey;autoIncrement;column:id"  json:"id"`
	ModuleID         int64   `gorm:"column:module_id;not null;index:idx_utm_module_group_module_name,unique" json:"moduleId"`
	GroupName        string  `gorm:"column:group_name;not null;size:255;index:idx_utm_module_group_module_name,unique" json:"groupName"`
	GroupDescription string  `gorm:"column:group_description"            json:"groupDescription"`
	Collector        *string `gorm:"column:collector_id;size:128;index"  json:"collector,omitempty"`

	ModuleGroupConfigurations []UtmModuleGroupConfiguration `gorm:"foreignKey:GroupID;references:ID;constraint:OnDelete:CASCADE" json:"moduleGroupConfigurations,omitempty"`
}

func (UtmModuleGroup) TableName() string { return "utm_module_group" }
