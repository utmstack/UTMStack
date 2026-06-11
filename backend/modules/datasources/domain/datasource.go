package domain

import (
	"time"

	"gorm.io/datatypes"
)

type Datasource struct {
	ID                   uint64         `gorm:"primaryKey;autoIncrement;column:id"`
	SourceRef            string         `gorm:"column:source_ref;size:255;uniqueIndex"`
	Name                 string         `gorm:"column:asset_name;size:255;index"` // == dataSource in OpenSearch (display, not unique)
	DataType             string         `gorm:"column:data_type;size:250"`        // == dataType in OpenSearch; "" for agents (host emits many types)
	IP                   string         `gorm:"column:asset_ip;size:255"`
	SourceKind           string         `gorm:"column:source_kind;size:20"` // agent | puller | direct
	Metadata             datatypes.JSON `gorm:"column:metadata;type:jsonb"` // free-form host info (mac/os/osPlatform/addresses) — agents only
	Labels               string         `gorm:"column:labels;size:255"`     // comma-separated free-text tags — frontend splits; filtering only
	AssetConfidentiality int            `gorm:"column:asset_confidentiality;not null;default:0"`
	AssetIntegrity       int            `gorm:"column:asset_integrity;not null;default:0"`
	AssetAvailability    int            `gorm:"column:asset_availability;not null;default:0"`
	GroupID              *uint64        `gorm:"column:group_id"`
	DiscoveredAt         *time.Time     `gorm:"column:discovered_at"`
	ModifiedAt           *time.Time     `gorm:"column:modified_at"`
	LastPingAt           *time.Time     `gorm:"column:last_ping_at"` // liveness fact; status derived (frontend) from staleness
	Group                *UtmAssetGroup `gorm:"foreignKey:GroupID;references:ID"`
}

func (Datasource) TableName() string { return "datasources" }
