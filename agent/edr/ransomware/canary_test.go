package ransomware

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

type fakeCanaryStore struct{ recs []cache.CanaryRecord }

func (f *fakeCanaryStore) StoreCanary(r cache.CanaryRecord) error {
	f.recs = append(f.recs, r)
	return nil
}
func (f *fakeCanaryStore) ListCanaries() ([]cache.CanaryRecord, error) { return f.recs, nil }

func testCanaryManager(t *testing.T, store canaryStore) *Manager {
	t.Helper()
	m := NewManager(store)
	t.Cleanup(func() {
		if err := m.Close(); err != nil {
			t.Errorf("close canary manager: %v", err)
		}
	})
	return m
}

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
	dir := canaryTestDir(t)
	fs := &fakeCanaryStore{}
	m := testCanaryManager(t, fs)
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
	dir := canaryTestDir(t)
	p := filepath.Join(dir, "recorded.xlsx")
	if err := os.WriteFile(p, []byte(canaryBody), 0600); err != nil {
		t.Fatal(err)
	}
	fs := &fakeCanaryStore{recs: []cache.CanaryRecord{{Path: p, SHA256: canaryHash()}}}
	m := testCanaryManager(t, fs)
	if err := m.Load(); err != nil {
		t.Fatal(err)
	}
	if !m.Contains(p) {
		t.Error("Load did not rehydrate membership")
	}
}

func canaryTestDir(t *testing.T) string {
	t.Helper()
	// macOS's temporary directory may itself be reached through /var -> /private.
	// Production intentionally refuses symlink components; fixtures use real paths.
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestCanaryPlantNeverOverwritesOccupiedName(t *testing.T) {
	for _, kind := range []string{"regular", "directory", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			dir := canaryTestDir(t)
			p := filepath.Join(dir, PlanCanaries([]string{dir}, 1)[0].Name)
			target := filepath.Join(canaryTestDir(t), "benign.txt")
			original := []byte("valuable ordinary user file")
			if err := os.WriteFile(target, original, 0600); err != nil {
				t.Fatal(err)
			}
			var err error
			switch kind {
			case "regular":
				err = os.WriteFile(p, original, 0600)
			case "directory":
				err = os.Mkdir(p, 0700)
			case "symlink":
				if err = os.Symlink(target, p); err != nil {
					t.Skipf("symlink fixture unavailable: %v", err)
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			m := testCanaryManager(t, &fakeCanaryStore{})
			if n, err := m.Plant([]string{dir}, 1); err == nil || n != 0 || m.Count() != 0 {
				t.Fatalf("occupied canary adopted: %d %v", n, err)
			}
			if got, err := os.ReadFile(target); err != nil || string(got) != string(original) {
				t.Fatal("symlink target changed", err)
			}
			if kind == "regular" {
				if got, err := os.ReadFile(p); err != nil || string(got) != string(original) {
					t.Fatal("existing benign file changed", err)
				}
			}
		})
	}
}

func TestCanaryParentSymlinkCannotEscapeScope(t *testing.T) {
	dir, outside := canaryTestDir(t), canaryTestDir(t)
	link := filepath.Join(dir, "redirect")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("directory symlink fixture unavailable: %v", err)
	}
	m := testCanaryManager(t, &fakeCanaryStore{})
	if n, err := m.Plant([]string{link}, 1); err == nil || n != 0 {
		t.Fatalf("parent symlink accepted: %d %v", n, err)
	}
	entries, err := os.ReadDir(outside)
	if err != nil || len(entries) != 0 {
		t.Fatal("plant escaped into symlink target", err)
	}
}

func TestCanaryDirectoryHandlePinsCreationDuringPathReplacement(t *testing.T) {
	base, outside := canaryTestDir(t), canaryTestDir(t)
	dir := filepath.Join(base, "pinned")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	d, err := openCanaryDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	if err := os.Rename(dir, dir+".held"); err != nil {
		t.Skipf("platform prevents renaming pinned directory: %v", err)
	}
	if err := os.Symlink(outside, dir); err != nil {
		t.Skipf("directory symlink fixture unavailable: %v", err)
	}
	f, err := d.open("controlled.fixture", true)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	if _, err := os.Stat(filepath.Join(outside, "controlled.fixture")); !os.IsNotExist(err) {
		t.Fatal("pinned directory create followed replaced pathname")
	}
	if _, err := os.Stat(filepath.Join(dir+".held", "controlled.fixture")); err != nil {
		t.Fatal("creation did not remain relative to original directory", err)
	}
}

func TestCanaryExistingAdoptionRequiresRecordedActualHash(t *testing.T) {
	for _, scenario := range []string{"recorded", "unrecorded", "modified", "wrong_hash"} {
		t.Run(scenario, func(t *testing.T) {
			dir := canaryTestDir(t)
			p := filepath.Join(dir, PlanCanaries([]string{dir}, 1)[0].Name)
			body := canaryBody
			if scenario == "modified" {
				body = "valuable replacement user file"
			}
			if err := os.WriteFile(p, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			store := &fakeCanaryStore{}
			if scenario != "unrecorded" {
				hash := canaryHash()
				if scenario == "wrong_hash" {
					hash = "unrecognized"
				}
				store.recs = []cache.CanaryRecord{{Path: p, SHA256: hash}}
			}
			m := testCanaryManager(t, store)
			n, err := m.Plant([]string{dir}, 1)
			if scenario == "recorded" {
				if err != nil || n != 1 || !m.Contains(p) {
					t.Fatalf("owned unchanged canary not adopted: %d %v", n, err)
				}
			} else if err == nil || n != 0 || m.Contains(p) {
				t.Fatalf("unproven existing file adopted: %d %v", n, err)
			}
			if got, err := os.ReadFile(p); err != nil || string(got) != body {
				t.Fatal("adoption modified existing content", err)
			}
		})
	}
}

func TestCanaryConcurrentCreatorsHaveOneExclusiveWinner(t *testing.T) {
	dir := canaryTestDir(t)
	results := make(chan int, 8)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m := testCanaryManager(t, &fakeCanaryStore{})
			n, _ := m.Plant([]string{dir}, 1)
			results <- n
		}()
	}
	wg.Wait()
	close(results)
	winners := 0
	for n := range results {
		winners += n
	}
	if winners != 1 {
		t.Fatalf("expected one exclusive creator, got %d", winners)
	}
}

func TestLoadCoveredCountsOnlyVerifiedCurrentScope(t *testing.T) {
	inside, outside := canaryTestDir(t), canaryTestDir(t)
	store := &fakeCanaryStore{}
	for _, dir := range []string{inside, outside} {
		p := filepath.Join(dir, "recorded.xlsx")
		if err := os.WriteFile(p, []byte(canaryBody), 0600); err != nil {
			t.Fatal(err)
		}
		store.recs = append(store.recs, cache.CanaryRecord{Path: p, SHA256: canaryHash()})
	}
	store.recs = append(store.recs, cache.CanaryRecord{Path: filepath.Join(inside, "missing.xlsx"), SHA256: canaryHash()})
	m := testCanaryManager(t, store)
	if err := m.LoadCovered([]string{inside}); err != nil || m.Count() != 1 || m.Contains(filepath.Join(outside, "recorded.xlsx")) {
		t.Fatalf("scope count %d: %v", m.Count(), err)
	}
	if err := os.WriteFile(filepath.Join(inside, "recorded.xlsx"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.LoadCovered([]string{inside}); err != nil || m.Count() != 0 {
		t.Fatal("stale modified canary remained active", err)
	}
}
