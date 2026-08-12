package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

// Verify partial-failure recording: one tenant fails, one succeeds.
func TestBulkIDPResult_PartialFailure(t *testing.T) {
	var result common_models.BulkResult
	result.Append("tenant-a", nil)
	result.Append("tenant-b", errors.New("not found"))

	if len(result.Succeeded) != 1 || result.Succeeded[0] != "tenant-a" {
		t.Fatalf("expected tenant-a in succeeded, got %v", result.Succeeded)
	}
	if len(result.Failed) != 1 || result.Failed[0].TenantID != "tenant-b" {
		t.Fatalf("expected tenant-b in failed, got %v", result.Failed)
	}
}

// Verify resolveIDPTenants respects AllTenants flag.
func TestResolveIDPTenants(t *testing.T) {
	lister := func(_ context.Context) ([]string, error) { return []string{"x", "y"}, nil }

	// explicit list
	ids, err := resolveIDPTenants(context.Background(), common_models.BulkTenantSelector{TenantIDs: []string{"a"}}, lister)
	if err != nil || len(ids) != 1 || ids[0] != "a" {
		t.Fatalf("explicit: got %v %v", ids, err)
	}

	// all tenants
	ids, err = resolveIDPTenants(context.Background(), common_models.BulkTenantSelector{AllTenants: true}, lister)
	if err != nil || len(ids) != 2 {
		t.Fatalf("allTenants: got %v %v", ids, err)
	}
}
