package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
)

// NetworkScanDTO is the full asset payload (mirrors Java NetworkScanDTO).
// JSON tags are lowerCamelCase to match the existing frontend contract.
type NetworkScanDTO struct {
	ID                  uint64       `json:"id,omitempty"`
	AssetIP             string       `json:"assetIp"`
	AssetAddresses      string       `json:"assetAddresses,omitempty"`
	AssetMAC            string       `json:"assetMac,omitempty"`
	AssetOS             string       `json:"assetOs,omitempty"`
	AssetOSArch         string       `json:"assetOsArch,omitempty"`
	AssetOSMajorVersion string       `json:"assetOsMajorVersion,omitempty"`
	AssetOSMinorVersion string       `json:"assetOsMinorVersion,omitempty"`
	AssetOSPlatform     string       `json:"assetOsPlatform,omitempty"`
	AssetOSVersion      string       `json:"assetOsVersion,omitempty"`
	AssetName           string       `json:"assetName"`
	AssetAliases        string       `json:"assetAliases,omitempty"`
	AssetAlias          string       `json:"assetAlias,omitempty"`
	ServerName          string       `json:"serverName,omitempty"`
	AssetAlive          *bool        `json:"assetAlive,omitempty"`
	RegisteredMode      string       `json:"registeredMode,omitempty"`
	AssetStatus         string       `json:"assetStatus,omitempty"`
	AssetSeverityMetric *float64     `json:"assetSeverityMetric,omitempty"`
	AssetType           *AssetTypeRef `json:"assetType,omitempty"`
	AssetNotes          string       `json:"assetNotes,omitempty"`
	DiscoveredAt        *time.Time   `json:"discoveredAt,omitempty"`
	ModifiedAt          *time.Time   `json:"modifiedAt,omitempty"`
	Group               *AssetGroupRef `json:"group,omitempty"`
	IsAgent             *bool        `json:"isAgent,omitempty"`
	Metrics             map[string]int64 `json:"metrics,omitempty"`
	Ports               []ProbePort  `json:"ports,omitempty"`
	DataInputList       []string     `json:"dataInputList,omitempty"`
}

type AssetTypeRef struct {
	ID       uint64 `json:"id"`
	TypeName string `json:"typeName"`
}

type AssetGroupRef struct {
	ID               uint64 `json:"id"`
	GroupName        string `json:"groupName"`
	GroupDescription string `json:"groupDescription,omitempty"`
}

type ProbePort struct {
	Port int    `json:"port"`
	TCP  string `json:"tcp,omitempty"`
	UDP  string `json:"udp,omitempty"`
}

type NetworkScanListResponse struct {
	Data     []NetworkScanDTO `json:"data"`
	PageInfo PageInfo         `json:"pageInfo"`
}

type UpdateTypeRequest struct {
	AssetIDs    []uint64 `json:"assetsIds" binding:"required,min=1"`
	AssetTypeID *uint64  `json:"assetTypeId"`
}

type UpdateGroupRequest struct {
	AssetIDs     []uint64 `json:"assetsIds" binding:"required,min=1"`
	AssetGroupID *uint64  `json:"assetGroupId" binding:"required"`
}

// ProbeScanRequest is the input for POST /network-scans/probe/scan.
type ProbeScanRequest struct {
	ProbeID      uint64 `json:"probeId" binding:"required"`
	Interface    string `json:"interface" binding:"required"`
	NetworkRange string `json:"networkRange" binding:"required"`
}

// ToNetworkScanDTO converts a domain entity to the DTO. When `details` is true the
// caller should have preloaded AssetType, AssetGroup, Ports (and supplied metrics).
func ToNetworkScanDTO(e *domain.UtmNetworkScan, details bool, metrics map[string]int64) NetworkScanDTO {
	out := NetworkScanDTO{
		ID:                  e.ID,
		AssetIP:             e.AssetIP,
		AssetAddresses:      e.AssetAddresses,
		AssetMAC:            e.AssetMAC,
		AssetOS:             e.AssetOS,
		AssetOSArch:         e.AssetOSArch,
		AssetOSMajorVersion: e.AssetOSMajorVersion,
		AssetOSMinorVersion: e.AssetOSMinorVersion,
		AssetOSPlatform:     e.AssetOSPlatform,
		AssetOSVersion:      e.AssetOSVersion,
		AssetName:           e.AssetName,
		AssetAliases:        e.AssetAliases,
		AssetAlias:          e.AssetAlias,
		ServerName:          e.ServerName,
		AssetAlive:          e.AssetAlive,
		RegisteredMode:      e.RegisteredMode,
		AssetStatus:         e.AssetStatus,
		AssetSeverityMetric: e.AssetSeverityMetric,
		AssetNotes:          e.AssetNotes,
		DiscoveredAt:        e.DiscoveredAt,
		ModifiedAt:          e.ModifiedAt,
		IsAgent:             e.IsAgent,
	}
	if e.AssetType != nil {
		out.AssetType = &AssetTypeRef{ID: e.AssetType.ID, TypeName: e.AssetType.TypeName}
	}
	if e.AssetGroup != nil {
		out.Group = &AssetGroupRef{
			ID:               e.AssetGroup.ID,
			GroupName:        e.AssetGroup.GroupName,
			GroupDescription: e.AssetGroup.GroupDescription,
		}
	}
	if details {
		if len(e.Ports) > 0 {
			ports := make([]ProbePort, 0, len(e.Ports))
			for _, p := range e.Ports {
				ports = append(ports, ProbePort{Port: p.Port, TCP: p.TCP, UDP: p.UDP})
			}
			out.Ports = ports
		}
		if metrics != nil {
			out.Metrics = metrics
		}
	}
	return out
}

// FromNetworkScanDTO maps the inbound DTO onto a domain entity for save/update.
// Server-controlled fields (ID, DiscoveredAt, ModifiedAt) are left for the caller.
func FromNetworkScanDTO(in NetworkScanDTO) *domain.UtmNetworkScan {
	e := &domain.UtmNetworkScan{
		ID:                  in.ID,
		AssetIP:             in.AssetIP,
		AssetAddresses:      in.AssetAddresses,
		AssetMAC:            in.AssetMAC,
		AssetOS:             in.AssetOS,
		AssetOSArch:         in.AssetOSArch,
		AssetOSMajorVersion: in.AssetOSMajorVersion,
		AssetOSMinorVersion: in.AssetOSMinorVersion,
		AssetOSPlatform:     in.AssetOSPlatform,
		AssetOSVersion:      in.AssetOSVersion,
		AssetName:           in.AssetName,
		AssetAliases:        in.AssetAliases,
		AssetAlias:          in.AssetAlias,
		ServerName:          in.ServerName,
		AssetAlive:          in.AssetAlive,
		RegisteredMode:      in.RegisteredMode,
		AssetStatus:         in.AssetStatus,
		AssetSeverityMetric: in.AssetSeverityMetric,
		AssetNotes:          in.AssetNotes,
		IsAgent:             in.IsAgent,
	}
	if in.AssetType != nil {
		id := in.AssetType.ID
		e.AssetTypeID = &id
	}
	if in.Group != nil {
		id := in.Group.ID
		e.GroupID = &id
	}
	return e
}
