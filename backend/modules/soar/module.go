package soar

import (
	"context"
	"path/filepath"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/handler"
	"github.com/utmstack/utmstack/backend/modules/soar/repository"
	"github.com/utmstack/utmstack/backend/modules/soar/usecase"

	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	"github.com/utmstack/utmstack/backend/pkg/env"
	jwtpkg "github.com/utmstack/utmstack/backend/pkg/jwt"
	"github.com/utmstack/utmstack/backend/pkg/secret"
	"gorm.io/gorm"
)

// Module is the SOAR plane. It owns:
//   - the rule/template control-plane (CRUD + execution history); real-time rule
//     evaluation lives in the active-response plugin.
//   - the live command WebSocket that runs commands on agents via agent-manager.
type Module struct {
	ruleHandler      *handler.RuleHandler
	templateHandler  *handler.TemplateHandler
	executionHandler *handler.ExecutionHandler
	agentHandler     *handler.AgentHandler

	ruleUsecase          connectors.RuleUsecase
	templateUsecase      connectors.TemplateUsecase
	executionUsecase     connectors.ExecutionUsecase
	agentUsecase         connectors.AgentUsecase
	variableUsecase      connectors.VariableUsecase
	actionUsecase        connectors.ActionUsecase
	actionCommandUsecase connectors.ActionCommandUsecase
	jobUsecase           connectors.JobUsecase

	variableHandler      *handler.VariableHandler
	actionHandler        *handler.ActionHandler
	actionCommandHandler *handler.ActionCommandHandler
	jobHandler           *handler.JobHandler
	commandWSHandler     *handler.CommandWSHandler

	flowBootstrap *usecase.FlowBootstrap
}

func NewModule(
	db *gorm.DB,
	agentClient *agentmanager.AgentManagerClient,
	signer *jwtpkg.Signer,
	cipher *secret.Cipher,
) *Module {
	flowsRoot := env.String("SOAR_FLOWS_DIR", "/workdir/soar", false)
	flowsSrc := env.String("SOAR_FLOWS_SRC_DIR", "/utmstack/soar", false)
	flowStore := usecase.NewFlowStore(filepath.Join(flowsRoot, usecase.SystemSubdir), filepath.Join(flowsRoot, usecase.UserSubdir))
	flowBootstrap := usecase.NewFlowBootstrap(flowsSrc, flowStore, db)

	templateRepo := repository.NewTemplateRepository(db)
	resolveRepo := repository.NewResolveFilterRepository(db)
	executionRepo := repository.NewExecutionRepository(db)

	ruleUC := usecase.NewRuleUsecase(flowStore, resolveRepo)
	templateUC := usecase.NewTemplateUsecase(templateRepo)
	executionUC := usecase.NewExecutionUsecase(executionRepo)

	// Agent resolution: cross-table query into `datasources` for rule targeting.
	// Internal-only consumer (SOAR plugin).
	agentRepo := repository.NewAgentRepository(db)
	agentUC := usecase.NewAgentUsecase(agentRepo)

	// Incident variables: reusable (optionally secret) values interpolated into
	// commands run on agents through the live command WebSocket.
	variableRepo := repository.NewVariableRepository(db)
	variableUC := usecase.NewVariableUsecase(variableRepo, connectors.NewSecretCipherAdapter(cipher))

	// Incident-response automation: predefined actions, their per-OS commands,
	// and the jobs (responses) run against agents.
	actionUC := usecase.NewActionUsecase(repository.NewActionRepository(db))
	actionCommandUC := usecase.NewActionCommandUsecase(repository.NewActionCommandRepository(db))
	jobUC := usecase.NewJobUsecase(repository.NewJobRepository(db))

	return &Module{
		ruleHandler:      handler.NewRuleHandler(ruleUC),
		templateHandler:  handler.NewTemplateHandler(templateUC),
		executionHandler: handler.NewExecutionHandler(executionUC),
		agentHandler:     handler.NewAgentHandler(agentUC),
		ruleUsecase:      ruleUC,
		templateUsecase:  templateUC,
		executionUsecase: executionUC,
		agentUsecase:     agentUC,
		variableUsecase:  variableUC,
		actionUsecase:    actionUC,

		actionCommandUsecase: actionCommandUC,
		jobUsecase:           jobUC,

		variableHandler:      handler.NewVariableHandler(variableUC),
		actionHandler:        handler.NewActionHandler(actionUC),
		actionCommandHandler: handler.NewActionCommandHandler(actionCommandUC),
		jobHandler:           handler.NewJobHandler(jobUC),
		commandWSHandler:     handler.NewCommandWSHandler(agentClient, signer, variableUC),

		flowBootstrap: flowBootstrap,
	}
}

func (m *Module) Start(ctx context.Context) error {
	return m.flowBootstrap.Run(ctx)
}

func (m *Module) GetRuleHandler() *handler.RuleHandler           { return m.ruleHandler }
func (m *Module) GetTemplateHandler() *handler.TemplateHandler   { return m.templateHandler }
func (m *Module) GetExecutionHandler() *handler.ExecutionHandler { return m.executionHandler }
func (m *Module) GetAgentHandler() *handler.AgentHandler         { return m.agentHandler }
func (m *Module) GetVariableHandler() *handler.VariableHandler   { return m.variableHandler }
func (m *Module) GetActionHandler() *handler.ActionHandler       { return m.actionHandler }
func (m *Module) GetActionCommandHandler() *handler.ActionCommandHandler {
	return m.actionCommandHandler
}
func (m *Module) GetJobHandler() *handler.JobHandler               { return m.jobHandler }
func (m *Module) GetRuleUsecase() connectors.RuleUsecase           { return m.ruleUsecase }
func (m *Module) GetTemplateUsecase() connectors.TemplateUsecase   { return m.templateUsecase }
func (m *Module) GetExecutionUsecase() connectors.ExecutionUsecase { return m.executionUsecase }
