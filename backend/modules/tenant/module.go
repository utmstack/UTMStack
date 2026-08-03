package tenant

import (
	"gorm.io/gorm"

	"github.com/utmstack/utmstack/backend/modules/tenant/connectors"
	"github.com/utmstack/utmstack/backend/modules/tenant/handler"
	"github.com/utmstack/utmstack/backend/modules/tenant/repository"
	"github.com/utmstack/utmstack/backend/modules/tenant/usecase"
)

// TODO(multi-tenant): the platform administrator is not created here.
//
// It is the operator's own account: it belongs to no tenant, so tenant scoping
// leaves it unable to read any customer's data — which is the posture we want
// by default. It is created when the instance is deployed, not when a customer
// subscribes, so it belongs to bootstrap rather than to this module.
//
// What it still needs: the ROLE_PLATFORM_ADMIN role, a RequirePlatform gate on
// everything instance-wide, and a break-glass path with an audit trail for the
// times support genuinely has to enter a tenant.

type Module struct {
	tenantHandler *handler.TenantHandler
	tenantUC      connectors.TenantUsecase
}

func NewModule(db *gorm.DB) *Module {
	tenantUC := usecase.NewTenantUsecase(repository.NewTenantRepository(db))

	return &Module{
		tenantHandler: handler.NewTenantHandler(tenantUC),
		tenantUC:      tenantUC,
	}
}

func (m *Module) GetTenantHandler() *handler.TenantHandler   { return m.tenantHandler }
func (m *Module) GetTenantUsecase() connectors.TenantUsecase { return m.tenantUC }
