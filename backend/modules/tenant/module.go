package tenant

import (
	"gorm.io/gorm"

	"github.com/utmstack/utmstack/backend/modules/tenant/connectors"
	"github.com/utmstack/utmstack/backend/modules/tenant/handler"
	"github.com/utmstack/utmstack/backend/modules/tenant/repository"
	"github.com/utmstack/utmstack/backend/modules/tenant/usecase"
)

type Module struct {
	tenantHandler *handler.TenantHandler
	tenantUC      connectors.TenantUsecase
	bootstrapUC   connectors.BootstrapUsecase
}

func NewModule(db *gorm.DB, admin connectors.UserProvisioner, extras ...connectors.TenantPurgeFunc) *Module {
	repo := repository.NewTenantRepository(db)
	tenantUC := usecase.NewTenantUsecase(repo, admin, extras)

	return &Module{
		tenantHandler: handler.NewTenantHandler(tenantUC),
		tenantUC:      tenantUC,
		bootstrapUC:   usecase.NewBootstrapUsecase(repo, admin),
	}
}

func (m *Module) GetTenantHandler() *handler.TenantHandler         { return m.tenantHandler }
func (m *Module) GetTenantUsecase() connectors.TenantUsecase       { return m.tenantUC }
func (m *Module) GetBootstrapUsecase() connectors.BootstrapUsecase { return m.bootstrapUC }
