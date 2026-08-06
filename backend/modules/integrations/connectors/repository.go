package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type TenantRepository interface {
	Load(module string) ([]domain.Tenant, error)
	Save(module string, tenants []domain.Tenant) error
	Upsert(module string, t domain.Tenant) error
	Delete(module, name, tenantId string) error
	SetActiveByModule(ctx context.Context, module string, active bool) error
}

type ModuleRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.UtmModule, error)
	GetByName(ctx context.Context, moduleName string) (*domain.UtmModule, error)
	List(ctx context.Context, filter ModuleListFilter) ([]domain.UtmModule, int64, error)
	Save(ctx context.Context, module *domain.UtmModule) error
	Delete(ctx context.Context, id int64) error
	Categories(ctx context.Context) ([]string, error)
	DataTypes(ctx context.Context) ([]domain.UtmModule, error)
	CountActiveByName(ctx context.Context, moduleName string) (int64, error)
}

type ModuleListFilter struct {
	ModuleCategory *string
	ModuleActive   *bool
	NameContains   *string
	database.Params
}

type SchemaProvider interface {
	Schema(module string) (map[string]string, error)
}
