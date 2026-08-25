package usecase

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync/atomic"

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

type bootstrapUsecase struct {
	repo   connectors.TenantRepository
	admin  connectors.UserProvisioner
	healed atomic.Bool
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
		if existing.Domain != "" {
			u.healed.Store(true)
		}
		return false, nil
	}

	if adminPassword == "" {
		return false, ErrBootstrapPasswordRequired
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
	if tenantDomain != "" {
		u.healed.Store(true)
	}

	if err := provisionAdmin(ctx, u.admin, t.ID, adminEmail, adminPassword, false); err != nil {
		if delErr := u.repo.Delete(all, t.ID); delErr != nil {
			return false, fmt.Errorf("%w (and the tenant row could not be rolled back: %v)", err, delErr)
		}
		return false, err
	}
	return true, nil
}

// TryHealDefaultDomain stamps the default tenant's Domain from the first
// request's Host when UTMSTACK_DEFAULT_DOMAIN was not set at install time.
// Runs at most once per process (atomic fast path), and no-ops once the row
// already has a domain.
func (u *bootstrapUsecase) TryHealDefaultDomain(ctx context.Context, host string) error {
	if u.healed.Load() {
		return nil
	}
	host = normalizeHealHost(host)
	if host == "" {
		return nil
	}
	all := tenancy.WithAllTenants(ctx)
	t, err := u.repo.FindByID(all, defaultTenantID)
	if err != nil || t == nil {
		return err
	}
	if t.Domain != "" {
		u.healed.Store(true)
		return nil
	}
	t.Domain = host
	if err := u.repo.Update(all, t); err != nil {
		return err
	}
	u.healed.Store(true)
	return nil
}

func normalizeHealHost(raw string) string {
	// X-Forwarded-Host may carry a comma-separated chain; the first hop is the
	// original client-facing hostname.
	if i := strings.IndexByte(raw, ','); i >= 0 {
		raw = raw[:i]
	}
	raw = strings.TrimSpace(raw)
	if h, _, err := net.SplitHostPort(raw); err == nil {
		raw = h
	}
	raw = strings.ToLower(raw)
	// Single-label hosts ("backend", "localhost", any compose service name)
	// are internal identities, never the public domain the admin browsed to.
	// Real user-facing hosts are FQDNs (dotted) or bare IP literals.
	if !strings.Contains(raw, ".") && net.ParseIP(raw) == nil {
		return ""
	}
	return raw
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
