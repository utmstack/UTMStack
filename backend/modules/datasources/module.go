package datasources

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/datasources/handler"
	"github.com/utmstack/utmstack/backend/modules/datasources/usecase"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

type Module struct {
	datasourceHandler    *handler.DatasourceHandler
	connectionKeyHandler *handler.ConnectionKeyHandler
	reconciler           *usecase.StatsReconciler
	projection           *usecase.AssetProjection
	datasourceUC         connectors.DatasourceUsecase
}

func NewModule(dsUC connectors.DatasourceUsecase, reconciler *usecase.StatsReconciler, agentClient *agentmanager.AgentManagerClient) *Module {
	m := &Module{
		datasourceHandler:    handler.NewDatasourceHandler(dsUC),
		connectionKeyHandler: handler.NewConnectionKeyHandler(agentClient),
		reconciler:           reconciler,
		datasourceUC:         dsUC,
	}

	m.projection = usecase.NewAssetProjection(dsUC)
	if setter, ok := dsUC.(interface{ SetAssetNotifier(usecase.Notifier) }); ok {
		setter.SetAssetNotifier(m.projection)
	}
	return m
}

func (m *Module) Start(ctx context.Context) {
	if m.reconciler != nil {
		go m.reconciler.Start(ctx)
	}
	go m.projection.Start(ctx)
	if err := m.datasourceUC.ProjectAssets(tenancy.WithAllTenants(ctx)); err != nil {
		_ = catcher.Error("datasources: initial asset projection failed", err, nil)
	}
}

func (m *Module) GetDatasourceHandler() *handler.DatasourceHandler { return m.datasourceHandler }
func (m *Module) GetConnectionKeyHandler() *handler.ConnectionKeyHandler {
	return m.connectionKeyHandler
}

func (m *Module) GetDatasourceUsecase() connectors.DatasourceUsecase { return m.datasourceUC }
