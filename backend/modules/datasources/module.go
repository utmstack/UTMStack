package datasources

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/datasources/handler"
	"github.com/utmstack/utmstack/backend/modules/datasources/usecase"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
)

type Module struct {
	datasourceHandler    *handler.DatasourceHandler
	assetGroupHandler    *handler.AssetGroupHandler
	connectionKeyHandler *handler.ConnectionKeyHandler
	reconciler           *usecase.StatsReconciler
}

func NewModule(dsUC connectors.DatasourceUsecase, groupUC connectors.AssetGroupUsecase, reconciler *usecase.StatsReconciler, license connectors.LicenseCapProvider, agentClient *agentmanager.AgentManagerClient) *Module {
	return &Module{
		datasourceHandler:    handler.NewDatasourceHandler(dsUC, license),
		assetGroupHandler:    handler.NewAssetGroupHandler(groupUC),
		connectionKeyHandler: handler.NewConnectionKeyHandler(agentClient),
		reconciler:           reconciler,
	}
}

func (m *Module) Start(ctx context.Context) {
	if m.reconciler != nil {
		go m.reconciler.Start(ctx)
	}
}

func (m *Module) GetDatasourceHandler() *handler.DatasourceHandler { return m.datasourceHandler }
func (m *Module) GetAssetGroupHandler() *handler.AssetGroupHandler { return m.assetGroupHandler }
func (m *Module) GetConnectionKeyHandler() *handler.ConnectionKeyHandler {
	return m.connectionKeyHandler
}
