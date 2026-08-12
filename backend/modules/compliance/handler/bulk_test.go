package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

// stubLister returns fixed tenant IDs.
func stubLister(ids []string) func(context.Context) ([]string, error) {
	return func(_ context.Context) ([]string, error) { return ids, nil }
}

func TestResolveTenants_AllTenants(t *testing.T) {
	want := []string{"a", "b"}
	got, err := resolveTenants(context.Background(),
		common_models.BulkTenantSelector{AllTenants: true},
		stubLister(want))
	if err != nil || len(got) != len(want) {
		t.Fatalf("expected %v, got %v err %v", want, got, err)
	}
}

func TestResolveTenants_ExplicitIDs(t *testing.T) {
	ids := []string{"x", "y"}
	got, err := resolveTenants(context.Background(),
		common_models.BulkTenantSelector{TenantIDs: ids},
		func(_ context.Context) ([]string, error) { return nil, errors.New("should not be called") })
	if err != nil || len(got) != 2 {
		t.Fatalf("expected explicit ids, got %v err %v", got, err)
	}
}

// TestBulkResult_PartialFailure checks that Append records both success and failure.
func TestBulkResult_PartialFailure(t *testing.T) {
	var r common_models.BulkResult
	r.Append("t1", nil)
	r.Append("t2", errors.New("boom"))
	r.Append("t3", nil)
	if len(r.Succeeded) != 2 || len(r.Failed) != 1 {
		t.Fatalf("succeeded=%d failed=%d", len(r.Succeeded), len(r.Failed))
	}
	if r.Failed[0].TenantID != "t2" {
		t.Fatalf("wrong failed tenant: %s", r.Failed[0].TenantID)
	}
}
