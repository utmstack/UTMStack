package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

// stubLister returns the given IDs.
func stubLister(ids []string) func(context.Context) ([]string, error) {
	return func(context.Context) ([]string, error) { return ids, nil }
}

func TestResolveTenants_AllTenants(t *testing.T) {
	ids, err := resolveTenants(context.Background(),
		common_models.BulkTenantSelector{AllTenants: true},
		stubLister([]string{"a", "b"}))
	if err != nil || len(ids) != 2 {
		t.Fatalf("expected 2 ids, got %v err=%v", ids, err)
	}
}

func TestResolveTenants_Specific(t *testing.T) {
	ids, err := resolveTenants(context.Background(),
		common_models.BulkTenantSelector{TenantIDs: []string{"x"}},
		stubLister([]string{"ignored"}))
	if err != nil || len(ids) != 1 || ids[0] != "x" {
		t.Fatalf("expected [x], got %v err=%v", ids, err)
	}
}

func TestBulkResult_PartialFailure(t *testing.T) {
	var r common_models.BulkResult
	r.Append("ok-tenant", nil)
	r.Append("bad-tenant", errors.New("boom"))

	if len(r.Succeeded) != 1 || r.Succeeded[0] != "ok-tenant" {
		t.Fatalf("unexpected succeeded: %v", r.Succeeded)
	}
	if len(r.Failed) != 1 || r.Failed[0].TenantID != "bad-tenant" {
		t.Fatalf("unexpected failed: %v", r.Failed)
	}
}
