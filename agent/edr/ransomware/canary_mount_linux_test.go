//go:build linux

package ransomware

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"golang.org/x/sys/unix"
)

func TestLinuxCanaryFailedMountScopeDoesNotRemainUnrestricted(t *testing.T) {
	m := testCanaryManager(t, &fakeCanaryStore{})
	if err := m.SetMountScope(nil); err == nil {
		t.Fatal("empty mount scope accepted")
	}
	if n, err := m.Plant([]string{canaryTestDir(t)}, 1); err == nil || n != 0 {
		t.Fatal("failed scope allowed unrestricted creation")
	}
}

func TestNativeCanaryMountScopeRejectsUnmonitoredFilesystem(t *testing.T) {
	root := os.Getenv("OPENEDR_FANOTIFY_TEST_ROOT")
	if os.Getenv("OPENEDR_ISOLATED_VM") != "1" || root == "" {
		t.Skip("requires explicit isolated-VM tmpfs fixture")
	}
	if os.Geteuid() != 0 || !strings.HasPrefix(root, "/tmp/openedr-fanotify-") {
		t.Fatal("requires dedicated owned root tmpfs fixture")
	}
	var fs unix.Statfs_t
	if err := unix.Statfs(root, &fs); err != nil || uint64(fs.Type) != uint64(unix.TMPFS_MAGIC) {
		t.Fatal("fixture must be dedicated tmpfs", err)
	}
	inside, err := os.MkdirTemp(root, "canary-scope-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(inside)
	outside := canaryTestDir(t)
	store := &fakeCanaryStore{}
	m := testCanaryManager(t, store)
	if err := m.SetMountScope([]string{root}); err != nil {
		t.Fatal(err)
	}
	if n, err := m.Plant([]string{inside}, 1); err != nil || n != 1 {
		t.Fatalf("declared mount canary failed: %d %v", n, err)
	}
	if n, err := m.Plant([]string{outside}, 1); err == nil || n != 0 {
		t.Fatalf("unmonitored mount accepted: %d %v", n, err)
	}
	p := filepath.Join(outside, "recorded.xlsx")
	if err := os.WriteFile(p, []byte(canaryBody), 0600); err != nil {
		t.Fatal(err)
	}
	store.recs = append(store.recs, cache.CanaryRecord{Path: p, SHA256: canaryHash()})
	if err := m.LoadCovered([]string{inside, outside}); err != nil || m.Count() != 1 || m.Contains(p) {
		t.Fatalf("unmonitored stored canary counted: %d %v", m.Count(), err)
	}
}
