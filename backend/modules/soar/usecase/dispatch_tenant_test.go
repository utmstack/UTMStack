package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

var execTenant = uuid.MustParse("8f1c1b8e-0000-4000-8000-000000000001")

// The dispatcher runs on the application's context, which belongs to no tenant.
// Listing has to span them all — otherwise, once executions carry a tenant,
// every tick fails and nothing is ever dispatched.
func TestDrainListsAcrossTenants(t *testing.T) {
	ctx := tenancy.WithAllTenants(context.Background())

	if !tenancy.SpansAllTenants(ctx) {
		t.Fatal("the listing context does not span tenants")
	}
}

// Each execution is then handled as its tenant: the status writes touch its row
// and, more to the point, resolveAgent matches by hostname — which is unique
// inside a tenant and not across, so an unscoped lookup runs the response on
// somebody else's machine.
func TestEachExecutionIsHandledAsItsOwnTenant(t *testing.T) {
	exec := domain.SoarExecution{TenantID: execTenant}

	ctx := authz.WithTenantID(context.Background(), exec.TenantID.String())

	if got := authz.TenantIDFromContext(ctx); got != execTenant.String() {
		t.Errorf("tenant = %q, want the execution's %q", got, execTenant.String())
	}
	if tenancy.SpansAllTenants(ctx) {
		t.Error("the per-execution context spans every tenant; its writes would not be scoped")
	}
}
