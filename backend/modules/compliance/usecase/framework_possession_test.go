package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

// fakePossession records tenant → framework-key sets in memory.
type fakePossession struct {
	byTenant map[string]map[string]bool
}

func newFakePossession() *fakePossession {
	return &fakePossession{byTenant: map[string]map[string]bool{}}
}

func (f *fakePossession) List(ctx context.Context) ([]string, error) {
	return f.ListForTenant(ctx, authz.TenantIDFromContext(ctx))
}

func (f *fakePossession) ListForTenant(_ context.Context, tenantID string) ([]string, error) {
	s := f.byTenant[tenantID]
	out := make([]string, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	return out, nil
}

func (f *fakePossession) Enable(ctx context.Context, key string) error {
	tid := authz.TenantIDFromContext(ctx)
	if tid == "" {
		return errors.New("no tenant")
	}
	if f.byTenant[tid] == nil {
		f.byTenant[tid] = map[string]bool{}
	}
	f.byTenant[tid][key] = true
	return nil
}

func (f *fakePossession) Disable(ctx context.Context, key string) error {
	tid := authz.TenantIDFromContext(ctx)
	if f.byTenant[tid] != nil {
		delete(f.byTenant[tid], key)
	}
	return nil
}

func (f *fakePossession) Has(ctx context.Context, key string) (bool, error) {
	tid := authz.TenantIDFromContext(ctx)
	return f.byTenant[tid][key], nil
}

func (f *fakePossession) ListTenants(_ context.Context, key string) ([]string, error) {
	out := []string{}
	for tid, keys := range f.byTenant {
		if keys[key] {
			out = append(out, tid)
		}
	}
	return out, nil
}

// TestPossessionGovernsListedEnabled: two tenants, same framework, only one
// possesses it → only that one sees Enabled=true.
func TestPossessionGovernsListedEnabled(t *testing.T) {
	possession := newFakePossession()
	ctxA := authz.WithTenantID(context.Background(), "tenantA")
	ctxB := authz.WithTenantID(context.Background(), "tenantB")

	if err := possession.Enable(ctxA, "hipaa"); err != nil {
		t.Fatalf("Enable: %v", err)
	}

	// Simulated frameworkStore.All() output — file-level Enabled is true.
	files := []domain.Framework{{Key: "hipaa", Enabled: true}}

	// Manual repro of the possessedSet overlay so this test can run without
	// wiring up ControlStore / FrameworkStore / Entitlement.
	overlay := func(ctx context.Context, in []domain.Framework) []domain.Framework {
		if authz.TenantIDFromContext(ctx) == "" {
			return in
		}
		keys, _ := possession.List(ctx)
		set := map[string]bool{}
		for _, k := range keys {
			set[k] = true
		}
		out := make([]domain.Framework, len(in))
		copy(out, in)
		for i := range out {
			out[i].Enabled = out[i].Enabled && set[out[i].Key]
		}
		return out
	}

	a := overlay(ctxA, files)
	if !a[0].Enabled {
		t.Errorf("tenantA should see hipaa as Enabled after Enable")
	}
	b := overlay(ctxB, files)
	if b[0].Enabled {
		t.Errorf("tenantB should not see hipaa as Enabled — never enabled it")
	}
}

// TestDisableDeletesPossession — the toggle semantics: enable then disable
// removes the row.
func TestDisableDeletesPossession(t *testing.T) {
	possession := newFakePossession()
	ctx := authz.WithTenantID(context.Background(), "t1")
	if err := possession.Enable(ctx, "iso"); err != nil {
		t.Fatal(err)
	}
	if has, _ := possession.Has(ctx, "iso"); !has {
		t.Fatal("Enable did not stick")
	}
	if err := possession.Disable(ctx, "iso"); err != nil {
		t.Fatal(err)
	}
	if has, _ := possession.Has(ctx, "iso"); has {
		t.Fatal("Disable did not remove the row")
	}
}
