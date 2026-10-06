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

// TestCanaryCaseSensitivity pins the membership-set key semantics: on POSIX
// the key is case-sensitive (a feed path differing only in case is a
// different file), on Windows it lower-cases. The per-platform helper makes
// the test express the host rule directly.
func TestCanaryCaseSensitivity(t *testing.T) {
	planted := "/home/u/docs/00__accounts.xlsx"
	feed := "/home/u/docs/00__Accounts.xlsx"
	a := normCanaryPath(planted)
	b := normCanaryPath(feed)
	// On this host: POSIX → keys differ (feed is NOT the canary);
	// Windows → keys equal (same file, case-insensitive FS).
	if runningOnWindows() {
		if a != b {
			t.Fatalf("Windows: normCanaryPath should be case-insensitive: %q vs %q", a, b)
		}
	} else {
		if a == b {
			t.Fatalf("POSIX: normCanaryPath should be case-sensitive: %q vs %q", a, b)
		}
	}
	// Trailing separator variants of the SAME spelling normalize identically.
	if normCanaryPath(planted+"/") != a {
		t.Fatal("trailing slash should be trimmed")
	}
}
