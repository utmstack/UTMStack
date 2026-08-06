package dashboards

import (
	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/handler"
	"github.com/utmstack/utmstack/backend/modules/dashboards/repository"
	"github.com/utmstack/utmstack/backend/modules/dashboards/usecase"
	"gorm.io/gorm"
)

type Module struct {
	queryHandler         *handler.QueryHandler
	dashboardHandler     *handler.DashboardHandler
	visualizationHandler *handler.VisualizationHandler
	dashboardUC          connectors.DashboardUsecase
	visualizationUC      connectors.VisualizationUsecase
}

func NewModule(db *gorm.DB, events usecase.Reader) *Module {
	dashRepo := repository.NewDashboardRepository(db)
	vizRepo := repository.NewVisualizationRepository(db)

	dashUC := usecase.NewDashboardUsecase(dashRepo)
	vizUC := usecase.NewVisualizationUsecase(vizRepo)

	m := &Module{
		dashboardHandler:     handler.NewDashboardHandler(dashUC),
		visualizationHandler: handler.NewVisualizationHandler(vizUC),
		dashboardUC:          dashUC,
		visualizationUC:      vizUC,
	}
	if events != nil {
		m.queryHandler = handler.NewQueryHandler(usecase.NewQueryService(events))
	}
	return m
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
