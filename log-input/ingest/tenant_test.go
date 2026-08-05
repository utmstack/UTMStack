package ingest

import (
	"context"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"

	"github.com/utmstack/UTMStack/log-input/config"
)

func newServer() *Server {
	return &Server{cfg: &config.Config{DefaultTenant: "default-tenant"}}
}

// A connector that names someone else's tenant writes into its own. This is the
// whole reason the tenant comes from the credential.
func TestPushedTenantCannotBeChosen(t *testing.T) {
	s := newServer()
	ctx := WithTenant(context.Background(), "tenant-a")

	l := &plugins.Log{TenantId: "tenant-b"}
	s.applyDefaults(ctx, l)

	if l.TenantId != "tenant-a" {
		t.Errorf("tenant = %q, want the credential's tenant", l.TenantId)
	}
}

func TestTenantIsTakenFromTheCredential(t *testing.T) {
	s := newServer()
	ctx := WithTenant(context.Background(), "tenant-a")

	l := &plugins.Log{}
	s.applyDefaults(ctx, l)

	if l.TenantId != "tenant-a" {
		t.Errorf("tenant = %q, want tenant-a", l.TenantId)
	}
}

// The internal key belongs to UTMStack's own services, which push on behalf of
// every tenant, so what they send stands.
func TestInternalCallerKeepsTheTenantItSent(t *testing.T) {
	s := newServer()

	l := &plugins.Log{TenantId: "tenant-b"}
	s.applyDefaults(context.Background(), l)

	if l.TenantId != "tenant-b" {
		t.Errorf("tenant = %q, want tenant-b", l.TenantId)
	}
}

func TestNoTenantAnywhereFallsBackToTheDefault(t *testing.T) {
	s := newServer()

	l := &plugins.Log{}
	s.applyDefaults(context.Background(), l)

	if l.TenantId != "default-tenant" {
		t.Errorf("tenant = %q, want the default", l.TenantId)
	}
}
