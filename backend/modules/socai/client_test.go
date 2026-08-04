package socai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

// The agent has to be told which tenant it is working for, or it picks the
// wrong provider and its lookups back into the MCP arrive unscoped.
func TestClientSendsTheActingTenant(t *testing.T) {
	var gotTenant, gotKey string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTenant = r.Header.Get(middleware.TenantHeader)
		gotKey = r.Header.Get(middleware.InternalKeyHeader)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewSocAIClient(srv.URL, "the-key")
	ctx := authz.WithTenantID(context.Background(), "tenant-a")

	if _, _, err := c.Analyze(ctx, []byte(`{}`)); err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if gotTenant != "tenant-a" {
		t.Fatalf("%s = %q, want %q", middleware.TenantHeader, gotTenant, "tenant-a")
	}
	if gotKey != "the-key" {
		t.Fatalf("%s = %q, want the internal key", middleware.InternalKeyHeader, gotKey)
	}
}

// On a single-tenant install there is no acting tenant, and the header must be
// absent rather than empty so the plugin falls back to the instance default.
func TestClientOmitsTheHeaderWithoutATenant(t *testing.T) {
	var present bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, present = r.Header[http.CanonicalHeaderKey(middleware.TenantHeader)]
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewSocAIClient(srv.URL, "the-key")

	if _, _, err := c.Analyze(context.Background(), []byte(`{}`)); err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if present {
		t.Fatal("the tenant header was sent with no acting tenant")
	}
}

func TestStreamAgentTaskSendsTheActingTenant(t *testing.T) {
	var gotTenant string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTenant = r.Header.Get(middleware.TenantHeader)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewSocAIClient(srv.URL, "the-key")
	ctx := authz.WithTenantID(context.Background(), "tenant-b")

	resp, err := c.StreamAgentTask(ctx, []byte(`{}`))
	if err != nil {
		t.Fatalf("StreamAgentTask: %v", err)
	}
	defer resp.Body.Close()

	if gotTenant != "tenant-b" {
		t.Fatalf("%s = %q, want %q", middleware.TenantHeader, gotTenant, "tenant-b")
	}
}
