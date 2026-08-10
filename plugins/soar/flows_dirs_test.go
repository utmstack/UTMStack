package main

import (
	"path/filepath"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
)

// The backend seeds the shipped flows and writes each tenant's enabled.yaml
// under SOAR_FLOWS_DIR. Ignoring that variable would put the files the backend
// writes and the files this plugin reads in different places: the UI would
// report a flow as switched on while no alert ever reaches it.
func TestBothTreesAreReadFromWhereTheBackendWritesThem(t *testing.T) {
	t.Setenv("SOAR_FLOWS_DIR", "/somewhere/else/soar")

	systemDir, userRoot := flowDirs()

	if want := "/somewhere/else/soar/system"; systemDir != want {
		t.Errorf("reading shipped flows from %q, backend seeds them to %q", systemDir, want)
	}
	if want := "/somewhere/else/soar/user"; userRoot != want {
		t.Errorf("reading tenant flows from %q, backend writes them to %q", userRoot, want)
	}
}

// Neither side sets the variable in a stock install, so the two agree only if
// this default matches the backend's own default of /workdir/soar.
func TestTheDefaultRootMatchesTheBackends(t *testing.T) {
	systemDir, userRoot := flowDirs()

	if want := filepath.Join(plugins.WorkDir, "soar", "system"); systemDir != want {
		t.Errorf("default shipped-flow root is %q, want %q", systemDir, want)
	}
	if want := filepath.Join(plugins.WorkDir, "soar", "user"); userRoot != want {
		t.Errorf("default tenant root is %q, want %q", userRoot, want)
	}
}

// The two trees carry different authority — one is seeded from the release and
// pruned on every boot, the other is the tenant's own writable space.
func TestTheTwoTreesAreNotTheSameDirectory(t *testing.T) {
	systemDir, userRoot := flowDirs()

	if systemDir == userRoot {
		t.Fatalf("shipped and tenant flows both resolve to %q", systemDir)
	}
}
