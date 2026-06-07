package loganalyzer

import (
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/handler"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/repository"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/usecase"
	"gorm.io/gorm"
)

type Module struct {
	queryHandler    *handler.QueryHandler
	analyzerHandler *handler.AnalyzerHandler
}

func NewModule(db *gorm.DB) *Module {
	queryRepo := repository.NewQueryRepository(db)
	analyzerRepo := repository.NewAnalyzerRepository()

	return &Module{
		queryHandler:    handler.NewQueryHandler(usecase.NewQueryUsecase(queryRepo)),
		analyzerHandler: handler.NewAnalyzerHandler(usecase.NewAnalyzerUsecase(analyzerRepo)),
	}
}

func (m *Module) GetQueryHandler() *handler.QueryHandler       { return m.queryHandler }
func (m *Module) GetAnalyzerHandler() *handler.AnalyzerHandler { return m.analyzerHandler }
