package integrations

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"
	ds_connectors "github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/repository"
	"github.com/utmstack/utmstack/backend/modules/integrations/usecase"
	"github.com/utmstack/utmstack/backend/modules/integrations/verifier"
	"gorm.io/gorm"
)

type Module struct {
	db        *gorm.DB
	tenants   *usecase.TenantUsecase
	modules   connectors.ModuleUsecase
	bootstrap *usecase.Bootstrap
}

func NewModule(db *gorm.DB, cipher connectors.Cipher, tenantDir string, datasources ds_connectors.DatasourceUsecase) *Module {
	store := repository.NewTenantStore(tenantDir)
	schema := repository.NewCodeSchemaProvider() // field schema lives in code
	verif := verifier.NewBackendVerifier()

	moduleRepo := repository.NewModuleRepository(db)
	// store doubles as the tenant-file toggler: enable/disable <module>.yaml(.disabled).
	moduleUC := usecase.NewModuleUsecase(moduleRepo, store)

	return &Module{
		db:        db,
		tenants:   usecase.NewTenantUsecase(store, schema, verif, cipher, moduleRepo, datasources),
		modules:   moduleUC,
		bootstrap: usecase.NewBootstrap(db, store),
	}
}

func (m *Module) Start(ctx context.Context) {
	if err := m.bootstrap.Run(ctx); err != nil {
		_ = catcher.Error("integrations bootstrap failed", err, nil)
	}

	if err := m.tenants.SyncDatasources(ctx); err != nil {
		_ = catcher.Error("integrations: datasource sync failed", err, nil)
	}
}

func (m *Module) Tenants() *usecase.TenantUsecase { return m.tenants }

func (m *Module) Modules() connectors.ModuleUsecase { return m.modules }
