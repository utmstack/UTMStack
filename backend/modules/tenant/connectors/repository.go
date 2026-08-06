package connectors

import (
	"context"
	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/tenant/domain"
	"github.com/utmstack/utmstack/backend/modules/tenant/dto"
)

type TenantRepository interface {
	Create(ctx context.Context, t *domain.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, t *domain.Tenant) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	FindByDomain(ctx context.Context, domain string) (*domain.Tenant, error)
	List(ctx context.Context, f dto.Filter) ([]domain.Tenant, int64, error)
}
