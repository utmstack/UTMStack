package docker

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/utmstack/UTMStack/installer/config"
)

func TestComposeFilePathLocation(t *testing.T) {
	cfg := config.GetConfig()
	want := filepath.Join(cfg.DataDir, "compose.yml")
	if config.ComposeFilePath != want {
		t.Fatalf("ComposeFilePath = %q, want %q", config.ComposeFilePath, want)
	}

	binDir := filepath.Dir(config.InstallerBinPath)
	if filepath.Dir(config.ComposeFilePath) == binDir {
		t.Fatalf("ComposeFilePath %q must not live in binary directory %q", config.ComposeFilePath, binDir)
	}

	if filepath.Dir(config.ComposeFilePath) == filepath.Clean(cfg.UpdatesFolder) {
		t.Fatalf("ComposeFilePath %q must not live in UpdatesFolder %q (mounted into containers)", config.ComposeFilePath, cfg.UpdatesFolder)
	}
}

func TestWriteComposeFileCreatesFileAndRemovesLegacy(t *testing.T) {
	tmp := t.TempDir()
	composePath := filepath.Join(tmp, "data", "compose.yml")
	legacyPath := filepath.Join(tmp, "bin", "compose.yml")

	if err := os.MkdirAll(filepath.Dir(legacyPath), 0755); err != nil {
		t.Fatalf("mkdir legacy dir: %v", err)
	}
	if err := os.WriteFile(legacyPath, []byte("legacy: true\n"), 0644); err != nil {
		t.Fatalf("seed legacy compose: %v", err)
	}

	want := "services:\n  backend:\n    image: test\n"
	if err := writeComposeFile(composePath, []byte(want), legacyPath); err != nil {
		t.Fatalf("writeComposeFile: %v", err)
	}

	got, err := os.ReadFile(composePath)
	if err != nil {
		t.Fatalf("read written compose: %v", err)
	}
	if string(got) != want {
		t.Fatalf("compose content = %q, want %q", string(got), want)
	}

	if runtime.GOOS != "windows" {
		fi, err := os.Stat(composePath)
		if err != nil {
			t.Fatalf("stat compose: %v", err)
		}
		if fi.Mode().Perm() != 0600 {
			t.Fatalf("compose permissions = %o, want 0600", fi.Mode().Perm())
		}
	}

	if _, err := os.Stat(legacyPath); !os.IsNotExist(err) {
		t.Fatalf("legacy compose file %q should have been removed, stat err = %v", legacyPath, err)
	}
}

func TestWriteComposeFileNoLegacyAndSamePathGuard(t *testing.T) {
	tmp := t.TempDir()
	composePath := filepath.Join(tmp, "compose.yml")
	missingLegacy := filepath.Join(tmp, "nonexistent", "compose.yml")

	payload := []byte("services: {}\n")
	if err := writeComposeFile(composePath, payload, missingLegacy); err != nil {
		t.Fatalf("writeComposeFile with missing legacy: %v", err)
	}

	// Passing the same path as both target and legacy must never delete the target.
	if err := writeComposeFile(composePath, payload, composePath); err != nil {
		t.Fatalf("writeComposeFile with same path: %v", err)
	}
	if _, err := os.Stat(composePath); err != nil {
		t.Fatalf("compose file should still exist when legacyPath == composePath: %v", err)
	}
}
