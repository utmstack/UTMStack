package middleware

import (
	"context"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

const otherTenant = "8f1c1b8e-0000-4000-8000-000000000001"

// cross runs setActor for a credential belonging to one tenant against a host
// belonging to another, and reports what the request was allowed to become.
func cross(t *testing.T, method string, host string, support string, a Actor) (*authz.Actor, bool) {
	t.Helper()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(method, "/x", nil)
	c.Set(HostTenantKey, host)
	c.Set(HostSupportKey, support)

	if !setActor(c, a, nil) {
		return nil, false
	}
	return ActorFromGin(c), true
}

// crossByHeader is the other way in: the credential is on its own host and names
// the tenant it wants to support.
func crossByHeader(t *testing.T, method, target, support string, a Actor) (*authz.Actor, bool) {
	t.Helper()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(method, "/x", nil)
	c.Request.Header.Set(TenantHeader, target)
	c.Set(HostTenantKey, a.TenantID)

	supportOf := func(context.Context, string) (string, error) { return support, nil }

	if !setActor(c, a, supportOf) {
		return nil, false
	}
	return ActorFromGin(c), true
}

func platformAdmin() Actor {
	return Actor{
		UserID:      uuid.New(),
		SessionID:   uuid.New(),
		Roles:       []string{authz.RoleAdmin},
		Permissions: []string{"alerts.read", "alerts.write", "users.read", "users.delete"},
		TenantID:    authz.DefaultTenantID,
	}
}

// A tenant that has granted nothing is unreachable, which is the state every
// tenant is in until its administrator says otherwise.
func TestSupportDeniedWhenNotGranted(t *testing.T) {
	for _, level := range []string{"", authz.SupportNone, "anything-else"} {
		if _, ok := cross(t, http.MethodGet, otherTenant, level, platformAdmin()); ok {
			t.Errorf("support access %q let a platform admin in", level)
		}
	}
}

func TestSupportReadAllowsReadsOnly(t *testing.T) {
	actor, ok := cross(t, http.MethodGet, otherTenant, authz.SupportRead, platformAdmin())
	if !ok {
		t.Fatal("a READ grant turned away a GET")
	}

	if actor.TenantID != otherTenant {
		t.Errorf("tenant = %q, want the supported tenant %q", actor.TenantID, otherTenant)
	}
	if actor.Support != authz.SupportRead {
		t.Errorf("support = %q, want READ", actor.Support)
	}

	// The allowlist keeps reads and drops everything else, including
	// users.delete, which is neither a read nor a write by name.
	want := []string{"alerts.read", "users.read"}
	if !slices.Equal(actor.Permissions, want) {
		t.Errorf("permissions = %v, want %v", actor.Permissions, want)
	}
}

// The permission allowlist alone is not enough: a route that requires no
// permission would be a way through, so the method is checked too.
func TestSupportReadRefusesWrites(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		if _, ok := cross(t, method, otherTenant, authz.SupportRead, platformAdmin()); ok {
			t.Errorf("a READ grant allowed %s", method)
		}
	}
}

func TestSupportFullAllowsWrites(t *testing.T) {
	actor, ok := cross(t, http.MethodDelete, otherTenant, authz.SupportFull, platformAdmin())
	if !ok {
		t.Fatal("a FULL grant turned away a DELETE")
	}
	if actor.TenantID != otherTenant {
		t.Errorf("tenant = %q, want the supported tenant", actor.TenantID)
	}
	if !slices.Contains(actor.Permissions, "alerts.write") {
		t.Errorf("permissions = %v, want writes kept", actor.Permissions)
	}
}

// Support access is what the platform may use, not a door into any tenant from
// any other. An administrator of one customer stays inside it even where a
// grant exists, because the grant was made to the platform.
func TestSupportIsNotAWayBetweenTenants(t *testing.T) {
	a := platformAdmin()
	a.TenantID = "some-other-customer"

	for _, level := range []string{authz.SupportRead, authz.SupportFull} {
		if _, ok := cross(t, http.MethodGet, otherTenant, level, a); ok {
			t.Errorf("a tenant administrator crossed into another tenant under %s", level)
		}
	}
}

// Belonging to the default tenant is not itself the platform plane; an analyst
// there is an ordinary user of an ordinary tenant.
func TestSupportRequiresAdminNotJustDefaultTenant(t *testing.T) {
	a := platformAdmin()
	a.Roles = []string{"ROLE_USER"}

	if _, ok := cross(t, http.MethodGet, otherTenant, authz.SupportFull, a); ok {
		t.Error("a non-administrator of the default tenant crossed into a tenant")
	}
}

// Nothing changes for a credential used against its own tenant: no support
// session, no narrowing.
func TestOwnTenantIsUntouched(t *testing.T) {
	a := platformAdmin()

	actor, ok := cross(t, http.MethodDelete, authz.DefaultTenantID, authz.SupportNone, a)
	if !ok {
		t.Fatal("a credential was turned away from its own tenant")
	}
	if actor.Support != "" {
		t.Errorf("support = %q, want empty on an ordinary request", actor.Support)
	}
	if !slices.Equal(actor.Permissions, a.Permissions) {
		t.Errorf("permissions = %v, want them untouched", actor.Permissions)
	}
}

func TestReadOnlyPermissionsIsAnAllowlist(t *testing.T) {
	got := authz.ReadOnlyPermissions([]string{
		"alerts.read", "alerts.write", "users.delete", "checkEmailConfiguration", "tenant.read",
	})
	want := []string{"alerts.read", "tenant.read"}

	if !slices.Equal(got, want) {
		t.Errorf("ReadOnlyPermissions = %v, want %v", got, want)
	}
}

func TestSupportByHeader(t *testing.T) {
	actor, ok := crossByHeader(t, http.MethodGet, otherTenant, authz.SupportRead, platformAdmin())
	if !ok {
		t.Fatal("a READ grant turned away a platform administrator naming the tenant")
	}
	if actor.TenantID != otherTenant {
		t.Errorf("tenant = %q, want the supported tenant", actor.TenantID)
	}
	if actor.Support != authz.SupportRead {
		t.Errorf("support = %q, want READ", actor.Support)
	}
	if slices.Contains(actor.Permissions, "alerts.write") {
		t.Errorf("permissions = %v, want writes dropped", actor.Permissions)
	}
}

// Naming a tenant is not itself permission to enter it.
func TestSupportByHeaderNeedsTheGrant(t *testing.T) {
	if _, ok := crossByHeader(t, http.MethodGet, otherTenant, authz.SupportNone, platformAdmin()); ok {
		t.Error("a header let a platform admin into a tenant that granted nothing")
	}
}

func TestSupportByHeaderRefusesNonPlatform(t *testing.T) {
	a := platformAdmin()
	a.TenantID = "some-other-customer"

	if _, ok := crossByHeader(t, http.MethodGet, otherTenant, authz.SupportFull, a); ok {
		t.Error("a tenant administrator crossed into another tenant by naming it")
	}
}

func TestSupportByHeaderKeepsTheMethodGuard(t *testing.T) {
	if _, ok := crossByHeader(t, http.MethodPost, otherTenant, authz.SupportRead, platformAdmin()); ok {
		t.Error("a READ grant allowed a POST")
	}
}

// An API key has no session. Support access is a person looking at a customer's
// data, and a key outlives the reason it was issued for.
func TestSupportRefusesCredentialsWithoutASession(t *testing.T) {
	a := platformAdmin()
	a.SessionID = uuid.Nil

	if _, ok := crossByHeader(t, http.MethodGet, otherTenant, authz.SupportFull, a); ok {
		t.Error("an API key crossed into a tenant by naming it")
	}
	if _, ok := cross(t, http.MethodGet, otherTenant, authz.SupportFull, a); ok {
		t.Error("an API key crossed into a tenant by host")
	}
}

// A tenant that granted nothing is not a reason to reject a request that was
// not trying to cross in the first place.
func TestOwnTenantHeaderIsNotACrossing(t *testing.T) {
	a := platformAdmin()

	actor, ok := crossByHeader(t, http.MethodDelete, a.TenantID, authz.SupportNone, a)
	if !ok {
		t.Fatal("naming its own tenant turned a request away")
	}
	if actor.Support != "" {
		t.Errorf("support = %q, want empty", actor.Support)
	}
}
