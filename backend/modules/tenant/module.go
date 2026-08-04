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
}

func NewModule(db *gorm.DB, admin connectors.AdminProvisioner) *Module {
	tenantUC := usecase.NewTenantUsecase(repository.NewTenantRepository(db), admin)

	return &Module{
		tenantHandler: handler.NewTenantHandler(tenantUC),
		tenantUC:      tenantUC,
	}
}

func (m *Module) GetTenantHandler() *handler.TenantHandler   { return m.tenantHandler }
func (m *Module) GetTenantUsecase() connectors.TenantUsecase { return m.tenantUC }
