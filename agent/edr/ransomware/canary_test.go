package ransomware

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

type fakeCanaryStore struct{ recs []cache.CanaryRecord }

func (f *fakeCanaryStore) StoreCanary(r cache.CanaryRecord) error {
	f.recs = append(f.recs, r)
	return nil
}
func (f *fakeCanaryStore) ListCanaries() ([]cache.CanaryRecord, error) { return f.recs, nil }

func TestPlanCanaries_NamesSortEarly(t *testing.T) {
	specs := PlanCanaries([]string{`C:\Users\x\Documents`}, 2)
	if len(specs) != 2 {
		t.Fatalf("got %d specs", len(specs))
	}
	for _, s := range specs {
		if s.Name[0] != '0' {
			t.Errorf("canary %q should sort early (lead with a digit)", s.Name)
		}
	}
	if specs[0].Name == specs[1].Name {
		t.Error("canary names must be distinct per dir")
	}
}

func TestManager_PlantPersistsAndMatches(t *testing.T) {
	dir := t.TempDir()
	fs := &fakeCanaryStore{}
	m := NewManager(fs)
	n, err := m.Plant([]string{dir}, 2)
	if err != nil || n != 2 {
		t.Fatalf("plant = %d, %v", n, err)
	}
	// Files exist on disk, are non-empty, and registered.
	entries, _ := os.ReadDir(dir)
	if len(entries) != 2 {
		t.Fatalf("expected 2 canary files, got %d", len(entries))
	}
	p := filepath.Join(dir, entries[0].Name())
	if !m.Contains(p) {
		t.Errorf("planted canary %q not recognized", p)
	}
	// Case-insensitive / separator-agnostic membership.
	if !m.Contains(filepath.ToSlash(p)) {
		t.Error("membership must be separator-agnostic")
	}
	if len(fs.recs) != 2 {
		t.Errorf("canaries not persisted: %d", len(fs.recs))
	}
}

func TestManager_LoadRehydratesMembership(t *testing.T) {
	fs := &fakeCanaryStore{recs: []cache.CanaryRecord{{Path: `C:\Users\x\00__a.xlsx`}}}
	m := NewManager(fs)
	if err := m.Load(); err != nil {
		t.Fatal(err)
	}
	if !m.Contains(`c:\users\x\00__A.xlsx`) {
		t.Error("Load did not rehydrate membership")
	}
}
