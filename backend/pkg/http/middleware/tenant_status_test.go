package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/pkg/jwt"
)

// ACC-F34: a suspended/terminated tenant's already-issued JWTs must stop
// working on the very next request — token verification is a pure
// signature/expiry check with no DB lookup, so this has to be enforced here.
func TestAuthenticateRejectsSuspendedTenant(t *testing.T) {
	signer := jwt.NewSigner("test-secret", time.Hour)
	token, _, err := signer.Sign(uuid.New(), "user@x.com", nil, nil, uuid.New(), "suspended-tenant", false)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	tenantActive := func(_ context.Context, tenantID string) (bool, error) {
		return tenantID != "suspended-tenant", nil
	}

	rec := httptest.NewRecorder()
	engine := gin.New()
	reached := false
	engine.Use(Authenticate(signer, nil, "", nil, tenantActive))
	engine.GET("/x", func(c *gin.Context) { reached = true; c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if reached {
		t.Fatal("handler ran for a suspended tenant's token")
	}
}

// A healthy tenant's token must keep working unaffected.
func TestAuthenticateAllowsActiveTenant(t *testing.T) {
	signer := jwt.NewSigner("test-secret", time.Hour)
	token, _, err := signer.Sign(uuid.New(), "user@x.com", nil, nil, uuid.New(), "active-tenant", false)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	tenantActive := func(_ context.Context, tenantID string) (bool, error) {
		return tenantID == "active-tenant", nil
	}

	rec := httptest.NewRecorder()
	engine := gin.New()
	reached := false
	engine.Use(Authenticate(signer, nil, "", nil, tenantActive))
	engine.GET("/x", func(c *gin.Context) { reached = true; c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || !reached {
		t.Fatalf("status = %d reached=%v, want 200/true", rec.Code, reached)
	}
}

// Internal (service-to-service) callers never carry a tenant-plane JWT, so the
// check must not apply to them: it would otherwise break every internal job
// that has no tenant selected.
func TestAuthenticateInternalCallerBypassesTenantStatus(t *testing.T) {
	calls := 0
	tenantActive := func(context.Context, string) (bool, error) {
		calls++
		return false, nil
	}

	rec := httptest.NewRecorder()
	engine := gin.New()
	engine.Use(Authenticate(nil, nil, testInternalKey, nil, tenantActive))
	engine.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set(InternalKeyHeader, testInternalKey)
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if calls != 0 {
		t.Fatalf("tenantActive called %d times for an internal caller, want 0", calls)
	}
}
