package connectors

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
)

type NetworkScanRepository interface {
	FindByID(ctx context.Context, id uint64) (*domain.UtmNetworkScan, error)
	FindByIDWithDetails(ctx context.Context, id uint64) (*domain.UtmNetworkScan, error)
	FindByAssetName(ctx context.Context, name string) (*domain.UtmNetworkScan, error)
	FindByNameOrIP(ctx context.Context, q string) (*domain.UtmNetworkScan, error)
	FindByAssetIPsOrNames(ctx context.Context, ips, names []string) ([]domain.UtmNetworkScan, error)

	ListByCriteria(ctx context.Context, c domain.NetworkScanCriteria, p domain.Pagination) ([]domain.UtmNetworkScan, int64, error)
	CountByCriteria(ctx context.Context, c domain.NetworkScanCriteria) (int64, error)
	SearchByFilters(ctx context.Context, f domain.NetworkScanFilter, p domain.Pagination) ([]domain.UtmNetworkScan, int64, error)
	SearchPropertyValues(ctx context.Context, prop domain.Property, value string, forGroups bool, p domain.Pagination) ([]string, error)
	CountByStatus(ctx context.Context, status domain.AssetStatus) (int64, error)

	Save(ctx context.Context, scan *domain.UtmNetworkScan) error
	SaveAll(ctx context.Context, scans []domain.UtmNetworkScan) error
	UpdateType(ctx context.Context, assetIDs []uint64, typeID *uint64) error
	UpdateGroup(ctx context.Context, assetIDs []uint64, groupID *uint64) error
	Delete(ctx context.Context, id uint64) error

	FindAllByUpdateLevelIn(ctx context.Context, levels []domain.UpdateLevel) ([]domain.UtmNetworkScan, error)
	FindAllAssetGroupMappings(ctx context.Context) ([]domain.AssetGroupMapping, error)
	FindAgentsOSPlatforms(ctx context.Context) ([]string, error)
	FindAgentNamesByPlatform(ctx context.Context, platform string) ([]string, error)
	FindAllAgents(ctx context.Context) ([]domain.UtmNetworkScan, error)
}

type PortsRepository interface {
	Save(ctx context.Context, p *domain.UtmPorts) error
	SaveAll(ctx context.Context, ports []domain.UtmPorts) error
	FindByID(ctx context.Context, id uint64) (*domain.UtmPorts, error)
	ListByCriteria(ctx context.Context, c domain.PortsCriteria, p domain.Pagination) ([]domain.UtmPorts, int64, error)
	CountByCriteria(ctx context.Context, c domain.PortsCriteria) (int64, error)
	Delete(ctx context.Context, id uint64) error
	DeleteByScanID(ctx context.Context, scanID uint64) error
	DeleteByScanIDIn(ctx context.Context, scanIDs []uint64) error
	UpdateInBatchForScan(ctx context.Context, scanID uint64, ports []domain.UtmPorts) ([]domain.UtmPorts, error)
}

type AssetGroupRepository interface {
	Save(ctx context.Context, g *domain.UtmAssetGroup) error
	FindByID(ctx context.Context, id uint64) (*domain.UtmAssetGroup, error)
	List(ctx context.Context, p domain.Pagination) ([]domain.UtmAssetGroup, int64, error)
	SearchByFilter(ctx context.Context, f domain.AssetGroupFilter, p domain.Pagination) ([]AssetGroupRow, int64, error)
	Delete(ctx context.Context, id uint64) error
}

// AssetGroupRow is the aggregate projection (group + asset count) returned by SearchByFilter.
type AssetGroupRow struct {
	Group       domain.UtmAssetGroup
	AssetsCount int64
}

type AssetTypesRepository interface {
	List(ctx context.Context, p domain.Pagination) ([]domain.UtmAssetTypes, int64, error)
}

// ProbeClient is the HTTP client to the external scanning probe (the Java NetworkScanService
// delegated to a sibling microservice over REST — same pattern preserved here).
type ProbeClient interface {
	Scan(ctx context.Context, probeID uint64, iface string, networkRange string) ([]domain.NetworkScan, error)
	Ping(ctx context.Context, probeID uint64) error
	CheckInterface(ctx context.Context, probeID uint64, iface string) (bool, error)
}

// DataInputStatusGateway is the subset of utm_data_input_status access that network_scan needs.
// Network_scan owns this gateway (rather than reusing the datainput module's repo) because
// the sync job needs methods the datainput repo does not expose (DeleteByDataSource, FindAll,
// Upsert), and we want to avoid mutating that module's public interface from here.
type DataInputStatusGateway interface {
	DeleteByDataSource(ctx context.Context, dataSource string) error
	FindAll(ctx context.Context) ([]DataInputStatus, error)
	Upsert(ctx context.Context, s DataInputStatus) error
	GetCheckpoint(ctx context.Context) (*Checkpoint, error)
	SaveCheckpoint(ctx context.Context, cp Checkpoint) error
}

// DataInputStatus is a local projection of utm_data_input_status — duplicated here to keep
// network_scan independent of the datainput module's domain package.
type DataInputStatus struct {
	ID        string
	Source    string
	DataType  string
	Timestamp int64 // epoch seconds
	Median    *int64
	Alias     *string
}

type Checkpoint struct {
	ID                     int64
	LastProcessedTimestamp time.Time
}

// AgentGateway is the subset of the agentmanager client that network_scan needs.
type AgentGateway interface {
	DeleteAgentByName(ctx context.Context, name string) error
	ListAgentNames(ctx context.Context) ([]string, error)
}
