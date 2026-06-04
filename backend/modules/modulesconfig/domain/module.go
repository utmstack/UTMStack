package domain

type UtmModule struct {
	ID                int64  `gorm:"primaryKey;autoIncrement;column:id"           json:"id"`
	ServerID          int64  `gorm:"column:server_id;index"                       json:"serverId"`
	PrettyName        string `gorm:"column:pretty_name"                           json:"prettyName"`
	ModuleName        string `gorm:"column:module_name;size:128;index:idx_utm_module_name_server,unique" json:"moduleName"`
	ModuleDescription string `gorm:"column:module_description"                    json:"moduleDescription"`
	ModuleActive      bool   `gorm:"column:module_active;default:false"           json:"moduleActive"`
	ModuleIcon        string `gorm:"column:module_icon"                           json:"moduleIcon"`
	ModuleCategory    string `gorm:"column:module_category;size:128"              json:"moduleCategory"`
	LiteVersion       bool   `gorm:"column:lite_version;default:false"            json:"liteVersion"`
	NeedsRestart      bool   `gorm:"column:needs_restart;default:false"           json:"needsRestart"`
	IsActivatable     bool   `gorm:"column:is_activatable;default:true"           json:"isActivatable"`

	ModuleGroups []UtmModuleGroup `gorm:"foreignKey:ModuleID;references:ID;constraint:OnDelete:CASCADE" json:"moduleGroups,omitempty"`
}

func (UtmModule) TableName() string { return "utm_module" }
