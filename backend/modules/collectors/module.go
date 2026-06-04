package collectors

import (
	"github.com/utmstack/utmstack/backend/modules/collectors/handler"
	"github.com/utmstack/utmstack/backend/modules/collectors/repository"
	"github.com/utmstack/utmstack/backend/modules/collectors/usecase"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	"gorm.io/gorm"
)

type Module struct {
	collectorHandler *handler.CollectorHandler
	agentHandler     *handler.AgentHandler
}

func NewModule(db *gorm.DB, agentClient *agentmanager.AgentManagerClient) *Module {
	collectorRepo := repository.NewCollectorPGRepository(db)
	agentRepo := repository.NewAgentManagerRepository(agentClient)

	collectorUC := usecase.NewCollectorUsecase(collectorRepo, agentRepo)
	agentUC := usecase.NewAgentUsecase(agentRepo)

	return &Module{
		collectorHandler: handler.NewCollectorHandler(collectorUC),
		agentHandler:     handler.NewAgentHandler(agentUC),
	}
}
