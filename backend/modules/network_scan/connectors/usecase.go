package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
	"github.com/utmstack/utmstack/backend/modules/network_scan/dto"
)

type NetworkScanUsecase interface {
	SaveOrUpdateCustomAsset(ctx context.Context, input dto.NetworkScanDTO) error
	Update(ctx context.Context, scan *domain.UtmNetworkScan) (*domain.UtmNetworkScan, error)
	UpdateType(ctx context.Context, req dto.UpdateTypeRequest) error
	UpdateGroup(ctx context.Context, req dto.UpdateGroupRequest) error
	ListByCriteria(ctx context.Context, c domain.NetworkScanCriteria, p domain.Pagination) ([]domain.UtmNetworkScan, int64, error)
	CountByCriteria(ctx context.Context, c domain.NetworkScanCriteria) (int64, error)
	GetByID(ctx context.Context, id uint64) (*dto.NetworkScanDTO, error)
	SearchByFilters(ctx context.Context, f domain.NetworkScanFilter, p domain.Pagination) (*dto.NetworkScanListResponse, error)
	SearchPropertyValues(ctx context.Context, propKey, value string, forGroups bool, p domain.Pagination) ([]string, error)
	CountNewAssets(ctx context.Context) (int64, error)
	DeleteCustomAsset(ctx context.Context, id uint64) error
	CanRunCommand(ctx context.Context, assetName string) (bool, error)
	GetAgentsOSPlatform(ctx context.Context) ([]string, error)

	// AssetGroupsForAlerts exposes the asset-name → group mapping consumed by other modules
	// (the alerts scheduler in particular — see modules/alerts/usecase/scheduler.go:78 TODO).
	AssetGroupsForAlerts(ctx context.Context) ([]domain.AssetGroupMapping, error)
}

type PortsUsecase interface {
	Create(ctx context.Context, p *domain.UtmPorts) (*domain.UtmPorts, error)
	Update(ctx context.Context, p *domain.UtmPorts) (*domain.UtmPorts, error)
	GetByID(ctx context.Context, id uint64) (*domain.UtmPorts, error)
	ListByCriteria(ctx context.Context, c domain.PortsCriteria, p domain.Pagination) ([]domain.UtmPorts, int64, error)
	CountByCriteria(ctx context.Context, c domain.PortsCriteria) (int64, error)
	Delete(ctx context.Context, id uint64) error
}

type AssetGroupUsecase interface {
	Create(ctx context.Context, g *domain.UtmAssetGroup) (*domain.UtmAssetGroup, error)
	Update(ctx context.Context, g *domain.UtmAssetGroup) (*domain.UtmAssetGroup, error)
	GetByID(ctx context.Context, id uint64) (*domain.UtmAssetGroup, error)
	SearchByFilter(ctx context.Context, f domain.AssetGroupFilter, p domain.Pagination) (*dto.AssetGroupListResponse, error)
	Delete(ctx context.Context, id uint64) error
}

type AssetTypesUsecase interface {
	List(ctx context.Context, p domain.Pagination) ([]domain.UtmAssetTypes, int64, error)
}

type ProbeUsecase interface {
	Scan(ctx context.Context, req dto.ProbeScanRequest) ([]domain.NetworkScan, error)
	Ping(ctx context.Context, probeID uint64) error
	CheckInterface(ctx context.Context, probeID uint64, iface string) (bool, error)
}

type ReportUsecase interface {
	GenerateNetworkScanReport(ctx context.Context, f domain.NetworkScanFilter) ([]byte, error)
}
