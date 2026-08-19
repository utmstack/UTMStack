package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

// noopLister returns two tenant IDs for testing.
func noopLister(_ context.Context) ([]string, error) {
	return []string{"tenant-a", "tenant-b"}, nil
}

func TestBulkResult_PartialFailure(t *testing.T) {
	var r common_models.BulkResult
	r.Append("tenant-a", nil)
	r.Append("tenant-b", errors.New("smtp refused"))

	if len(r.Succeeded) != 1 || r.Succeeded[0] != "tenant-a" {
		t.Fatalf("expected one success, got %v", r.Succeeded)
	}
	if len(r.Failed) != 1 || r.Failed[0].TenantID != "tenant-b" {
		t.Fatalf("expected one failure, got %v", r.Failed)
	}
}

func TestResolveTenants_AllExcludesDefault(t *testing.T) {
	lister := func(_ context.Context) ([]string, error) {
		return []string{"tenant-a", "ce66672c-e36d-4761-a8c8-90058fee1a24", "tenant-b"}, nil
	}
	sel := common_models.BulkTenantSelector{AllTenants: true}
	ids, err := resolveTenants(context.Background(), sel, lister)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range ids {
		if id == "ce66672c-e36d-4761-a8c8-90058fee1a24" {
			t.Fatal("DefaultTenantID must be excluded from AllTenants enumeration")
		}
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 tenants, got %d", len(ids))
	}
}

func TestResolveTenants_ExplicitList(t *testing.T) {
	sel := common_models.BulkTenantSelector{TenantIDs: []string{"tenant-x"}}
	ids, err := resolveTenants(context.Background(), sel, noopLister)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != "tenant-x" {
		t.Fatalf("expected explicit list passthrough, got %v", ids)
	}
}
