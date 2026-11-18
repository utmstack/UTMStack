//go:build linux

package quarantine

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

func TestPurgeNativeCrossVolume(t *testing.T) {
	volume := os.Getenv("OPENEDR_RESTORE_TEST_VOLUME")
	if volume == "" {
		t.Skip("requires an explicitly supplied isolated second volume")
	}
	for _, restored := range []bool{false, true} {
		dir := t.TempDir()
		originalDir, err := os.MkdirTemp(volume, "openedr-purge-")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.RemoveAll(originalDir) })
		c, err := cache.Open(filepath.Join(dir, "edr.db"))
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		s, err := New(filepath.Join(dir, "quarantine"), c)
		if err != nil {
			t.Fatal(err)
		}
		from, _ := os.Stat(s.dir)
		to, _ := os.Stat(originalDir)
		if from.Sys().(*syscall.Stat_t).Dev == to.Sys().(*syscall.Stat_t).Dev {
			t.Fatal("cross-volume purge control uses the same filesystem")
		}
		original := filepath.Join(originalDir, "original")
		if err = os.WriteFile(original, []byte("safe"), 0600); err != nil {
			t.Fatal(err)
		}
		id, err := s.Quarantine(original, payloadHash("safe"), "Safe.CrossVolume")
		if err != nil {
			t.Fatal(err)
		}
		if restored {
			if err = s.Restore(id); err != nil {
				t.Fatal(err)
			}
		}
		s.responseFault = func(phase string) error {
			if phase == "raw-unlinked" {
				return errors.New("safe interruption on second volume")
			}
			return nil
		}
		if err = s.Purge(id); err == nil {
			t.Fatal("cross-volume checkpoint was not reached")
		}
		s.responseFault = nil
		if _, err = s.Recover(); err != nil {
			t.Fatal(err)
		}
		requirePurgeRecord(t, c, id, true)
		if restored {
			requirePurgePayload(t, original, "safe")
		} else if _, err = os.Lstat(original); !os.IsNotExist(err) {
			t.Fatal("purge published an original path", err)
		}
	}
}

func TestPurgeNativeForeignOwnedUnreadableArtifacts(t *testing.T) {
	if os.Getenv("OPENEDR_PURGE_OWNER_TEST") != "1" {
		t.Skip("requires an isolated root fixture with CHOWN and no DAC read override")
	}
	if os.Geteuid() != 0 {
		t.Fatal("owner fixture must run in its isolated root container")
	}
	s, c, id, _, stored := restoreFixture(t)
	r, _, _ := c.GetQuarantine(id)
	for _, path := range []string{stored, filepath.Join(r.SourceCaptureDir, "source")} {
		if err := os.Chmod(path, 0); err != nil {
			t.Fatal(err)
		}
		if err := os.Chown(path, 65534, 65534); err != nil {
			t.Fatal(err)
		}
		file, err := os.Open(path)
		if file != nil {
			file.Close()
		}
		if !os.IsPermission(err) {
			t.Fatalf("fixture can read data; DAC override was not excluded: %v", err)
		}
	}
	if err := s.Purge(id); err != nil {
		t.Fatalf("metadata-only purge failed without payload read access: %v", err)
	}
	requirePurgeRecord(t, c, id, true)
}
