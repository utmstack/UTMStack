package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

// stubLister returns a fixed list for testing.
func stubLister(ids ...string) func(context.Context) ([]string, error) {
	return func(_ context.Context) ([]string, error) { return ids, nil }
}

func TestResolveTenants_explicit(t *testing.T) {
	sel := common_models.BulkTenantSelector{TenantIDs: []string{"a", "b"}}
	got, err := resolveTenants(context.Background(), sel, nil)
	if err != nil || len(got) != 2 {
		t.Fatalf("want 2 ids, got %v %v", got, err)
	}
}

func TestResolveTenants_allFiltersDefault(t *testing.T) {
	const defaultTID = "ce66672c-e36d-4761-a8c8-90058fee1a24"
	sel := common_models.BulkTenantSelector{AllTenants: true}
	got, err := resolveTenants(context.Background(), sel, stubLister(defaultTID, "tenant-2"))
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range got {
		if id == defaultTID {
			t.Fatalf("DefaultTenantID must be filtered out, got %v", got)
		}
	}
	if len(got) != 1 || got[0] != "tenant-2" {
		t.Fatalf("want [tenant-2], got %v", got)
	}
}

func TestBulkResultPartialFailure(t *testing.T) {
	var r common_models.BulkResult
	r.Append("t1", nil)
	r.Append("t2", errors.New("boom"))
	r.Append("t3", nil)
	if len(r.Succeeded) != 2 || len(r.Failed) != 1 {
		t.Fatalf("want 2 succeeded 1 failed, got %v / %v", r.Succeeded, r.Failed)
	}
	if r.Failed[0].TenantID != "t2" {
		t.Fatalf("wrong failed tenant: %v", r.Failed[0])
	}
}
