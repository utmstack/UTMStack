package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestStatusFileAtomicallyReplacesSymlink(t *testing.T) {
	dir := t.TempDir()
	old := config.StatusFile
	config.StatusFile = filepath.Join(dir, "status.json")
	t.Cleanup(func() { config.StatusFile = old })
	victim := filepath.Join(dir, "victim")
	if err := os.WriteFile(victim, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(victim, config.StatusFile); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := writeStatusFile([]byte(`{"running":true}`)); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{victim: "keep", config.StatusFile: `{"running":true}`} {
		got, err := os.ReadFile(path)
		if err != nil || string(got) != want {
			t.Fatalf("%s = %q, %v", path, got, err)
		}
	}
	info, err := os.Lstat(config.StatusFile)
	if err != nil || !info.Mode().IsRegular() {
		t.Fatalf("status not regular after replacement: %v %v", info, err)
	}
	temps, err := filepath.Glob(filepath.Join(dir, ".edr-status-*"))
	if err != nil || len(temps) != 0 {
		t.Fatalf("temporary status files retained: %v %v", temps, err)
	}
}
