package datasources

import (
	"context"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"

	"github.com/threatwinds/go-sdk/catcher"
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
	datasourceUC         connectors.DatasourceUsecase
	assetGroupUC         connectors.AssetGroupUsecase
}

func NewModule(dsUC connectors.DatasourceUsecase, groupUC connectors.AssetGroupUsecase, reconciler *usecase.StatsReconciler, agentClient *agentmanager.AgentManagerClient) *Module {
	return &Module{
		datasourceHandler:    handler.NewDatasourceHandler(dsUC),
		assetGroupHandler:    handler.NewAssetGroupHandler(groupUC),
		connectionKeyHandler: handler.NewConnectionKeyHandler(agentClient),
		reconciler:           reconciler,
		datasourceUC:         dsUC,
		assetGroupUC:         groupUC,
	}
}

func (m *Module) Start(ctx context.Context) {
	if m.reconciler != nil {
		go m.reconciler.Start(ctx)
	}

	// The projection writes one file per tenant, so it reads across all of them.
	if err := m.datasourceUC.ProjectAssets(tenancy.WithAllTenants(ctx)); err != nil {
		_ = catcher.Error("datasources: initial asset projection failed", err, nil)
	}
}

func (m *Module) GetDatasourceHandler() *handler.DatasourceHandler { return m.datasourceHandler }
func (m *Module) GetAssetGroupHandler() *handler.AssetGroupHandler { return m.assetGroupHandler }
func (m *Module) GetConnectionKeyHandler() *handler.ConnectionKeyHandler {
	return m.connectionKeyHandler
}

func (m *Module) GetDatasourceUsecase() connectors.DatasourceUsecase { return m.datasourceUC }
func (m *Module) GetAssetGroupUsecase() connectors.AssetGroupUsecase { return m.assetGroupUC }
