package connectors

import (
	"context"
	"github.com/google/uuid"

	iam_connectors "github.com/utmstack/utmstack/backend/modules/iam/connectors"
	iam_dto "github.com/utmstack/utmstack/backend/modules/iam/dto"
	"github.com/utmstack/utmstack/backend/modules/tenant/domain"
	"github.com/utmstack/utmstack/backend/modules/tenant/dto"
)

// TenantPurgeFunc is a purger contributed by a subsystem that owns
// tenant-scoped data outside PostgreSQL (ClickHouse tables, on-disk config
// directories, etc). Called during PermanentlyDelete before the SQL purge so
// that a failure leaves the tenant row intact and the whole operation stays
// retryable.
type TenantPurgeFunc func(ctx context.Context, id uuid.UUID) error

// UserProvisioner is iam's create, nothing more. Tenant owns tenancy, so it is
// this module that puts the tenant on the context before calling; iam only makes
// the account it is asked for, wherever the caller says it belongs.
type UserProvisioner interface {
	Create(ctx context.Context, input iam_dto.CreateUserRequest, opts iam_connectors.CreateUserOptions) (*iam_dto.UserDetailResponse, error)
}

type BootstrapUsecase interface {
	EnsureDefaultTenant(ctx context.Context, adminEmail, adminPassword, domain string) (created bool, err error)
	TryHealDefaultDomain(ctx context.Context, host string) error
}

type TenantUsecase interface {
	Create(ctx context.Context, req dto.CreateRequest) (*domain.Tenant, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateRequest) (*domain.Tenant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	List(ctx context.Context, f dto.Filter) ([]domain.Tenant, int64, error)
	SetSupportAccess(ctx context.Context, id uuid.UUID, level domain.SupportAccess) (*domain.Tenant, error)
	Terminate(ctx context.Context, id uuid.UUID) error
	Reactivate(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	PermanentlyDelete(ctx context.Context, id uuid.UUID) error
	ResolveDomain(ctx context.Context, host string) (*domain.Tenant, error)
}
