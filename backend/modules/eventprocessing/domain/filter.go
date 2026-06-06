package domain

import "time"

type UtmLogstashFilter struct {
	ID             int64      `gorm:"primaryKey;autoIncrement"                              json:"id"`
	FilterName     string     `gorm:"column:filter_name"                      json:"filterName"`
	LogstashFilter string     `gorm:"column:logstash_filter;not null"         json:"logstashFilter"`
	DataTypeID     *int64     `gorm:"column:data_type_id"                     json:"-"`
	SystemOwner    bool       `gorm:"column:system_owner;not null;default:false" json:"systemOwner"`
	IsActive       bool       `gorm:"column:is_active;not null;default:true"  json:"isActive"`
	ModuleName     *string    `gorm:"column:module_name"                      json:"moduleName"`
	FilterVersion  *string    `gorm:"column:filter_version"                   json:"filterVersion"`
	UpdatedAt      *time.Time `gorm:"column:updated_at"                       json:"updatedAt"`
}

func (UtmLogstashFilter) TableName() string {
	return "utm_logstash_filter"
}
