package engine

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestRetireLegacyUpdaterFiles(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "database")
	if err := os.MkdirAll(filepath.Join(dir, "generations"), 0o700); err != nil {
		t.Fatal(err)
	}
	legacy := []string{"main.cvd", "daily.cld", "bytecode.cvd", "daily-28132.cdiff", "main-63.cvd.sign", "freshclam.dat", "mirrors.dat"}
	kept := []string{"local.ndb", "CURRENT", "approved.hdb", "maincvd.txt"}
	for _, name := range append(append([]string{}, legacy...), kept...) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	moved, err := RetireLegacyUpdaterFiles(dir)
	if err != nil || moved != len(legacy) {
		t.Fatalf("moved %d files, err %v", moved, err)
	}
	remaining, _ := os.ReadDir(dir)
	var names []string
	for _, e := range remaining {
		names = append(names, e.Name())
	}
	want := append([]string{"generations"}, kept...)
	sort.Strings(want)
	sort.Strings(names)
	if len(names) != len(want) {
		t.Fatalf("remaining %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("remaining %v, want %v", names, want)
		}
	}
	targets, _ := filepath.Glob(filepath.Join(parent, "legacy-database-*"))
	if len(targets) != 1 {
		t.Fatalf("legacy directories: %v", targets)
	}
	for _, name := range legacy {
		if b, err := os.ReadFile(filepath.Join(targets[0], name)); err != nil || string(b) != name {
			t.Fatalf("%s not preserved: %v", name, err)
		}
	}
	// Without the retired updater's state file, nothing is touched again.
	if moved, err = RetireLegacyUpdaterFiles(dir); err != nil || moved != 0 {
		t.Fatalf("second run moved %d, err %v", moved, err)
	}
}

func TestRetireLegacyUpdaterFilesNeedsUpdaterState(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.cvd"), []byte("administrator provided"), 0o600); err != nil {
		t.Fatal(err)
	}
	if moved, err := RetireLegacyUpdaterFiles(dir); err != nil || moved != 0 {
		t.Fatalf("moved %d, err %v", moved, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "main.cvd")); err != nil {
		t.Fatal("administrator file moved")
	}
}
