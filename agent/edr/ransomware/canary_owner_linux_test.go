//go:build linux

package ransomware

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"golang.org/x/sys/unix"
)

func requireLinuxCanaryOwnerFixture(t *testing.T) {
	t.Helper()
	if os.Getenv("OPENEDR_ISOLATED_VM") != "1" {
		t.Skip("ownership changes require explicitly isolated disposable Linux VM")
	}
	if os.Geteuid() != 0 {
		t.Fatal("native ownership fixture requires root")
	}
	b, err := os.ReadFile("/run/openedr-disposable-lab")
	if err != nil || strings.TrimSpace(string(b)) != "OPENEDR_DISPOSABLE_LAB_ONLY" {
		t.Fatal("disposable VM marker missing")
	}
}
func canaryFDState(t *testing.T, f *os.File) canaryOwnerState {
	t.Helper()
	var st unix.Stat_t
	if err := unix.Fstat(int(f.Fd()), &st); err != nil {
		t.Fatal(err)
	}
	return canaryOwnerState{uid: st.Uid, mode: st.Mode}
}
func TestNativeLinuxCanaryOwnerAlignmentAndExistingRefusal(t *testing.T) {
	requireLinuxCanaryOwnerFixture(t)
	const targetUID = 65534
	// Only newly created, private fixture paths are changed. Retain the caller's
	// umask after proving ownership alignment does not widen creation permissions.
	previousMask := unix.Umask(0077)
	defer unix.Umask(previousMask)
	parent := t.TempDir()
	if err := os.Chown(parent, targetUID, -1); err != nil {
		t.Fatal(err)
	}
	store := &fakeCanaryStore{}
	m := testCanaryManager(t, store)
	if n, err := m.Plant([]string{parent}, 1); err != nil || n != 1 {
		t.Fatalf("new owner-aligned planting=%d %v", n, err)
	}
	path := filepath.Join(parent, PlanCanaries([]string{parent}, 1)[0].Name)
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	state := canaryFDState(t, f)
	f.Close()
	if state.uid != targetUID || state.mode&0777 != 0600 {
		t.Fatalf("new UID/umask changed: %+v", state)
	}
	data, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(data, []byte(canaryBody)) {
		t.Fatal("new decoy body changed", err)
	}
	// Existing wrong-UID data is neither adopted nor migrated, even when a stale
	// record claims the expected inert body/hash. It must not inflate coverage.
	legacyDir := t.TempDir()
	if err = os.Chown(legacyDir, targetUID, -1); err != nil {
		t.Fatal(err)
	}
	legacy := filepath.Join(legacyDir, PlanCanaries([]string{legacyDir}, 1)[0].Name)
	if err = os.WriteFile(legacy, []byte(canaryBody), 0644); err != nil {
		t.Fatal(err)
	}
	legacyStore := &fakeCanaryStore{recs: []cache.CanaryRecord{{Path: legacy, SHA256: canaryHash()}}}
	loaded := testCanaryManager(t, legacyStore)
	if err = loaded.LoadCovered([]string{legacyDir}); err != nil {
		t.Fatal(err)
	}
	if loaded.Count() != 0 || loaded.Contains(legacy) {
		t.Fatal("wrong-owner legacy decoy inflated coverage")
	}
	if n, e := loaded.Plant([]string{legacyDir}, 1); e == nil || n != 0 {
		t.Fatalf("wrong-owner existing file admitted: %d %v", n, e)
	}
	f, err = os.Open(legacy)
	if err != nil {
		t.Fatal(err)
	}
	after := canaryFDState(t, f)
	f.Close()
	body, err := os.ReadFile(legacy)
	if err != nil || !bytes.Equal(body, []byte(canaryBody)) || after.uid != 0 || after.mode&0777 != 0600 {
		t.Fatal("existing UID/mode/bytes were changed", after, err)
	}
}
func TestNativeLinuxCanaryParentOwnerRaceAndMetadataFailure(t *testing.T) {
	requireLinuxCanaryOwnerFixture(t)
	parent := t.TempDir()
	if err := os.Chown(parent, 65534, -1); err != nil {
		t.Fatal(err)
	}
	dir, err := openCanaryDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	defer dir.Close()
	path := filepath.Join(parent, "new-inert-race.txt")
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	before, err := canaryDirectoryOwner(dir.file)
	if err != nil {
		t.Fatal(err)
	}
	state := func() (canaryOwnerState, error) {
		var st unix.Stat_t
		if err := unix.Fstat(int(f.Fd()), &st); err != nil {
			return canaryOwnerState{}, err
		}
		return canaryOwnerState{uid: st.Uid, mode: st.Mode}, nil
	}
	err = checkCanaryOwnership(before, true, func() (uint32, error) { return canaryDirectoryOwner(dir.file) }, state, func(uid uint32) error {
		if err := unix.Fchown(int(f.Fd()), int(uid), -1); err != nil {
			return err
		}
		// A controlled exact-descriptor directory-owner change at the race boundary.
		return unix.Fchown(int(dir.file.Fd()), 65533, -1)
	})
	if err == nil {
		t.Fatal("changed pinned directory owner admitted")
	}
	if got := canaryFDState(t, f); got.uid != 65534 {
		t.Fatal("operation rebound to another inode", got)
	}
	closed, err := openCanaryDir(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	closed.Close()
	if got, e := closed.open("must-not-create.txt", true); e == nil {
		got.Close()
		t.Fatal("closed parent descriptor admitted")
	} else if !errors.Is(e, unix.EBADF) {
		t.Fatalf("metadata error lost: %v", e)
	}
}
