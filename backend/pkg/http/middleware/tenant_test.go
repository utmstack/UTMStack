package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/utmstack/utmstack/backend/pkg/authz"
)

const testInternalKey = "internal-key"

func init() { gin.SetMode(gin.TestMode) }

// run wires ResolveTenant in front of a handler that records what it saw.
func run(t *testing.T, isMSSP func() bool, resolve func(context.Context, string) (string, string, error), req *http.Request) (*httptest.ResponseRecorder, string, bool) {
	t.Helper()

	rec := httptest.NewRecorder()
	engine := gin.New()

	var sawTenant string
	var reached bool
	engine.Use(ResolveTenant(isMSSP, testInternalKey, resolve))
	engine.GET("/x", func(c *gin.Context) {
		reached = true
		sawTenant = authz.TenantIDFromContext(c.Request.Context())
		c.Status(http.StatusOK)
	})
	engine.ServeHTTP(rec, req)

	return rec, sawTenant, reached
}

func newReq(host string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.Host = host
	return r
}

func byDomain(m map[string]string) func(context.Context, string) (string, string, error) {
	return func(_ context.Context, host string) (string, string, error) {
		if id, ok := m[host]; ok {
			return id, "", nil
		}
		return "", "", errors.New("not found")
	}
}

func TestResolveTenantPutsTheHostTenantOnTheContext(t *testing.T) {
	rec, tenant, reached := run(t,
		func() bool { return true },
		byDomain(map[string]string{"a.example.com": "tenant-a"}),
		newReq("a.example.com"))

	if !reached || rec.Code != http.StatusOK {
		t.Fatalf("handler not reached; status = %d", rec.Code)
	}
	if tenant != "tenant-a" {
		t.Fatalf("tenant on context = %q, want %q", tenant, "tenant-a")
	}
}

// The port is part of Host often enough that it has to be stripped somewhere;
// ResolveDomain does it, so an unresolvable host is genuinely unknown.
func TestResolveTenantRefusesAnUnknownHost(t *testing.T) {
	rec, _, reached := run(t,
		func() bool { return true },
		byDomain(map[string]string{"a.example.com": "tenant-a"}),
		newReq("nobody.example.com"))

	if reached {
		t.Fatal("the handler ran for a host that belongs to no tenant")
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

// The regression this guards: every plugin and the agent manager reach the
// backend at http://backend:8080, which is nobody's domain. Refusing them
// breaks datasource pings, the alerts cache, feeds and soc-ai on every MSSP
// instance — and single-tenant testing never shows it, because the middleware
// is inert there.
func TestResolveTenantLetsInternalCallersThrough(t *testing.T) {
	req := newReq("backend:8080")
	req.Header.Set(InternalKeyHeader, testInternalKey)

	rec, tenant, reached := run(t,
		func() bool { return true },
		byDomain(map[string]string{"a.example.com": "tenant-a"}),
		req)

	if !reached || rec.Code != http.StatusOK {
		t.Fatalf("an internal caller was refused; status = %d", rec.Code)
	}
	// No tenant: an internal caller is the platform plane. Anything
	// tenant-scoped it touches still fails closed.
	if tenant != "" {
		t.Fatalf("tenant on context = %q, want empty for an internal caller", tenant)
	}
}

func TestResolveTenantIgnoresAWrongInternalKey(t *testing.T) {
	req := newReq("backend:8080")
	req.Header.Set(InternalKeyHeader, "not-the-key")

	rec, _, reached := run(t,
		func() bool { return true },
		byDomain(map[string]string{"a.example.com": "tenant-a"}),
		req)

	if reached || rec.Code != http.StatusNotFound {
		t.Fatalf("a wrong internal key was accepted; status = %d", rec.Code)
	}
}

// On a single-tenant install the whole thing is inert, so an install that never
// takes an MSSP licence behaves exactly as it did before.
func TestResolveTenantIsInertWithoutMSSP(t *testing.T) {
	called := false
	resolve := func(context.Context, string) (string, string, error) {
		called = true
		return "", "", errors.New("should not be consulted")
	}

	rec, tenant, reached := run(t, func() bool { return false }, resolve, newReq("anything"))

	if !reached || rec.Code != http.StatusOK {
		t.Fatalf("handler not reached; status = %d", rec.Code)
	}
	if called {
		t.Error("the resolver was consulted on a single-tenant install")
	}
	if tenant != "" {
		t.Errorf("tenant on context = %q, want empty", tenant)
	}
}

// The convention: an internal caller declares which tenant it is acting for,
// because nothing else in the request says. Without it the datasource ping,
// the soc-ai agent and every plugin callback arrive with no tenant and every
// scoped write fails closed.
func TestInternalCallerDeclaresItsTenant(t *testing.T) {
	req := newReq("backend:8080")
	req.Header.Set(InternalKeyHeader, testInternalKey)
	req.Header.Set(TenantHeader, "tenant-a")

	rec, tenant := runAuthenticate(t, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if tenant != "tenant-a" {
		t.Fatalf("acting tenant = %q, want %q", tenant, "tenant-a")
	}
}

// The security property: naming a tenant requires proving you are internal, or
// reading another tenant's data would be one header away.
func TestTenantHeaderIsIgnoredWithoutTheInternalKey(t *testing.T) {
	req := newReq("backend:8080")
	req.Header.Set(TenantHeader, "tenant-a")

	rec, _ := runAuthenticate(t, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401: the header alone must not authenticate", rec.Code)
	}
}

func TestInternalCallerWithoutATenantGetsNone(t *testing.T) {
	req := newReq("backend:8080")
	req.Header.Set(InternalKeyHeader, testInternalKey)

	rec, tenant := runAuthenticate(t, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	// Not an error: platform-plane work has no tenant, and anything scoped it
	// touches still fails closed.
	if tenant != "" {
		t.Fatalf("acting tenant = %q, want empty", tenant)
	}
}

// runAuthenticate drives Authenticate and reports the acting tenant it left on
// the request context.
func runAuthenticate(t *testing.T, req *http.Request) (*httptest.ResponseRecorder, string) {
	t.Helper()

	rec := httptest.NewRecorder()
	engine := gin.New()

	var sawTenant string
	engine.Use(Authenticate(nil, nil, testInternalKey, nil))
	engine.GET("/x", func(c *gin.Context) {
		sawTenant = authz.TenantIDFromContext(c.Request.Context())
		c.Status(http.StatusOK)
	})
	engine.ServeHTTP(rec, req)

	return rec, sawTenant
}
