package domain

import "time"

// NetworkScanFilter mirrors Java NetworkScanFilter — used by /search-by-filters.
type NetworkScanFilter struct {
	AssetIPMacName     string     `form:"assetIpMacName"`
	OS                 []string   `form:"os"`
	Alive              []bool     `form:"alive"`
	Status             []string   `form:"status"`
	Type               []uint64   `form:"type"`
	Alias              []string   `form:"alias"`
	Probe              []string   `form:"probe"`
	OpenPorts          []int      `form:"openPorts"`
	Groups             []uint64   `form:"groups"`
	DiscoveredInitDate *time.Time `form:"discoveredInitDate" time_format:"2006-01-02T15:04:05Z07:00"`
	DiscoveredEndDate  *time.Time `form:"discoveredEndDate" time_format:"2006-01-02T15:04:05Z07:00"`
	RegisteredMode     string     `form:"registeredMode"`
	Agent              []bool     `form:"agent"`
	OSPlatform         []string   `form:"osPlatform"`
	DataTypes          []string   `form:"dataTypes"`
}

// AssetGroupFilter mirrors Java AssetGroupFilter — used by /asset-groups/searchGroupsByFilter.
type AssetGroupFilter struct {
	ID        *uint64    `form:"id"`
	GroupName string     `form:"groupName"`
	Type      []uint64   `form:"type"`
	InitDate  *time.Time `form:"initDate" time_format:"2006-01-02T15:04:05Z07:00"`
	EndDate   *time.Time `form:"endDate" time_format:"2006-01-02T15:04:05Z07:00"`
	Probe     []string   `form:"probe"`
	OS        []string   `form:"os"`
	AssetIP   []string   `form:"assetIp"`
	AssetName []string   `form:"assetName"`
	AssetType *uint64    `form:"assetType"`
}

// NetworkScanCriteria is a slimmer criteria object for the generic /utm-network-scans listing.
type NetworkScanCriteria struct {
	ID           *uint64    `form:"id"`
	IP           string     `form:"ip"`
	MAC          string     `form:"mac"`
	OS           string     `form:"os"`
	Name         string     `form:"name"`
	Alive        *bool      `form:"alive"`
	Status       string     `form:"status"`
	DiscoveredAt *time.Time `form:"discoveredAt" time_format:"2006-01-02T15:04:05Z07:00"`
	ModifiedAt   *time.Time `form:"modifiedAt" time_format:"2006-01-02T15:04:05Z07:00"`
}

type PortsCriteria struct {
	ID     *uint64 `form:"id"`
	ScanID *uint64 `form:"scanId"`
	Port   *int    `form:"port"`
	TCP    string  `form:"tcp"`
	UDP    string  `form:"udp"`
}

// Pagination is shared by repo methods that paginate.
type Pagination struct {
	Page     int
	PageSize int
	Sort     string // "column,asc" or "column,desc"
}

// Property describes a queryable column for /searchPropertyValues. Mirrors Java Property interface.
type Property struct {
	Name      string // column name
	FromTable string
	JoinTable string // optional; empty if single-table
}

// PropertyFilters enumerates the single-column properties exposed by /searchPropertyValues.
// Mirrors Java PropertyFilter enum.
var PropertyFilters = map[string]Property{
	"IP":              {Name: "asset_ip", FromTable: "utm_network_scan"},
	"MAC":             {Name: "asset_mac", FromTable: "utm_network_scan"},
	"ALIAS":           {Name: "asset_alias", FromTable: "utm_network_scan"},
	"NAME":            {Name: "asset_name", FromTable: "utm_network_scan"},
	"OS":              {Name: "asset_os", FromTable: "utm_network_scan"},
	"ALIVE":           {Name: "asset_alive", FromTable: "utm_network_scan"},
	"STATUS":          {Name: "asset_status", FromTable: "utm_network_scan"},
	"PROBE":           {Name: "server_name", FromTable: "utm_network_scan"},
	"TYPE":            {Name: "type_name", FromTable: "utm_asset_types", JoinTable: "utm_network_scan ON utm_network_scan.asset_type_id = utm_asset_types.id"},
	"SEVERITY":        {Name: "asset_severity_metric", FromTable: "utm_network_scan"},
	"PORTS":           {Name: "port", FromTable: "utm_ports", JoinTable: "utm_network_scan ON utm_network_scan.id = utm_ports.scan_id"},
	"GROUP":           {Name: "group_name", FromTable: "utm_asset_group", JoinTable: "utm_network_scan ON utm_network_scan.group_id = utm_asset_group.id"},
	"OS_PLATFORM":     {Name: "asset_os_platform", FromTable: "utm_network_scan"},
}

// LookupProperty resolves a property by case-insensitive key as the frontend passes them.
func LookupProperty(key string) (Property, bool) {
	p, ok := PropertyFilters[key]
	return p, ok
}
