// Package authz holds transport-agnostic authorization predicates and the
// Actor type representing an authenticated caller. The HTTP middleware in
// pkg/http/middleware and the MCP tool handlers in modules/mcp both call into
// this package so the gates that reject a REST request reject the equivalent
// MCP tool call.
package authz

import (
	"errors"
	"slices"
)

const RoleAdmin = "ROLE_ADMIN"

type Actor struct {
	UserID      uint64
	Login       string
	Email       string
	Roles       []string
	Permissions []string
	SessionID   uint64
	Internal    bool
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

func RequireInternal(a *Actor) error {
	if a == nil {
		return ErrUnauthenticated
	}
	if a.Internal {
		return nil
	}
	return ErrInternalOnly
}
