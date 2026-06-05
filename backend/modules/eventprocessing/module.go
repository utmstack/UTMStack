package correlation

import (
	"context"

	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	"github.com/utmstack/utmstack/backend/modules/correlation/connectors"
	"github.com/utmstack/utmstack/backend/modules/correlation/handler"
	"github.com/utmstack/utmstack/backend/modules/correlation/repository"
	"github.com/utmstack/utmstack/backend/modules/correlation/usecase"
	"github.com/utmstack/utmstack/backend/pkg/env"
	"gorm.io/gorm"
)

type Module struct {
	regexPatternHandler    *handler.RegexPatternHandler
	tenantConfigHandler    *handler.TenantConfigHandler
	dataTypeHandler        *handler.DataTypeHandler
	correlationRuleHandler *handler.CorrelationRuleHandler

	regexPatternUsecase    connectors.RegexPatternUsecase
	tenantConfigUsecase    connectors.TenantConfigUsecase
	dataTypeUsecase        connectors.DataTypeUsecase
	correlationRuleUsecase connectors.CorrelationRuleUsecase

	dataTypeScheduler     *usecase.DataTypeScheduler
	definitionSyncService *usecase.DefinitionSyncService
}

func NewModule(db *gorm.DB, auditLogger audit_connectors.Logger, dataInputReader connectors.DataInputStatusReader) *Module {
	// assetSyncer: replaced when module #33 is ported.
	assetSyncer := connectors.NewNoopAssetSyncer()

	// Repositories
	regexPatternRepo := repository.NewRegexPatternRepository(db)
	tenantConfigRepo := repository.NewTenantConfigRepository(db)
	dataTypeRepo := repository.NewDataTypeRepository(db)
	correlationRuleRepo := repository.NewCorrelationRuleRepository(db)

	// Usecases
	regexPatternUC := usecase.NewRegexPatternUsecase(regexPatternRepo)
	tenantConfigUC := usecase.NewTenantConfigUsecase(tenantConfigRepo)
	dataTypeUC := usecase.NewDataTypeUsecase(dataTypeRepo, assetSyncer)
	correlationRuleUC := usecase.NewCorrelationRuleUsecase(correlationRuleRepo)

	// Scheduler
	dataTypeSched := usecase.NewDataTypeScheduler(dataTypeRepo, dataInputReader)

	// DefinitionSyncService — reads YAML rules from disk on startup.
	// rulesDir defaults to "./utmstack/rules"; override with CORRELATION_RULES_DIR.
	rulesDir := env.String("CORRELATION_RULES_DIR", "./utmstack/rules", false)
	defSyncSvc := usecase.NewDefinitionSyncService(correlationRuleRepo, dataTypeRepo, rulesDir)

	// Handlers
	regexPatternH := handler.NewRegexPatternHandler(regexPatternUC)
	tenantConfigH := handler.NewTenantConfigHandler(tenantConfigUC)
	dataTypeH := handler.NewDataTypeHandler(dataTypeUC)
	correlationRuleH := handler.NewCorrelationRuleHandler(correlationRuleUC)

	_ = auditLogger // used by routes.go, stored here for future use

	return &Module{
		regexPatternHandler:    regexPatternH,
		tenantConfigHandler:    tenantConfigH,
		dataTypeHandler:        dataTypeH,
		correlationRuleHandler: correlationRuleH,
		regexPatternUsecase:    regexPatternUC,
		tenantConfigUsecase:    tenantConfigUC,
		dataTypeUsecase:        dataTypeUC,
		correlationRuleUsecase: correlationRuleUC,
		dataTypeScheduler:      dataTypeSched,
		definitionSyncService:  defSyncSvc,
	}
}

func (m *Module) Start(ctx context.Context) {
	// Run the definition sync once at startup (non-blocking).
	go m.definitionSyncService.Sync(ctx)

	// 60s data-type scheduler.
	go m.dataTypeScheduler.Start(ctx)
}

// Getters

func (m *Module) GetRegexPatternHandler() *handler.RegexPatternHandler { return m.regexPatternHandler }
func (m *Module) GetTenantConfigHandler() *handler.TenantConfigHandler { return m.tenantConfigHandler }
func (m *Module) GetDataTypeHandler() *handler.DataTypeHandler         { return m.dataTypeHandler }
func (m *Module) GetCorrelationRuleHandler() *handler.CorrelationRuleHandler {
	return m.correlationRuleHandler
}

func (m *Module) GetRegexPatternUsecase() connectors.RegexPatternUsecase {
	return m.regexPatternUsecase
}
func (m *Module) GetTenantConfigUsecase() connectors.TenantConfigUsecase {
	return m.tenantConfigUsecase
}
func (m *Module) GetDataTypeUsecase() connectors.DataTypeUsecase {
	return m.dataTypeUsecase
}
func (m *Module) GetCorrelationRuleUsecase() connectors.CorrelationRuleUsecase {
	return m.correlationRuleUsecase
}
