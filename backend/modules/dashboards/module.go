package dashboards

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/handler"
	"github.com/utmstack/utmstack/backend/modules/dashboards/repository"
	"github.com/utmstack/utmstack/backend/modules/dashboards/usecase"
	"github.com/utmstack/utmstack/backend/pkg/env"
	"gorm.io/gorm"
)

type Module struct {
	queryHandler         *handler.QueryHandler
	dashboardHandler     *handler.DashboardHandler
	visualizationHandler *handler.VisualizationHandler
	dashboardUC          connectors.DashboardUsecase
	visualizationUC      connectors.VisualizationUsecase
	dashboardBootstrap   *repository.DashboardBootstrap
}

func NewModule(db *gorm.DB, events usecase.Reader) *Module {
	dashRepo := repository.NewDashboardRepository(db)
	vizRepo := repository.NewVisualizationRepository(db)

	dashUC := usecase.NewDashboardUsecase(dashRepo)
	vizUC := usecase.NewVisualizationUsecase(vizRepo)

	dashboardBootstrap := repository.NewDashboardBootstrap(
		env.String(repository.DashboardsSrcDirEnv, repository.DefaultDashboardsSrcDir, false), db)

	m := &Module{
		dashboardHandler:     handler.NewDashboardHandler(dashUC),
		visualizationHandler: handler.NewVisualizationHandler(vizUC),
		dashboardUC:          dashUC,
		visualizationUC:      vizUC,
		dashboardBootstrap:   dashboardBootstrap,
	}
	if events != nil {
		m.queryHandler = handler.NewQueryHandler(usecase.NewQueryService(events))
	}
	return m
}

// Start seeds the system-owned default dashboards. Safe to call every
// process start — see DashboardBootstrap's doc comment for the resync model.
func (m *Module) Start(ctx context.Context) {
	if err := m.dashboardBootstrap.Run(ctx); err != nil {
		_ = catcher.Error("dashboard bootstrap failed", err, nil)
	}
}

func (m *Module) GetDashboardHandler() *handler.DashboardHandler { return m.dashboardHandler }
func (m *Module) GetVisualizationHandler() *handler.VisualizationHandler {
	return m.visualizationHandler
}

func (m *Module) GetDashboardUsecase() connectors.DashboardUsecase { return m.dashboardUC }
func (m *Module) GetVisualizationUsecase() connectors.VisualizationUsecase {
	return m.visualizationUC
}

func (m *Module) QueryHandler() *handler.QueryHandler { return m.queryHandler }
