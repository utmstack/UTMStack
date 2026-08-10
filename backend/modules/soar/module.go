package soar

import (
	"context"
	"path/filepath"

	"github.com/threatwinds/go-sdk/catcher"
	"gorm.io/gorm"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/handler"
	"github.com/utmstack/utmstack/backend/modules/soar/repository"
	"github.com/utmstack/utmstack/backend/modules/soar/usecase"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	"github.com/utmstack/utmstack/backend/pkg/env"
	jwtpkg "github.com/utmstack/utmstack/backend/pkg/jwt"
	"github.com/utmstack/utmstack/backend/pkg/secret"
)

type Module struct {
	ruleHandler      *handler.RuleHandler
	executionHandler *handler.ExecutionHandler
	variableHandler  *handler.VariableHandler
	commandWSHandler *handler.CommandWSHandler

	ruleUsecase      connectors.RuleUsecase
	executionUsecase connectors.ExecutionUsecase
	agentUsecase     connectors.AgentUsecase
	variableUsecase  connectors.VariableUsecase

	flowsSrc   string
	flowStore  *usecase.FlowStore
	dispatcher *usecase.Dispatcher
}

func NewModule(
	db *gorm.DB,
	agentClient *agentmanager.AgentManagerClient,
	signer *jwtpkg.Signer,
	cipher *secret.Cipher,
) *Module {
	flowsSrc := env.String("SOAR_FLOWS_SRC_DIR", "/utmstack/soar", false)
	flowsRoot := env.String("SOAR_FLOWS_DIR", "/workdir/soar", false)
	flowStore := usecase.NewFlowStore(
		filepath.Join(flowsRoot, usecase.SystemSubdir),
		filepath.Join(flowsRoot, usecase.UserSubdir),
	)

	resolveRepo := repository.NewResolveFilterRepository(db)
	executionRepo := repository.NewExecutionRepository(db)

	variableRepo := repository.NewVariableRepository(db)
	variableUC := usecase.NewVariableUsecase(variableRepo, cipher)

	dispatcher := usecase.NewDispatcher(executionRepo, flowStore, agentClient, variableUC)

	agentRepo := repository.NewAgentRepository(db)
	agentUC := usecase.NewAgentUsecase(agentRepo)

	ruleUC := usecase.NewRuleUsecase(flowStore, resolveRepo)
	executionUC := usecase.NewExecutionUsecase(executionRepo, flowStore, agentUC, dispatcher.Kick)

	return &Module{
		ruleHandler:      handler.NewRuleHandler(ruleUC),
		executionHandler: handler.NewExecutionHandler(executionUC),
		variableHandler:  handler.NewVariableHandler(variableUC),
		commandWSHandler: handler.NewCommandWSHandler(agentClient, signer, variableUC, executionUC),

		ruleUsecase:      ruleUC,
		executionUsecase: executionUC,
		agentUsecase:     agentUC,
		variableUsecase:  variableUC,

		flowsSrc:   flowsSrc,
		flowStore:  flowStore,
		dispatcher: dispatcher,
	}
}

func (m *Module) Start(ctx context.Context) error {
	if err := m.flowStore.SeedSystem(m.flowsSrc); err != nil {
		_ = catcher.Error("soar: seeding the shipped flows failed", err, nil)
	}
	if err := m.flowStore.Load(); err != nil {
		return err
	}
	go m.flowStore.Watch(ctx)
	go m.dispatcher.Start(ctx)
	return nil
}

func (m *Module) GetRuleHandler() *handler.RuleHandler             { return m.ruleHandler }
func (m *Module) GetExecutionHandler() *handler.ExecutionHandler   { return m.executionHandler }
func (m *Module) GetVariableHandler() *handler.VariableHandler     { return m.variableHandler }
func (m *Module) GetRuleUsecase() connectors.RuleUsecase           { return m.ruleUsecase }
func (m *Module) GetExecutionUsecase() connectors.ExecutionUsecase { return m.executionUsecase }
func (m *Module) GetVariableUsecase() connectors.VariableUsecase   { return m.variableUsecase }
func (m *Module) GetAgentUsecase() connectors.AgentUsecase         { return m.agentUsecase }
