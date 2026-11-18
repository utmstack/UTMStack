//go:build linux

package quarantine

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// These controls require two disposable writable volumes, never a host path.
func TestRestoreNativeCrossVolume(t *testing.T) {
	destination := os.Getenv("OPENEDR_RESTORE_TEST_VOLUME")
	if destination == "" {
		t.Skip("requires an explicitly supplied isolated second volume")
	}
	for _, kind := range []string{"verified", "occupied", "tampered"} {
		t.Run(kind, func(t *testing.T) {
			s, c, id, original, stored := restoreFixture(t)
			dir, err := os.MkdirTemp(destination, "restore-test-")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.RemoveAll(dir) })
			from, err := os.Stat(stored)
			if err != nil {
				t.Fatal(err)
			}
			to, err := os.Stat(dir)
			if err != nil {
				t.Fatal(err)
			}
			if from.Sys().(*syscall.Stat_t).Dev == to.Sys().(*syscall.Stat_t).Dev {
				t.Fatal("cross-volume control is on the same device")
			}
			rec, _, err := c.GetQuarantine(id)
			if err != nil {
				t.Fatal(err)
			}
			rec.OriginalPath = filepath.Join(dir, "restored.txt")
			if err := c.StoreQuarantine(rec); err != nil {
				t.Fatal(err)
			}
			if kind == "occupied" {
				if err := os.WriteFile(rec.OriginalPath, []byte("benign"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if kind == "tampered" {
				if err := os.WriteFile(stored, []byte("altered"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			err = s.Restore(id)
			if kind == "verified" {
				if err != nil {
					t.Fatal(err)
				}
				if body, e := os.ReadFile(rec.OriginalPath); e != nil || string(body) != "payload" {
					t.Fatal("cross-volume payload lost", e)
				}
				if _, e := os.Stat(stored); !os.IsNotExist(e) {
					t.Fatal("verified source not retired", e)
				}
			} else {
				if err == nil {
					t.Fatal("unsafe cross-volume restore succeeded")
				}
				if _, e := os.Stat(stored); e != nil {
					t.Fatal("rejected source lost", e)
				}
				if kind == "occupied" {
					if body, e := os.ReadFile(rec.OriginalPath); e != nil || string(body) != "benign" {
						t.Fatal("destination overwritten", e)
					}
				} else if _, e := os.Lstat(rec.OriginalPath); !os.IsNotExist(e) {
					t.Fatal("tampered output published", e)
				}
			}
			noRestoreStages(t, s.dir, filepath.Dir(original), dir)
		})
	}
}

func TestRestoreNativePreservesStoredOwner(t *testing.T) {
	if os.Getenv("OPENEDR_RESTORE_OWNER_TEST") != "1" {
		t.Skip("requires an isolated root fixture with CHOWN and DAC_OVERRIDE")
	}
	if os.Geteuid() != 0 {
		t.Fatal("owner control requires isolated root")
	}
	s, _, id, original, stored := restoreFixture(t)
	if err := os.Chown(stored, 65534, 65534); err != nil {
		t.Fatal(err)
	}
	if err := s.Restore(id); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(original)
	if err != nil {
		t.Fatal(err)
	}
	st := info.Sys().(*syscall.Stat_t)
	if st.Uid != 65534 || st.Gid != 65534 || info.Mode().Perm() != 0600 {
		t.Fatalf("stored ownership/privacy not preserved: %d:%d %v", st.Uid, st.Gid, info.Mode())
	}
}
