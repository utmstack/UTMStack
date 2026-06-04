package domain

const (
	ConfTypeText     = "text"
	ConfTypePassword = "password"
	ConfTypeFile     = "file"
	ConfTypeBoolean  = "boolean"
	ConfTypeNumber   = "number"
)

const MaskedValue = "*****"

func IsSensitive(confDataType string) bool {
	return confDataType == ConfTypePassword || confDataType == ConfTypeFile
}

type UtmModuleGroupConfiguration struct {
	ID              int64  `gorm:"primaryKey;autoIncrement;column:id"   json:"id"`
	GroupID         int64  `gorm:"column:group_id;not null;index:idx_utm_module_group_conf_key,unique" json:"groupId"`
	ConfKey         string `gorm:"column:conf_key;not null;size:255;index:idx_utm_module_group_conf_key,unique" json:"confKey"`
	ConfValue       string `gorm:"column:conf_value;type:text"          json:"confValue"`
	ConfName        string `gorm:"column:conf_name;size:255"            json:"confName"`
	ConfDescription string `gorm:"column:conf_description;type:text"    json:"confDescription"`
	ConfDataType    string `gorm:"column:conf_data_type;not null;size:64" json:"confDataType"`
	ConfRequired    bool   `gorm:"column:conf_required;not null;default:false" json:"confRequired"`
	ConfOptions     string `gorm:"column:conf_options;type:text"        json:"confOptions"`
	ConfVisibility  string `gorm:"column:conf_visibility;size:64"       json:"confVisibility"`
}

func (UtmModuleGroupConfiguration) TableName() string { return "utm_module_group_configuration" }
