package domain

import "time"

type UtmNetworkScan struct {
	ID                  uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	AssetIP             string     `gorm:"column:asset_ip;size:255"`
	AssetAddresses      string     `gorm:"column:asset_addresses;type:text"`
	AssetMAC            string     `gorm:"column:asset_mac;size:255"`
	AssetOS             string     `gorm:"column:asset_os;size:255"`
	AssetName           string     `gorm:"column:asset_name;size:255;uniqueIndex"`
	AssetAliases        string     `gorm:"column:asset_aliases;type:text"`
	AssetAlive          *bool      `gorm:"column:asset_alive"`
	AssetStatus         string     `gorm:"column:asset_status;size:255"`
	AssetTypeID         *uint64    `gorm:"column:asset_type_id"`
	DiscoveredAt        *time.Time `gorm:"column:discovered_at"`
	ModifiedAt          *time.Time `gorm:"column:modified_at"`
	AssetNotes          string     `gorm:"column:asset_notes;type:text"`
	AssetSeverityMetric *float64   `gorm:"column:asset_severity_metric"`
	ServerName          string     `gorm:"column:server_name"`
	GroupID             *uint64    `gorm:"column:group_id"`
	RegisteredMode      string     `gorm:"column:registered_mode;size:20"`
	AssetAlias          string     `gorm:"column:asset_alias;size:500"`
	IsAgent             *bool      `gorm:"column:is_agent"`
	RegisterIP          string     `gorm:"column:register_ip;size:50"`
	AssetOSArch         string     `gorm:"column:asset_os_arch;size:100"`
	AssetOSMajorVersion string     `gorm:"column:asset_os_major_version;size:20"`
	AssetOSMinorVersion string     `gorm:"column:asset_os_minor_version;size:20"`
	AssetOSPlatform     string     `gorm:"column:asset_os_platform;size:100"`
	AssetOSVersion      string     `gorm:"column:asset_os_version;size:100"`
	UpdateLevel         string     `gorm:"column:update_level;size:50"`

	AssetType  *UtmAssetTypes `gorm:"foreignKey:AssetTypeID;references:ID"`
	AssetGroup *UtmAssetGroup `gorm:"foreignKey:GroupID;references:ID"`
	Ports      []UtmPorts     `gorm:"foreignKey:ScanID;references:ID"`
}

func (UtmNetworkScan) TableName() string { return "utm_network_scan" }
