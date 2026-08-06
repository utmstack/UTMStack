package integrations

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"
	ds_connectors "github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/repository"
	"github.com/utmstack/utmstack/backend/modules/integrations/usecase"
	"github.com/utmstack/utmstack/backend/modules/integrations/verifier"
	os_connectors "github.com/utmstack/utmstack/backend/modules/opensearch/connectors"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	"gorm.io/gorm"
)

type Module struct {
	db         *gorm.DB
	tenants    *usecase.TenantUsecase
	modules    connectors.ModuleUsecase
	collectors connectors.CollectorUsecase
	opensearch os_connectors.IndexPatternUsecase
}

func NewModule(
	db *gorm.DB,
	cipher connectors.Cipher,
	tenantDir string,
	datasources ds_connectors.DatasourceUsecase,
	opensearch os_connectors.IndexPatternUsecase,
	agentClient *agentmanager.AgentManagerClient,
) *Module {
	store := repository.NewTenantStore(tenantDir)
	schema := repository.NewCodeSchemaProvider() // field schema lives in code
	verif := verifier.NewBackendVerifier()

	moduleRepo := repository.NewModuleRepository(db)
	// store doubles as the tenant-file toggler: enable/disable <module>.yaml(.disabled).
	moduleUC := usecase.NewModuleUsecase(moduleRepo, store, opensearch)

	var collectorClient connectors.AgentManagerCollectorClient
	if agentClient != nil {
		collectorClient = agentClient
	}
	collectorUC := usecase.NewCollectorUsecase(collectorClient)

	return &Module{
		db:         db,
		tenants:    usecase.NewTenantUsecase(store, schema, verif, cipher, moduleRepo, datasources),
		modules:    moduleUC,
		collectors: collectorUC,
	}
}

func (m *Module) Start(ctx context.Context) {

	if err := m.tenants.SyncDatasources(ctx); err != nil {
		_ = catcher.Error("integrations: datasource sync failed", err, nil)
	}
}

func (m *Module) Tenants() *usecase.TenantUsecase { return m.tenants }

func (m *Module) Modules() connectors.ModuleUsecase { return m.modules }

func (m *Module) Collectors() connectors.CollectorUsecase { return m.collectors }
