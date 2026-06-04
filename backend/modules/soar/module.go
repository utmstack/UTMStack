package soar

import (
	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/handler"
	"github.com/utmstack/utmstack/backend/modules/soar/repository"
	"github.com/utmstack/utmstack/backend/modules/soar/usecase"
	"gorm.io/gorm"
)

// Module is the SOAR control-plane: it owns rule/template CRUD and the execution
// history. Real-time rule evaluation lives in the active-response plugin and
// command execution in the agent-manager, so this module no longer evaluates,
// dispatches, or schedules anything.
type Module struct {
	ruleHandler      *handler.RuleHandler
	templateHandler  *handler.TemplateHandler
	executionHandler *handler.ExecutionHandler

	ruleUsecase      connectors.RuleUsecase
	templateUsecase  connectors.TemplateUsecase
	executionUsecase connectors.ExecutionUsecase
}

func NewModule(db *gorm.DB) *Module {
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
	}
}

func (m *Module) GetRuleHandler() *handler.RuleHandler             { return m.ruleHandler }
func (m *Module) GetTemplateHandler() *handler.TemplateHandler     { return m.templateHandler }
func (m *Module) GetExecutionHandler() *handler.ExecutionHandler   { return m.executionHandler }
func (m *Module) GetRuleUsecase() connectors.RuleUsecase           { return m.ruleUsecase }
func (m *Module) GetTemplateUsecase() connectors.TemplateUsecase   { return m.templateUsecase }
func (m *Module) GetExecutionUsecase() connectors.ExecutionUsecase { return m.executionUsecase }
