package eventprocessing

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/threatwinds/go-sdk/catcher"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/handler"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/infrastructure"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/repository"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/usecase"
	"github.com/utmstack/utmstack/backend/pkg/env"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
	"gorm.io/gorm"
)

type Module struct {
	regexPatternHandler    *handler.RegexPatternHandler
	correlationRuleHandler *handler.CorrelationRuleHandler

	regexPatternUsecase    connectors.RegexPatternUsecase
	assetProjectionUsecase connectors.AssetProjectionUsecase
	correlationRuleUsecase connectors.CorrelationRuleUsecase
	pipelineUsecase        connectors.PipelineUsecase
	ingestionStatsUsecase  connectors.IngestionStatsUsecase

	ruleStore             *repository.RuleStore
	ruleBootstrap         *repository.RuleBootstrap
	pipelineStore         *repository.PipelineStore
	pipelineBootstrap     *repository.PipelineBootstrap
	engineConfigBootstrap *repository.EngineConfigBootstrap

	pipelineHandler       *handler.PipelineHandler
	ingestionStatsHandler *handler.IngestionStatsHandler
	playgroundHandler     *handler.PlaygroundHandler
	playgroundUsecase     connectors.PlaygroundUsecase
}

func NewModule(db *gorm.DB, events *eventstore.Store, auditLogger audit_connectors.Logger, playgroundBaseURL, internalKey string) *Module {
	pipelineDir := env.String(repository.PipelineDirEnv, repository.DefaultPipelineDir, false)
	engineConfig := repository.NewEngineConfig(pipelineDir)
	engineConfigBootstrap := repository.NewEngineConfigBootstrap(engineConfig)

	// Rule store — YAML-direct overlays.
	rulesRoot := env.String(repository.RulesDirEnv, repository.DefaultRulesDir, false)
	ruleStore := repository.NewRuleStore(
		filepath.Join(rulesRoot, repository.SystemSubdir),
		filepath.Join(rulesRoot, repository.UserSubdir),
		engineConfig,
	)
	_ = ruleStore.Load()
	ruleBootstrap := repository.NewRuleBootstrap(env.String(repository.RulesSrcDirEnv, repository.DefaultSystemRulesSrcDir, false), ruleStore, db)

	// Filter store — YAML-direct overlays (pipeline: format).
	filtersRoot := env.String(repository.PipelinesDirEnv, repository.DefaultPipelinesDir, false)
	pipelineStore := repository.NewPipelineStore(
		filepath.Join(filtersRoot, repository.SystemSubdir),
		filepath.Join(filtersRoot, repository.UserSubdir),
		engineConfig,
	)
	_ = pipelineStore.Load()
	pipelineBootstrap := repository.NewPipelineBootstrap(env.String(repository.PipelinesSrcDirEnv, repository.DefaultSystemPipelinesSrcDir, false), pipelineStore, db)

	regexPatternUC := usecase.NewRegexPatternUsecase(engineConfig)
	assetProjectionUC := usecase.NewAssetProjection(engineConfig)
	correlationRuleUC := usecase.NewCorrelationRuleUsecase(ruleStore)
	pipelineUC := usecase.NewPipelineUsecase(pipelineStore, engineConfig)

	ingestionStatsUC := usecase.NewIngestionStatsUsecase(repository.NewIngestionStatsRepository(events))

	_ = auditLogger // used by routes.go

	playgroundClient := infrastructure.NewPlaygroundClient(playgroundBaseURL, internalKey)
	playgroundUC := usecase.NewPlaygroundUsecase(playgroundClient)
	playgroundH := handler.NewPlaygroundHandler(playgroundUC)

	return &Module{
		regexPatternHandler:    handler.NewRegexPatternHandler(regexPatternUC),
		correlationRuleHandler: handler.NewCorrelationRuleHandler(correlationRuleUC),
		regexPatternUsecase:    regexPatternUC,
		assetProjectionUsecase: assetProjectionUC,
		correlationRuleUsecase: correlationRuleUC,
		pipelineUsecase:        pipelineUC,
		ingestionStatsUsecase:  ingestionStatsUC,
		ruleStore:              ruleStore,
		ruleBootstrap:          ruleBootstrap,
		pipelineStore:          pipelineStore,
		pipelineBootstrap:      pipelineBootstrap,
		engineConfigBootstrap:  engineConfigBootstrap,
		pipelineHandler:       handler.NewPipelineHandler(pipelineUC),
		ingestionStatsHandler: handler.NewIngestionStatsHandler(ingestionStatsUC),
		playgroundHandler:      playgroundH,
		playgroundUsecase:      playgroundUC,
	}
}

func (m *Module) Start(ctx context.Context) {
	if err := m.ruleBootstrap.Run(ctx); err != nil {
		_ = catcher.Error("rule bootstrap failed", err, nil)
	}
	if err := m.pipelineBootstrap.Run(ctx); err != nil {
		_ = catcher.Error("pipeline bootstrap failed", err, nil)
	}
	if err := m.engineConfigBootstrap.Run(ctx); err != nil {
		_ = catcher.Error("engine config bootstrap (tenants/patterns) failed", err, nil)
	}

	go m.ruleStore.Watch(ctx)
	go m.pipelineStore.Watch(ctx)
}

func (m *Module) GetRegexPatternHandler() *handler.RegexPatternHandler { return m.regexPatternHandler }
func (m *Module) GetCorrelationRuleHandler() *handler.CorrelationRuleHandler {
	return m.correlationRuleHandler
}
func (m *Module) GetPipelineHandler() *handler.PipelineHandler { return m.pipelineHandler }
func (m *Module) GetIngestionStatsHandler() *handler.IngestionStatsHandler {
	return m.ingestionStatsHandler
}

func (m *Module) GetRegexPatternUsecase() connectors.RegexPatternUsecase {
	return m.regexPatternUsecase
}
func (m *Module) GetAssetProjectionUsecase() connectors.AssetProjectionUsecase {
	return m.assetProjectionUsecase
}
func (m *Module) GetCorrelationRuleUsecase() connectors.CorrelationRuleUsecase {
	return m.correlationRuleUsecase
}
func (m *Module) GetPipelineUsecase() connectors.PipelineUsecase { return m.pipelineUsecase }
func (m *Module) GetIngestionStatsUsecase() connectors.IngestionStatsUsecase {
	return m.ingestionStatsUsecase
}

func (m *Module) GetPlaygroundHandler() *handler.PlaygroundHandler {
	return m.playgroundHandler
}
func (m *Module) GetPlaygroundUsecase() connectors.PlaygroundUsecase {
	return m.playgroundUsecase
}

func (m *Module) AfterEventsByRuleName(name string) (json.RawMessage, bool) {
	return m.ruleStore.CorrelationsByName(name)
}
