package eventprocessing

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/threatwinds/go-sdk/catcher"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/handler"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/repository"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/usecase"
	"github.com/utmstack/utmstack/backend/pkg/env"
	"gorm.io/gorm"
)

type Module struct {
	regexPatternHandler    *handler.RegexPatternHandler
	tenantConfigHandler    *handler.TenantConfigHandler
	correlationRuleHandler *handler.CorrelationRuleHandler

	regexPatternUsecase    connectors.RegexPatternUsecase
	tenantConfigUsecase    connectors.TenantConfigUsecase
	correlationRuleUsecase connectors.CorrelationRuleUsecase
	filterUsecase          connectors.FilterUsecase
	ingestionStatsUsecase  connectors.IngestionStatsUsecase

	ruleStore         *usecase.RuleStore
	ruleBootstrap     *usecase.RuleBootstrap
	filterStore       *usecase.FilterStore
	filterBootstrap   *usecase.FilterBootstrap
	pipelineBootstrap *usecase.PipelineBootstrap

	filterHandler         *handler.FilterHandler
	ingestionStatsHandler *handler.IngestionStatsHandler
}

func NewModule(db *gorm.DB, auditLogger audit_connectors.Logger) *Module {
	// Pipeline writer — writes tenants.yaml and patterns.yaml on every mutation
	// and on bootstrap (migration from DB).
	pipelineDir := env.String(usecase.PipelineDirEnv, usecase.DefaultPipelineDir, false)
	pipelineWriter := usecase.NewPipelineWriter(pipelineDir)
	pipelineBootstrap := usecase.NewPipelineBootstrap(pipelineWriter, db)

	// Rule store — YAML-direct overlays.
	rulesRoot := env.String(usecase.RulesDirEnv, usecase.DefaultRulesDir, false)
	ruleStore := usecase.NewRuleStore(
		filepath.Join(rulesRoot, usecase.SystemSubdir),
		filepath.Join(rulesRoot, usecase.UserSubdir),
		pipelineWriter,
	)
	_ = ruleStore.Load()
	ruleBootstrap := usecase.NewRuleBootstrap(usecase.SystemRulesSrcDir, ruleStore, db)

	// Filter store — YAML-direct overlays (pipeline: format).
	filtersRoot := env.String(usecase.FiltersDirEnv, usecase.DefaultFiltersDir, false)
	filterStore := usecase.NewFilterStore(
		filepath.Join(filtersRoot, usecase.SystemSubdir),
		filepath.Join(filtersRoot, usecase.UserSubdir),
	)
	_ = filterStore.Load()
	filterBootstrap := usecase.NewFilterBootstrap(usecase.SystemFiltersSrcDir, filterStore, db)

	regexPatternUC := usecase.NewRegexPatternUsecase(pipelineWriter)
	tenantConfigUC := usecase.NewTenantConfigUsecase(pipelineWriter)
	correlationRuleUC := usecase.NewCorrelationRuleUsecase(ruleStore)
	filterUC := usecase.NewFilterUsecase(filterStore)

	ingestionStatsUC := usecase.NewIngestionStatsUsecase(repository.NewIngestionStatsRepository())

	_ = auditLogger // used by routes.go

	return &Module{
		regexPatternHandler:    handler.NewRegexPatternHandler(regexPatternUC),
		tenantConfigHandler:    handler.NewTenantConfigHandler(tenantConfigUC),
		correlationRuleHandler: handler.NewCorrelationRuleHandler(correlationRuleUC),
		regexPatternUsecase:    regexPatternUC,
		tenantConfigUsecase:    tenantConfigUC,
		correlationRuleUsecase: correlationRuleUC,
		filterUsecase:          filterUC,
		ingestionStatsUsecase:  ingestionStatsUC,
		ruleStore:              ruleStore,
		ruleBootstrap:          ruleBootstrap,
		filterStore:            filterStore,
		filterBootstrap:        filterBootstrap,
		pipelineBootstrap:      pipelineBootstrap,
		filterHandler:          handler.NewFilterHandler(filterUC),
		ingestionStatsHandler:  handler.NewIngestionStatsHandler(ingestionStatsUC),
	}
}

func (m *Module) Start(ctx context.Context) {
	if err := m.ruleBootstrap.Run(ctx); err != nil {
		_ = catcher.Error("rule bootstrap failed", err, nil)
	}
	if err := m.filterBootstrap.Run(ctx); err != nil {
		_ = catcher.Error("filter bootstrap failed", err, nil)
	}
	if err := m.pipelineBootstrap.Run(ctx); err != nil {
		_ = catcher.Error("pipeline bootstrap (tenant/patterns) failed", err, nil)
	}
}

func (m *Module) GetRegexPatternHandler() *handler.RegexPatternHandler { return m.regexPatternHandler }
func (m *Module) GetTenantConfigHandler() *handler.TenantConfigHandler { return m.tenantConfigHandler }
func (m *Module) GetCorrelationRuleHandler() *handler.CorrelationRuleHandler {
	return m.correlationRuleHandler
}
func (m *Module) GetFilterHandler() *handler.FilterHandler { return m.filterHandler }
func (m *Module) GetIngestionStatsHandler() *handler.IngestionStatsHandler {
	return m.ingestionStatsHandler
}

func (m *Module) GetRegexPatternUsecase() connectors.RegexPatternUsecase {
	return m.regexPatternUsecase
}
func (m *Module) GetTenantConfigUsecase() connectors.TenantConfigUsecase {
	return m.tenantConfigUsecase
}
func (m *Module) GetCorrelationRuleUsecase() connectors.CorrelationRuleUsecase {
	return m.correlationRuleUsecase
}
func (m *Module) GetFilterUsecase() connectors.FilterUsecase { return m.filterUsecase }
func (m *Module) GetIngestionStatsUsecase() connectors.IngestionStatsUsecase {
	return m.ingestionStatsUsecase
}

func (m *Module) AfterEventsByRuleName(name string) (json.RawMessage, bool) {
	return m.ruleStore.CorrelationsByName(name)
}
