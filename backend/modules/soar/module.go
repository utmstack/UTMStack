package soar

import (
	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/handler"
	"github.com/utmstack/utmstack/backend/modules/soar/repository"
	"github.com/utmstack/utmstack/backend/modules/soar/usecase"

	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	jwtpkg "github.com/utmstack/utmstack/backend/pkg/jwt"
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

	ruleUsecase      connectors.RuleUsecase
	templateUsecase  connectors.TemplateUsecase
	executionUsecase connectors.ExecutionUsecase

	commandWSHandler *handler.CommandWSHandler
}

func NewModule(
	db *gorm.DB,
	agentClient *agentmanager.AgentManagerClient,
	signer *jwtpkg.Signer,
) *Module {
	// Rule control-plane.
	ruleRepo := repository.NewRuleRepository(db)
	templateRepo := repository.NewTemplateRepository(db)
	resolveRepo := repository.NewResolveFilterRepository(db)
	executionRepo := repository.NewExecutionRepository(db)

	ruleUC := usecase.NewRuleUsecase(ruleRepo, resolveRepo)
	templateUC := usecase.NewTemplateUsecase(templateRepo)
	executionUC := usecase.NewExecutionUsecase(executionRepo)

	return &Module{
		ruleHandler:      handler.NewRuleHandler(ruleUC),
		templateHandler:  handler.NewTemplateHandler(templateUC),
		executionHandler: handler.NewExecutionHandler(executionUC),
		ruleUsecase:      ruleUC,
		templateUsecase:  templateUC,
		executionUsecase: executionUC,

		commandWSHandler: handler.NewCommandWSHandler(agentClient, signer),
	}
}

func (m *Module) GetRuleHandler() *handler.RuleHandler             { return m.ruleHandler }
func (m *Module) GetTemplateHandler() *handler.TemplateHandler     { return m.templateHandler }
func (m *Module) GetExecutionHandler() *handler.ExecutionHandler   { return m.executionHandler }
func (m *Module) GetRuleUsecase() connectors.RuleUsecase           { return m.ruleUsecase }
func (m *Module) GetTemplateUsecase() connectors.TemplateUsecase   { return m.templateUsecase }
func (m *Module) GetExecutionUsecase() connectors.ExecutionUsecase { return m.executionUsecase }
