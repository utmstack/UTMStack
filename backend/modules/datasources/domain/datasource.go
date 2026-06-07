package domain

import (
	"time"

	"gorm.io/datatypes"
)

type Datasource struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement;column:id"`
	Name           string         `gorm:"column:asset_name;size:255;uniqueIndex"` // == dataSource in OpenSearch
	IP             string         `gorm:"column:asset_ip;size:255"`
	SourceKind     string         `gorm:"column:source_kind;size:20"` // agent | collector | puller | direct
	Metadata       datatypes.JSON `gorm:"column:metadata;type:jsonb"` // free-form host info (mac/os/osPlatform/addresses) — agents/collectors only
	Labels         string         `gorm:"column:labels;size:255"`     // comma-separated free-text tags — frontend splits; filtering only
	GroupID        *uint64        `gorm:"column:group_id"`
	DiscoveredAt   *time.Time     `gorm:"column:discovered_at"`
	ModifiedAt     *time.Time     `gorm:"column:modified_at"`
	LastPingAt     *time.Time     `gorm:"column:last_ping_at"` // liveness fact; status derived (frontend) from staleness
	Group          *UtmAssetGroup `gorm:"foreignKey:GroupID;references:ID"`
}

func (Datasource) TableName() string { return "datasources" }
