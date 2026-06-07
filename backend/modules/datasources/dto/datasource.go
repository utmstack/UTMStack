package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"gorm.io/datatypes"
)

type DatasourceDTO struct {
	ID             uint64         `json:"id,omitempty"`
	Name           string         `json:"name" binding:"required"`
	IP             string         `json:"ip,omitempty"`
	SourceKind     string         `json:"sourceKind,omitempty"`
	Metadata       datatypes.JSON `json:"metadata,omitempty"` // free-form host info
	Labels         string         `json:"labels,omitempty"`   // comma-separated; frontend splits
	Group          *AssetGroupRef `json:"group,omitempty"`
	DiscoveredAt   *time.Time     `json:"discoveredAt,omitempty"`
	ModifiedAt     *time.Time     `json:"modifiedAt,omitempty"`
	LastPingAt     *time.Time     `json:"lastPingAt,omitempty"`
}

type AssetGroupRef struct {
	ID               uint64 `json:"id"`
	GroupName        string `json:"groupName"`
	GroupDescription string `json:"groupDescription,omitempty"`
}

type UpdateGroupRequest struct {
	IDs     []uint64 `json:"ids" binding:"required,min=1"`
	GroupID *uint64  `json:"groupId"`
}

type UpdateLabelsRequest struct {
	ID     uint64 `json:"id" binding:"required"`
	Labels string `json:"labels"`
}

type PingRequest struct {
	Datasources []PingEntry `json:"datasources" binding:"required,min=1,dive"`
}

type PingEntry struct {
	Name       string         `json:"name" binding:"required"`
	SourceKind string         `json:"sourceKind" binding:"required"` // agent | collector | puller | direct
	IP         string         `json:"ip,omitempty"`
	Metadata   datatypes.JSON `json:"metadata,omitempty"`
	LastPingAt *time.Time     `json:"lastPingAt,omitempty"`
}

func ToDatasourceDTO(e *domain.Datasource) DatasourceDTO {
	out := DatasourceDTO{
		ID:             e.ID,
		Name:           e.Name,
		IP:             e.IP,
		SourceKind:     e.SourceKind,
		Metadata:       e.Metadata,
		Labels:         e.Labels,
		DiscoveredAt:   e.DiscoveredAt,
		ModifiedAt:     e.ModifiedAt,
		LastPingAt:     e.LastPingAt,
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
