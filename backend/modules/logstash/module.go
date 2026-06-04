package logstash

import (
	"context"

	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	"github.com/utmstack/utmstack/backend/modules/logstash/handler"
	"github.com/utmstack/utmstack/backend/modules/logstash/repository"
	"github.com/utmstack/utmstack/backend/modules/logstash/usecase"
	"gorm.io/gorm"
)

type Module struct {
	filterGroupHandler *handler.FilterGroupHandler
	filterHandler      *handler.FilterHandler
	pipelineHandler    *handler.PipelineHandler

	pipelineScheduler *usecase.PipelineScheduler
	filterDefSync     *usecase.FilterDefinitionSyncService

	auditLogger audit_connectors.Logger //nolint:unused
}

func NewModule(db *gorm.DB, auditLogger audit_connectors.Logger) *Module {
	fgRepo := repository.NewFilterGroupRepository(db)
	fgUC := usecase.NewFilterGroupUsecase(fgRepo)
	fgHandler := handler.NewFilterGroupHandler(fgUC)

	filterRepo := repository.NewFilterRepository(db)
	pipelineFilterRepo := repository.NewPipelineFilterRepository(db)
	pipelineRepo := repository.NewPipelineRepository(db)
	filterUC := usecase.NewFilterUsecase(filterRepo, pipelineFilterRepo, pipelineRepo)
	filterHandler := handler.NewFilterHandler(filterUC)

	statsRepo := repository.NewStatisticsRepository()
	pipelineUC := usecase.NewPipelineUsecase(db, pipelineRepo, pipelineFilterRepo, filterRepo)
	pipelineHandler := handler.NewPipelineHandler(pipelineUC)
	pipelineScheduler := usecase.NewPipelineScheduler(pipelineRepo, filterRepo, statsRepo)

	filterDefSync := usecase.NewFilterDefinitionSyncService(filterRepo)

	return &Module{
		filterGroupHandler: fgHandler,
		filterHandler:      filterHandler,
		pipelineHandler:    pipelineHandler,
		pipelineScheduler:  pipelineScheduler,
		filterDefSync:      filterDefSync,
		auditLogger:        auditLogger,
	}
}

func (m *Module) GetFilterGroupHandler() *handler.FilterGroupHandler {
	return m.filterGroupHandler
}

func (m *Module) GetFilterHandler() *handler.FilterHandler {
	return m.filterHandler
}

func (m *Module) GetPipelineHandler() *handler.PipelineHandler {
	return m.pipelineHandler
}

func (m *Module) Start(ctx context.Context) {
	go m.filterDefSync.Sync(ctx)
	go m.pipelineScheduler.Start(ctx)
}
