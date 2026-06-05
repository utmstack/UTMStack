package network_scan

import (
	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/handler"
	"github.com/utmstack/utmstack/backend/modules/network_scan/usecase"
)

type Module struct {
	networkScanHandler *handler.NetworkScanHandler
	assetGroupHandler  *handler.AssetGroupHandler
	assetTypesHandler  *handler.AssetTypesHandler
	portsHandler       *handler.PortsHandler
	probeHandler       *handler.ProbeHandler

	networkScanUC connectors.NetworkScanUsecase
	assetGroupUC  connectors.AssetGroupUsecase
	assetTypesUC  connectors.AssetTypesUsecase
	portsUC       connectors.PortsUsecase
	probeUC       connectors.ProbeUsecase
	reportUC      connectors.ReportUsecase

	scheduler *usecase.Scheduler
}

func NewModule(
	networkScanUC connectors.NetworkScanUsecase,
	assetGroupUC connectors.AssetGroupUsecase,
	assetTypesUC connectors.AssetTypesUsecase,
	portsUC connectors.PortsUsecase,
	probeUC connectors.ProbeUsecase,
	reportUC connectors.ReportUsecase,
	scheduler *usecase.Scheduler,
) *Module {
	return &Module{
		networkScanHandler: handler.NewNetworkScanHandler(networkScanUC, reportUC),
		assetGroupHandler:  handler.NewAssetGroupHandler(assetGroupUC),
		assetTypesHandler:  handler.NewAssetTypesHandler(assetTypesUC),
		portsHandler:       handler.NewPortsHandler(portsUC),
		probeHandler:       handler.NewProbeHandler(probeUC),
		networkScanUC:      networkScanUC,
		assetGroupUC:       assetGroupUC,
		assetTypesUC:       assetTypesUC,
		portsUC:            portsUC,
		probeUC:            probeUC,
		reportUC:           reportUC,
		scheduler:          scheduler,
	}
}

func (m *Module) GetNetworkScanHandler() *handler.NetworkScanHandler { return m.networkScanHandler }
func (m *Module) GetAssetGroupHandler() *handler.AssetGroupHandler   { return m.assetGroupHandler }
func (m *Module) GetAssetTypesHandler() *handler.AssetTypesHandler   { return m.assetTypesHandler }
func (m *Module) GetPortsHandler() *handler.PortsHandler             { return m.portsHandler }
func (m *Module) GetProbeHandler() *handler.ProbeHandler             { return m.probeHandler }

// Exposed for cross-module consumers (e.g. the alerts scheduler's TODO(module-33)).
func (m *Module) GetNetworkScanUsecase() connectors.NetworkScanUsecase { return m.networkScanUC }

// Scheduler returns the scheduled sync runner (nil when the sync feature is disabled).
func (m *Module) Scheduler() *usecase.Scheduler { return m.scheduler }
