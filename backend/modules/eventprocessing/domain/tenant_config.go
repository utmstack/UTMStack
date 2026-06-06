package domain

import "time"

type UtmTenantConfig struct {
	ID                   int64      `gorm:"column:id;primaryKey"`
	AssetName            string     `gorm:"column:asset_name;size:250;not null"`
	AssetHostnameListDef string     `gorm:"column:asset_hostname_list_def;type:text"`
	AssetIpListDef       string     `gorm:"column:asset_ip_list_def;type:text"`
	AssetConfidentiality int        `gorm:"column:asset_confidentiality;not null"`
	AssetIntegrity       int        `gorm:"column:asset_integrity;not null"`
	AssetAvailability    int        `gorm:"column:asset_availability;not null"`
	LastUpdate           *time.Time `gorm:"column:last_update"`
}

func (UtmTenantConfig) TableName() string { return "utm_tenant_config" }
