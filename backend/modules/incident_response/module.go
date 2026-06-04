package incident_response

import (
	"github.com/utmstack/utmstack/backend/modules/incident_response/connectors"
	"github.com/utmstack/utmstack/backend/modules/incident_response/handler"
	"github.com/utmstack/utmstack/backend/modules/incident_response/repository"
	"github.com/utmstack/utmstack/backend/modules/incident_response/usecase"
	incidents_repository "github.com/utmstack/utmstack/backend/modules/incidents/repository"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	jwtpkg "github.com/utmstack/utmstack/backend/pkg/jwt"
	"gorm.io/gorm"
)

type Module struct {
	actionHandler        *handler.ActionHandler
	actionCommandHandler *handler.ActionCommandHandler
	jobHandler           *handler.JobHandler
	variableHandler      *handler.VariableHandler
	commandWSHandler     *handler.CommandWSHandler
}

func NewModule(db *gorm.DB, cipher connectors.VariableCipher, agentClient *agentmanager.AgentManagerClient, signer *jwtpkg.Signer) *Module {
	// Repositories
	actionRepo := repository.NewActionRepository(db)
	actionCmdRepo := repository.NewActionCommandRepository(db)
	jobRepo := repository.NewJobRepository(db)
	variableRepo := repository.NewVariableRepository(db)

	// Incident history writer
	historyRepo := incidents_repository.NewIncidentHistoryRepository(db)
	historyWriter := connectors.NewIncidentHistoryWriterAdapter(historyRepo)

	// Usecases
	actionUC := usecase.NewActionUsecase(actionRepo)
	actionCmdUC := usecase.NewActionCommandUsecase(actionCmdRepo)
	jobUC := usecase.NewJobUsecase(jobRepo, historyWriter)
	variableUC := usecase.NewVariableUsecase(variableRepo, cipher)

	return &Module{
		actionHandler:        handler.NewActionHandler(actionUC),
		actionCommandHandler: handler.NewActionCommandHandler(actionCmdUC),
		jobHandler:           handler.NewJobHandler(jobUC),
		variableHandler:      handler.NewVariableHandler(variableUC),
		commandWSHandler:     handler.NewCommandWSHandler(variableUC, agentClient, signer),
	}
}
