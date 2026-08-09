package dto

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
)

type DatasourceDTO struct {
	ID                   uuid.UUID         `json:"id"`
	Name                 string            `json:"name"`
	DataType             string            `json:"dataType,omitempty"` // join key with the ingestion stats
	IP                   string            `json:"ip,omitempty"`
	SourceKind           domain.SourceKind `json:"sourceKind,omitempty"`
	Metadata             datatypes.JSON    `json:"metadata,omitempty"` // free-form host info
	Labels               string            `json:"labels,omitempty"`   // comma-separated; frontend splits
	AssetConfidentiality int               `json:"assetConfidentiality"`
	AssetIntegrity       int               `json:"assetIntegrity"`
	AssetAvailability    int               `json:"assetAvailability"`
	DiscoveredAt         *time.Time        `json:"discoveredAt,omitempty"`
	ModifiedAt           *time.Time        `json:"modifiedAt,omitempty"`
	LastPingAt           *time.Time        `json:"lastPingAt,omitempty"`
}

type UpdateLabelsRequest struct {
	ID     uuid.UUID `json:"id" binding:"required"`
	Labels string    `json:"labels"`
}

type UpdateSensitivityRequest struct {
	ID                   uuid.UUID `json:"id" binding:"required"`
	AssetConfidentiality int       `json:"assetConfidentiality"`
	AssetIntegrity       int       `json:"assetIntegrity"`
	AssetAvailability    int       `json:"assetAvailability"`
}

type PingRequest struct {
	Datasources []PingEntry `json:"datasources" binding:"required,min=1,dive"`
}

type RegisterRequest struct {
	Name       string
	DataType   string
	SourceKind domain.SourceKind
	IP         string
	Metadata   datatypes.JSON
}

type PingEntry struct {
	TenantID   uuid.UUID         `json:"tenantId,omitempty"`
	Name       string            `json:"name" binding:"required"`
	DataType   string            `json:"dataType,omitempty"`
	SourceKind domain.SourceKind `json:"sourceKind" binding:"required,oneof=agent puller direct"`
	IP         string            `json:"ip,omitempty"`
	Metadata   datatypes.JSON    `json:"metadata,omitempty"`
	LastPingAt *time.Time        `json:"lastPingAt,omitempty"`
}

func ToDatasourceDTO(e *domain.Datasource) DatasourceDTO {
	return DatasourceDTO{
		ID:                   e.ID,
		Name:                 e.Name,
		DataType:             e.DataType,
		IP:                   e.IP,
		SourceKind:           e.SourceKind,
		Metadata:             e.Metadata,
		Labels:               e.Labels,
		AssetConfidentiality: e.AssetConfidentiality,
		AssetIntegrity:       e.AssetIntegrity,
		AssetAvailability:    e.AssetAvailability,
		DiscoveredAt:         e.DiscoveredAt,
		ModifiedAt:           e.ModifiedAt,
		LastPingAt:           e.LastPingAt,
	}
}
