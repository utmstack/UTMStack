package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	iam_connectors "github.com/utmstack/utmstack/backend/modules/iam/connectors"
	iam_dto "github.com/utmstack/utmstack/backend/modules/iam/dto"
	"github.com/utmstack/utmstack/backend/modules/tenant/connectors"
	"github.com/utmstack/utmstack/backend/modules/tenant/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

var defaultTenantID = uuid.MustParse(authz.DefaultTenantID)

const defaultTenantName = "UTMStack"

var ErrBootstrapPasswordRequired = errors.New(
	"UTMSTACK_ADMIN_PASSWORD is required to create the initial administrator")

var ErrBootstrapDomainRequired = errors.New(
	"UTMSTACK_DEFAULT_DOMAIN is required to create the default tenant")

type bootstrapUsecase struct {
	repo  connectors.TenantRepository
	admin connectors.UserProvisioner
}

func NewBootstrapUsecase(repo connectors.TenantRepository, admin connectors.UserProvisioner) connectors.BootstrapUsecase {
	return &bootstrapUsecase{repo: repo, admin: admin}
}

func (u *bootstrapUsecase) EnsureDefaultTenant(ctx context.Context, adminEmail, adminPassword, tenantDomain string) (bool, error) {
	// Only the tenant table is read and written across tenants. The
	// administrator is created on the plain context so the tenancy callback
	// still stamps it: a context that spans every tenant belongs to none, and
	// would leave the row with no owner.
	all := tenancy.WithAllTenants(ctx)

	existing, err := u.repo.FindByID(all, defaultTenantID)
	if err != nil {
		return false, fmt.Errorf("looking up the default tenant: %w", err)
	}
	if existing != nil {
		return false, nil
	}

	if adminPassword == "" {
		return false, ErrBootstrapPasswordRequired
	}
	if tenantDomain == "" {
		return false, ErrBootstrapDomainRequired
	}

	t := &domain.Tenant{
		ID:     defaultTenantID,
		Name:   defaultTenantName,
		Domain: tenantDomain,
		Status: domain.StatusActive,
	}
	if err := u.repo.Create(all, t); err != nil {
		return false, fmt.Errorf("creating the default tenant: %w", err)
	}

	if err := provisionAdmin(ctx, u.admin, t.ID, adminEmail, adminPassword, false); err != nil {
		if delErr := u.repo.Delete(all, t.ID); delErr != nil {
			return false, fmt.Errorf("%w (and the tenant row could not be rolled back: %v)", err, delErr)
		}
		return false, err
	}
	return true, nil
}

func provisionAdmin(ctx context.Context, admin connectors.UserProvisioner, tenantID uuid.UUID, email, password string, invite bool) error {
	if admin == nil {
		return nil
	}
	_, err := admin.Create(
		authz.WithTenantID(ctx, tenantID.String()),
		iam_dto.CreateUserRequest{
			Email:     email,
			RoleNames: []string{authz.RoleAdmin},
		},
		iam_connectors.CreateUserOptions{Password: password, Invite: invite},
	)
	return err
}
