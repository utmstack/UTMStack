package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"gorm.io/datatypes"
)

type DatasourceDTO struct {
	ID           uint64         `json:"id,omitempty"`
	Name         string         `json:"name" binding:"required"`
	DataType     string         `json:"dataType,omitempty"` // join key with ingestion stats; "" for agents
	IP           string         `json:"ip,omitempty"`
	SourceKind   string         `json:"sourceKind,omitempty"`
	Metadata     datatypes.JSON `json:"metadata,omitempty"` // free-form host info
	Labels       string         `json:"labels,omitempty"`   // comma-separated; frontend splits
	// Asset sensitivity (CIA, 0–3) — weights alert impact in the correlation engine.
	AssetConfidentiality int            `json:"assetConfidentiality"`
	AssetIntegrity       int            `json:"assetIntegrity"`
	AssetAvailability    int            `json:"assetAvailability"`
	Group                *AssetGroupRef `json:"group,omitempty"`
	DiscoveredAt         *time.Time     `json:"discoveredAt,omitempty"`
	ModifiedAt   *time.Time     `json:"modifiedAt,omitempty"`
	LastPingAt   *time.Time     `json:"lastPingAt,omitempty"`
}

type AssetGroupRef struct {
	ID               uint64 `json:"id"`
	GroupName        string `json:"groupName"`
	GroupDescription string `json:"groupDescription,omitempty"`
}

type DatasourceEnrichment struct {
	Name      string   `json:"name"`
	DataType  string   `json:"dataType,omitempty"`
	GroupID   *uint64  `json:"groupId,omitempty"`
	GroupName string   `json:"groupName,omitempty"`
	Labels    []string `json:"labels,omitempty"`
}

type UpdateGroupRequest struct {
	IDs     []uint64 `json:"ids" binding:"required,min=1"`
	GroupID *uint64  `json:"groupId"`
}

type UpdateLabelsRequest struct {
	ID     uint64 `json:"id" binding:"required"`
	Labels string `json:"labels"`
}

type UpdateSensitivityRequest struct {
	ID                   uint64 `json:"id" binding:"required"`
	AssetConfidentiality int    `json:"assetConfidentiality"`
	AssetIntegrity       int    `json:"assetIntegrity"`
	AssetAvailability    int    `json:"assetAvailability"`
}

type PingRequest struct {
	Datasources []PingEntry `json:"datasources" binding:"required,min=1,dive"`
}

type RegisterRequest struct {
	SourceRef  string         // stable origin identity (upsert key), e.g. "o365:Acme"
	Name       string         // display name == dataSource in OpenSearch (the tenant name)
	DataType   string         // == dataType in OpenSearch (the plugin/module)
	SourceKind string         // puller
	IP         string         // optional
	Metadata   datatypes.JSON // optional
}

type PingEntry struct {
	SourceRef  string         `json:"sourceRef" binding:"required"`  // stable origin identity (the upsert key)
	Name       string         `json:"name" binding:"required"`       // display name (== dataSource in OpenSearch)
	DataType   string         `json:"dataType,omitempty"`            // == dataType in OpenSearch; empty for agents
	SourceKind string         `json:"sourceKind" binding:"required"` // agent | puller | direct
	IP         string         `json:"ip,omitempty"`
	Metadata   datatypes.JSON `json:"metadata,omitempty"`
	LastPingAt *time.Time     `json:"lastPingAt,omitempty"`
}

func ToDatasourceDTO(e *domain.Datasource) DatasourceDTO {
	out := DatasourceDTO{
		ID:           e.ID,
		Name:         e.Name,
		DataType:     e.DataType,
		IP:           e.IP,
		SourceKind:   e.SourceKind,
		Metadata:     e.Metadata,
		Labels:       e.Labels,
		AssetConfidentiality: e.AssetConfidentiality,
		AssetIntegrity:       e.AssetIntegrity,
		AssetAvailability:    e.AssetAvailability,
		DiscoveredAt: e.DiscoveredAt,
		ModifiedAt:   e.ModifiedAt,
		LastPingAt:   e.LastPingAt,
	}
	if e.Group != nil {
		out.Group = &AssetGroupRef{
			ID:               e.Group.ID,
			GroupName:        e.Group.GroupName,
			GroupDescription: e.Group.GroupDescription,
		}
	}
	return out
}

func FromDatasourceDTO(in DatasourceDTO) *domain.Datasource {
	d := &domain.Datasource{
		ID:         in.ID,
		Name:       in.Name,
		IP:         in.IP,
		SourceKind: in.SourceKind,
		Metadata:   in.Metadata,
		Labels:     in.Labels,
	}
	if in.Group != nil {
		id := in.Group.ID
		d.GroupID = &id
	}
	return d
}
