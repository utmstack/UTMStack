package authz

import (
	"context"
	"errors"
	"slices"
)

const RoleAdmin = "ROLE_ADMIN"
const RolePlatformAdmin = "ROLE_PLATFORM_ADMIN"

type Actor struct {
	UserID      uint64
	Login       string
	Email       string
	Roles       []string
	Permissions []string
	SessionID   uint64
	Internal    bool
	TenantID    string
}

var (
	ErrUnauthenticated    = errors.New("unauthenticated")
	ErrMissingPermission  = errors.New("missing required permission")
	ErrMissingRole        = errors.New("missing required role")
	ErrEnterpriseRequired = errors.New("this feature requires an Enterprise license")
	ErrMSSPRequired       = errors.New("this feature requires an MSSP license")
	ErrInternalOnly       = errors.New("internal-only endpoint")
)

func HasPermission(a *Actor, perm string) bool {
	if a == nil {
		return false
	}
	if a.Internal {
		return true
	}
	return slices.Contains(a.Permissions, perm)
}

func HasRole(a *Actor, role string) bool {
	if a == nil {
		return false
	}
	if a.Internal {
		return true
	}
	return slices.Contains(a.Roles, role)
}

func RequirePermission(a *Actor, perm string) error {
	if a == nil {
		return ErrUnauthenticated
	}
	if HasPermission(a, perm) {
		return nil
	}
	return ErrMissingPermission
}

func RequireRole(a *Actor, role string) error {
	if a == nil {
		return ErrUnauthenticated
	}
	if HasRole(a, role) {
		return nil
	}
	return ErrMissingRole
}

func RequireAdmin(a *Actor) error { return RequireRole(a, RoleAdmin) }

func RequireEnterprise(a *Actor, isEnterprise func() bool) error {
	if a == nil {
		return ErrUnauthenticated
	}
	if a.Internal {
		return nil
	}
	if isEnterprise != nil && isEnterprise() {
		return nil
	}
	return ErrEnterpriseRequired
}

func RequireMSSP(a *Actor, isMSSP func() bool) error {
	if a == nil {
		return ErrUnauthenticated
	}
	if a.Internal {
		return nil
	}
	if isMSSP != nil && isMSSP() {
		return nil
	}
	return ErrMSSPRequired
}

// RequirePlatform gates the platform plane. On a single-tenant install there is
// no separate plane to protect, so the ordinary administrator is it; once the
// instance serves several tenants the platform role becomes mandatory, because
// a tenant admin holding ROLE_ADMIN must not reach the other tenants.
func RequirePlatform(a *Actor, isMSSP func() bool) error {
	if a == nil {
		return ErrUnauthenticated
	}
	if HasRole(a, RolePlatformAdmin) {
		return nil
	}
	if isMSSP != nil && isMSSP() {
		return ErrMissingRole
	}
	return RequireRole(a, RoleAdmin)
}

func RequireInternal(a *Actor) error {
	if a == nil {
		return ErrUnauthenticated
	}
	if a.Internal {
		return nil
	}
	return ErrInternalOnly
}

type ctxTenantKey struct{}

// WithTenantID stashes the acting tenant on ctx, for usecases that need it
// without threading a new parameter through every call. Handlers set this
// once (from the request's Actor) before calling into a usecase.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, ctxTenantKey{}, tenantID)
}

// TenantIDFromContext returns the acting tenant stashed by WithTenantID, or
// "" if none was set — the same "empty means global/on-prem" convention used
// everywhere else.
func TenantIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxTenantKey{}).(string)
	return v
}
