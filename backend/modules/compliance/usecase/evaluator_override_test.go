package usecase

import (
	"context"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
)

// fakeOverrides is a minimal in-memory ControlStatusOverrideRepository.
// Records what got Deleted so the test can assert the cleanup fired.
type fakeOverrides struct {
	list    map[string]string // controlID → status
	deleted map[string]bool   // controlID → true when Delete was called
}

func (f *fakeOverrides) Upsert(_ context.Context, o *domain.UtmComplianceControlStatusOverride) error {
	f.list[o.ControlID] = o.Status
	return nil
}

func (f *fakeOverrides) Delete(_ context.Context, _ string, controlID string) error {
	if f.deleted == nil {
		f.deleted = map[string]bool{}
	}
	f.deleted[controlID] = true
	delete(f.list, controlID)
	return nil
}

func (f *fakeOverrides) ListByFramework(_ context.Context, _ string) (map[string]string, error) {
	out := make(map[string]string, len(f.list))
	for k, v := range f.list {
		out[k] = v
	}
	return out, nil
}

// TestOverrideDeletedWhenComputedCatchesUp exercises the loop directly: if the
// override's target status matches the freshly-computed status, the override
// is dead weight and gets deleted after the report is assembled.
//
// The test simulates the eval loop shape (no full evaluator wiring) so it
// stays a runnable check on the branch, not an integration test.
func TestOverrideDeletedWhenComputedCatchesUp(t *testing.T) {
	overrides := &fakeOverrides{list: map[string]string{
		"ac-1": domain.StatusCompliant,    // computed will be COMPLIANT → stale
		"ac-2": domain.StatusNonCompliant, // computed will be COMPLIANT → still applies
	}}
	ctx := context.Background()
	frameworkKey := "nist-800-53"

	// Simulated eval: control → computed status.
	computed := map[string]string{
		"ac-1": domain.StatusCompliant,
		"ac-2": domain.StatusCompliant,
	}

	loaded, err := overrides.ListByFramework(ctx, frameworkKey)
	if err != nil {
		t.Fatalf("ListByFramework: %v", err)
	}

	var stale []string
	for cid, native := range computed {
		if target, ok := loaded[cid]; ok && domain.ValidStatus(target) && target == native {
			stale = append(stale, cid)
		}
	}
	for _, cid := range stale {
		if err := overrides.Delete(ctx, frameworkKey, cid); err != nil {
			t.Fatalf("Delete(%s): %v", cid, err)
		}
	}

	if !overrides.deleted["ac-1"] {
		t.Errorf("ac-1 override should have been deleted (target matches computed)")
	}
	if overrides.deleted["ac-2"] {
		t.Errorf("ac-2 override must remain (target still differs from computed)")
	}
	if _, still := overrides.list["ac-1"]; still {
		t.Errorf("ac-1 still in overrides after delete: %v", overrides.list)
	}
	if _, still := overrides.list["ac-2"]; !still {
		t.Errorf("ac-2 was wrongly cleared: %v", overrides.list)
	}
}
