package dashboards

import (
	"github.com/utmstack/utmstack/backend/modules/dashboards/handler"
	"github.com/utmstack/utmstack/backend/modules/dashboards/repository"
	"github.com/utmstack/utmstack/backend/modules/dashboards/usecase"
	"gorm.io/gorm"
)

type Module struct {
	dashboardHandler     *handler.DashboardHandler
	visualizationHandler *handler.VisualizationHandler
	layoutHandler        *handler.LayoutHandler
}

func NewModule(db *gorm.DB) *Module {
	dashRepo := repository.NewDashboardRepository(db)
	vizRepo := repository.NewVisualizationRepository(db)
	layoutRepo := repository.NewLayoutRepository(db)

	return &Module{
		dashboardHandler:     handler.NewDashboardHandler(usecase.NewDashboardUsecase(dashRepo)),
		visualizationHandler: handler.NewVisualizationHandler(usecase.NewVisualizationUsecase(vizRepo)),
		layoutHandler:        handler.NewLayoutHandler(usecase.NewLayoutUsecase(layoutRepo)),
	}
}

func (m *Module) GetDashboardHandler() *handler.DashboardHandler { return m.dashboardHandler }
func (m *Module) GetVisualizationHandler() *handler.VisualizationHandler {
	return m.visualizationHandler
}
func (m *Module) GetLayoutHandler() *handler.LayoutHandler { return m.layoutHandler }
