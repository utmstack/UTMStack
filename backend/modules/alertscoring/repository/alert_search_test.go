package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/threatwinds/go-sdk/store"

	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

// Scoring compares an alert against its own history: how often this fired
// before, what else touched the host. Reading that unscoped would score one
// customer's alert using another customer's traffic — and the answer would look
// perfectly reasonable, which is what makes it worth a test.
func TestTheScopeCarriesTheCallersTenant(t *testing.T) {
	const tenant = "8f1c1b8e-0000-4000-8000-00000000000a"

	got, err := scope(authz.WithTenantID(context.Background(), tenant))
	if err != nil {
		t.Fatalf("scope: %v", err)
	}
	if got.Tenant != tenant {
		t.Errorf("scope tenant = %q, want the caller's", got.Tenant)
	}
	if got.Dataset != eventstore.DatasetAlerts {
		t.Errorf("dataset = %q, want alerts", got.Dataset)
	}
}

// On-prem runs with no tenant and must keep scoring; only a multitenant install
// refuses, and it refuses rather than quietly reading everything.
func TestWithoutATenantItEitherRefusesOrReadsAll(t *testing.T) {
	got, err := scope(context.Background())

	if tenancy.Enabled() {
		if !errors.Is(err, ErrNoTenantScope) {
			t.Errorf("err = %v, want ErrNoTenantScope while multitenancy is on", err)
		}
		return
	}
	if err != nil {
		t.Fatalf("single-tenant install was refused: %v", err)
	}
	if got.Tenant != store.AllTenants {
		t.Errorf("tenant = %q, want the all-tenants scope on a single-tenant install", got.Tenant)
	}
}
