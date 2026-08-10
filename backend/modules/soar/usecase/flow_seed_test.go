package usecase

import (
	"os"
	"path/filepath"
	"testing"
)

func seedDirs(t *testing.T) (src string, store *FlowStore) {
	t.Helper()
	root := t.TempDir()
	src = filepath.Join(root, "image")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	return src, NewFlowStore(filepath.Join(root, "soar", SystemSubdir), filepath.Join(root, "soar", UserSubdir))
}

func writeShipped(t *testing.T, dir, rel, body string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// The plugin evaluating the flows cannot reach the backend's image. A shipped
// flow that never lands on the shared volume is one a tenant can switch on and
// never see fire.
func TestAShippedFlowReachesTheSharedVolume(t *testing.T) {
	src, store := seedDirs(t)
	writeShipped(t, src, "isolate-host.yaml", "- name: isolate\n")

	if err := store.SeedSystem(src); err != nil {
		t.Fatalf("seed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(store.systemDir, "isolate-host.yaml"))
	if err != nil {
		t.Fatalf("the plugin will not find the flow: %v", err)
	}
	if string(got) != "- name: isolate\n" {
		t.Errorf("seeded contents are %q", got)
	}
}

// An upgrade must land: the tenant enabled a flow by path, and the release
// decides what that path does.
func TestAnEditedFlowIsRefreshedOnTheNextBoot(t *testing.T) {
	src, store := seedDirs(t)
	writeShipped(t, src, "kill.yaml", "- name: old\n")
	if err := store.SeedSystem(src); err != nil {
		t.Fatal(err)
	}

	writeShipped(t, src, "kill.yaml", "- name: new\n")
	if err := store.SeedSystem(src); err != nil {
		t.Fatal(err)
	}

	got, _ := os.ReadFile(filepath.Join(store.systemDir, "kill.yaml"))
	if string(got) != "- name: new\n" {
		t.Errorf("overlay still holds %q after the upgrade", got)
	}
}

// A flow dropped from the release must not keep running commands on people's
// machines under a name the product no longer ships.
func TestAFlowDroppedFromTheReleaseIsRemoved(t *testing.T) {
	src, store := seedDirs(t)
	writeShipped(t, src, "retired.yaml", "- name: retired\n")
	writeShipped(t, src, "kept.yaml", "- name: kept\n")
	if err := store.SeedSystem(src); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(filepath.Join(src, "retired.yaml")); err != nil {
		t.Fatal(err)
	}
	if err := store.SeedSystem(src); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(store.systemDir, "retired.yaml")); !os.IsNotExist(err) {
		t.Error("a flow the release no longer ships is still on the volume")
	}
	if _, err := os.Stat(filepath.Join(store.systemDir, "kept.yaml")); err != nil {
		t.Error("pruning removed a flow that is still shipped")
	}
}

// Seeding rewrites the overlay from the release on every boot, so editing those
// files on the host is not a way to change what the system flows do.
func TestTamperingWithTheOverlayIsUndone(t *testing.T) {
	src, store := seedDirs(t)
	writeShipped(t, src, "shutdown.yaml", "- name: shutdown\n")
	if err := store.SeedSystem(src); err != nil {
		t.Fatal(err)
	}

	tampered := filepath.Join(store.systemDir, "shutdown.yaml")
	if err := os.WriteFile(tampered, []byte("- name: tampered\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := store.SeedSystem(src); err != nil {
		t.Fatal(err)
	}

	got, _ := os.ReadFile(tampered)
	if string(got) != "- name: shutdown\n" {
		t.Errorf("the edit survived the reseed: %q", got)
	}
}

// The tenants' own flows and their enabled.yaml share the volume; pruning walks
// the system tree only and must not reach them.
func TestSeedingLeavesTheTenantsFilesAlone(t *testing.T) {
	src, store := seedDirs(t)
	tenantDir := filepath.Join(store.userDir, "8f1c1b8e-0000-4000-8000-00000000000a")
	writeShipped(t, tenantDir, "mine.yaml", "- name: mine\n")
	writeShipped(t, tenantDir, EnabledFileName, "enabled:\n    - mine.yaml\n")

	if err := store.SeedSystem(src); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(tenantDir, "mine.yaml")); err != nil {
		t.Error("seeding deleted a tenant's own flow")
	}
	if _, err := os.Stat(filepath.Join(tenantDir, EnabledFileName)); err != nil {
		t.Error("seeding deleted a tenant's enabled list")
	}
}

// A first boot on a clean install has no overlay yet.
func TestSeedingCreatesTheOverlayOnAFreshInstall(t *testing.T) {
	src, store := seedDirs(t)
	writeShipped(t, src, "nested/deep.yaml", "- name: deep\n")

	if err := store.SeedSystem(src); err != nil {
		t.Fatalf("seed on a fresh install: %v", err)
	}

	if _, err := os.Stat(filepath.Join(store.systemDir, "nested", "deep.yaml")); err != nil {
		t.Errorf("nested flow not seeded: %v", err)
	}
}

// A backend built without the definitions must still start.
func TestAMissingSourceIsNotAStartupFailure(t *testing.T) {
	_, store := seedDirs(t)

	if err := store.SeedSystem("/nonexistent/definitions/soar"); err != nil {
		t.Errorf("a missing source dir stopped the backend: %v", err)
	}
}
