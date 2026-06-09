package loganalyzer

import (
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/connectors"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/handler"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/repository"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/usecase"
	"gorm.io/gorm"
)

type Module struct {
	queryHandler    *handler.QueryHandler
	analyzerHandler *handler.AnalyzerHandler
	queryUC         connectors.QueryUsecase
	analyzerUC      connectors.AnalyzerUsecase
}

func NewModule(db *gorm.DB) *Module {
	queryRepo := repository.NewQueryRepository(db)
	analyzerRepo := repository.NewAnalyzerRepository()

	queryUC := usecase.NewQueryUsecase(queryRepo)
	analyzerUC := usecase.NewAnalyzerUsecase(analyzerRepo)
	return &Module{
		queryHandler:    handler.NewQueryHandler(queryUC),
		analyzerHandler: handler.NewAnalyzerHandler(analyzerUC),
		queryUC:         queryUC,
		analyzerUC:      analyzerUC,
	}
}

func (m *Module) GetQueryHandler() *handler.QueryHandler       { return m.queryHandler }
func (m *Module) GetAnalyzerHandler() *handler.AnalyzerHandler { return m.analyzerHandler }
func (m *Module) GetQueryUsecase() connectors.QueryUsecase     { return m.queryUC }
func (m *Module) GetAnalyzerUsecase() connectors.AnalyzerUsecase {
	return m.analyzerUC
}
