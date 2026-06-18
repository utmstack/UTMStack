package dashboards

import (
	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/handler"
	"github.com/utmstack/utmstack/backend/modules/dashboards/repository"
	"github.com/utmstack/utmstack/backend/modules/dashboards/usecase"
	"gorm.io/gorm"
)

type Module struct {
	dashboardHandler     *handler.DashboardHandler
	visualizationHandler *handler.VisualizationHandler
	layoutHandler        *handler.LayoutHandler
	dashboardUC          connectors.DashboardUsecase
	visualizationUC      connectors.VisualizationUsecase
	layoutUC             connectors.LayoutUsecase
}

func NewModule(db *gorm.DB) *Module {
	dashRepo := repository.NewDashboardRepository(db)
	vizRepo := repository.NewVisualizationRepository(db)
	layoutRepo := repository.NewLayoutRepository(db)

	dashUC := usecase.NewDashboardUsecase(dashRepo)
	vizUC := usecase.NewVisualizationUsecase(vizRepo)
	layoutUC := usecase.NewLayoutUsecase(layoutRepo)

	return &Module{
		dashboardHandler:     handler.NewDashboardHandler(dashUC),
		visualizationHandler: handler.NewVisualizationHandler(vizUC),
		layoutHandler:        handler.NewLayoutHandler(layoutUC),
		dashboardUC:          dashUC,
		visualizationUC:      vizUC,
		layoutUC:             layoutUC,
	}
}

func (m *Module) GetDashboardHandler() *handler.DashboardHandler { return m.dashboardHandler }
func (m *Module) GetVisualizationHandler() *handler.VisualizationHandler {
	return m.visualizationHandler
}
func (m *Module) GetLayoutHandler() *handler.LayoutHandler { return m.layoutHandler }

func (m *Module) GetDashboardUsecase() connectors.DashboardUsecase { return m.dashboardUC }
func (m *Module) GetVisualizationUsecase() connectors.VisualizationUsecase {
	return m.visualizationUC
}
func (m *Module) GetLayoutUsecase() connectors.LayoutUsecase { return m.layoutUC }
