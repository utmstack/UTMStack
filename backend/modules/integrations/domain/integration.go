package domain

import "github.com/google/uuid"

type IngestType string

const (
	IngestTypeAgent     IngestType = "agent"
	IngestTypeCollector IngestType = "collector"
	IngestTypeForwarder IngestType = "forwarder"
	IngestTypePlugin    IngestType = "plugin"
)

func (t IngestType) Valid() bool {
	switch t {
	case IngestTypeAgent, IngestTypeCollector, IngestTypeForwarder, IngestTypePlugin:
		return true
	default:
		return false
	}
}

type Integration struct {
	ID          uuid.UUID  `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID  `gorm:"column:tenant_id;type:uuid;not null;uniqueIndex:idx_integration_tenant_name,priority:1;uniqueIndex:idx_integration_tenant_data_type,priority:1"`
	Name        string     `gorm:"column:name;size:128;not null;uniqueIndex:idx_integration_tenant_name,priority:2"`
	DataType    string     `gorm:"column:data_type;size:250;not null;uniqueIndex:idx_integration_tenant_data_type,priority:2"`
	Description string     `gorm:"column:description"`
	Icon        string     `gorm:"column:icon"`
	IngestType  IngestType `gorm:"column:ingest_type;type:varchar(32);not null;check:chk_integration_ingest_type,ingest_type IN ('agent','collector','forwarder','plugin')"`
	SystemOwner bool       `gorm:"column:system_owner;not null;default:false"`
}

func (Integration) TableName() string { return "integrations" }

func (Integration) SystemFlagColumn() string { return "system_owner" }
