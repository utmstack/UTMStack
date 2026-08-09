package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type SourceKind string

const (
	SourceKindAgent  SourceKind = "agent"
	SourceKindPuller SourceKind = "puller"
	SourceKindDirect SourceKind = "direct"
)

type Datasource struct {
	ID                   uuid.UUID      `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID             uuid.UUID      `gorm:"column:tenant_id;type:uuid;not null;uniqueIndex:idx_datasource_identity,priority:1" json:"-"`
	DataType             string         `gorm:"column:data_type;size:250;not null;uniqueIndex:idx_datasource_identity,priority:2"`
	Name                 string         `gorm:"column:asset_name;size:255;not null;index;uniqueIndex:idx_datasource_identity,priority:3"`
	IP                   string         `gorm:"column:asset_ip;size:255"`
	SourceKind           SourceKind     `gorm:"column:source_kind;size:20"`
	Metadata             datatypes.JSON `gorm:"column:metadata;type:jsonb"` // free-form host info (mac/os/osPlatform/addresses) — agents only
	Labels               string         `gorm:"column:labels;size:255"`     // comma-separated free-text tags — frontend splits; filtering only
	AssetConfidentiality int            `gorm:"column:asset_confidentiality;not null;default:0"`
	AssetIntegrity       int            `gorm:"column:asset_integrity;not null;default:0"`
	AssetAvailability    int            `gorm:"column:asset_availability;not null;default:0"`
	DiscoveredAt         *time.Time     `gorm:"column:discovered_at"`
	ModifiedAt           *time.Time     `gorm:"column:modified_at"`
	LastPingAt           *time.Time     `gorm:"column:last_ping_at"` // liveness fact; status derived (frontend) from staleness
}

func (Datasource) TableName() string { return "datasources" }
