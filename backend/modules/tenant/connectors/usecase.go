package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/tenant/domain"
	"github.com/utmstack/utmstack/backend/modules/tenant/dto"
)

type TenantUsecase interface {
	Create(ctx context.Context, req dto.CreateRequest) (*domain.Tenant, error)
	Update(ctx context.Context, id string, req dto.UpdateRequest) (*domain.Tenant, error)
	GetByID(ctx context.Context, id string) (*domain.Tenant, error)
	List(ctx context.Context, f dto.Filter) ([]domain.Tenant, int64, error)

	// Terminate marks a tenant for deletion. It never destroys the row: the
	// data outlives the subscription, and an operator has to be able to see
	// what was terminated and when.
	Terminate(ctx context.Context, id string) error

	// ResolveDomain maps a request host to its tenant. It answers which tenant
	// is being asked for, never which one the caller belongs to — that is the
	// tid claim in the token.
	ResolveDomain(ctx context.Context, host string) (*domain.Tenant, error)
}
