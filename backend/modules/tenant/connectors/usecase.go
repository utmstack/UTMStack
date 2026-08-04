package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/tenant/domain"
	"github.com/utmstack/utmstack/backend/modules/tenant/dto"
)

type AdminProvisioner interface {
	CreateTenantAdmin(ctx context.Context, tenantID, login, email string) error
}

type TenantUsecase interface {
	Create(ctx context.Context, req dto.CreateRequest) (*domain.Tenant, error)
	Update(ctx context.Context, id string, req dto.UpdateRequest) (*domain.Tenant, error)
	GetByID(ctx context.Context, id string) (*domain.Tenant, error)
	List(ctx context.Context, f dto.Filter) ([]domain.Tenant, int64, error)
	Terminate(ctx context.Context, id string) error
	ResolveDomain(ctx context.Context, host string) (*domain.Tenant, error)
}
