//go:build windows && edr_acceptance

package quarantine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"golang.org/x/sys/windows"
)

func windowsCrossVolume(t *testing.T, store string) string {
	t.Helper()
	volume := os.Getenv("OPENEDR_RESTORE_TEST_VOLUME")
	if os.Getenv("EDR_WINDOWS_ACCEPTANCE") != "isolated-vm" || volume == "" {
		t.Skip("requires a disposable Windows guest and an explicit second NTFS volume")
	}
	serial := func(path string) uint32 {
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		var info windows.ByHandleFileInformation
		if err := windows.GetFileInformationByHandle(windows.Handle(f.Fd()), &info); err != nil {
			t.Fatal(err)
		}
		return info.VolumeSerialNumber
	}
	from, to := serial(store), serial(volume)
	if from == to {
		t.Fatal("cross-volume fixture uses the same NTFS volume")
	}
	t.Logf("quarantine_volume=%08x second_volume=%08x", from, to)
	dir, err := os.MkdirTemp(volume, "openedr-owned-cross-volume-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

func TestWindowsAcceptanceAdmissionCrossVolume(t *testing.T) {
	s, c, unused, _ := admissionFixture(t)
	_ = os.Remove(unused)
	dir := windowsCrossVolume(t, s.dir)
	p := filepath.Join(dir, "source.txt")
	const payload = "cross volume harmless marker"
	if err := os.WriteFile(p, []byte(payload), 0600); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	id, err := s.QuarantineVerified(p, payloadHash(payload), "Safe.CrossVolume.Control", before)
	if err != nil {
		t.Fatal(err)
	}
	r, found, err := c.GetQuarantine(id)
	if err != nil || !found {
		t.Fatal("missing admission record", err)
	}
	raw, err := os.Stat(filepath.Join(r.SourceCaptureDir, "source"))
	if err != nil || !os.SameFile(before, raw) {
		t.Fatal("original identity lost", err)
	}
	snapshot, err := os.Stat(filepath.Join(s.dir, id+".quarantined"))
	if err != nil || os.SameFile(raw, snapshot) {
		t.Fatal("snapshot is not independent", err)
	}
	if err := s.Restore(id); err != nil {
		t.Fatal(err)
	}
	requirePurgePayload(t, p, payload)
	if err := s.Purge(id); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(r.SourceCaptureDir); !os.IsNotExist(err) {
		t.Fatal("capture retained after purge", err)
	}
}

func TestWindowsAcceptanceRestoreCrossVolume(t *testing.T) {
	for _, kind := range []string{"verified", "occupied", "tampered"} {
		t.Run(kind, func(t *testing.T) {
			s, c, id, original, stored := restoreFixture(t)
			dir := windowsCrossVolume(t, s.dir)
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
				requirePurgePayload(t, rec.OriginalPath, "payload")
				if _, err := os.Stat(stored); !os.IsNotExist(err) {
					t.Fatal("source not retired", err)
				}
			} else {
				if err == nil {
					t.Fatal("unsafe restore succeeded")
				}
				if _, err := os.Stat(stored); err != nil {
					t.Fatal("rejected source lost", err)
				}
				if kind == "occupied" {
					requirePurgePayload(t, rec.OriginalPath, "benign")
				} else if _, err := os.Lstat(rec.OriginalPath); !os.IsNotExist(err) {
					t.Fatal("tampered output published", err)
				}
			}
			for _, path := range []string{s.dir, filepath.Dir(original), dir} {
				entries, err := os.ReadDir(path)
				if err != nil {
					t.Fatal(err)
				}
				for _, entry := range entries {
					if strings.HasPrefix(entry.Name(), ".edr-restore-") {
						t.Fatalf("restore stage remains: %s", filepath.Join(path, entry.Name()))
					}
				}
			}
		})
	}
}

func TestWindowsAcceptancePurgeCrossVolumeRecovery(t *testing.T) {
	for _, restored := range []bool{false, true} {
		t.Run(map[bool]string{false: "quarantined", true: "restored"}[restored], func(t *testing.T) {
			c, err := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			s, err := New(filepath.Join(t.TempDir(), "quarantine"), c)
			if err != nil {
				t.Fatal(err)
			}
			dir := windowsCrossVolume(t, s.dir)
			original := filepath.Join(dir, "original.txt")
			if err := os.WriteFile(original, []byte("safe"), 0600); err != nil {
				t.Fatal(err)
			}
			id, err := s.Quarantine(original, payloadHash("safe"), "Safe.CrossVolume")
			if err != nil {
				t.Fatal(err)
			}
			if restored {
				if err := s.Restore(id); err != nil {
					t.Fatal(err)
				}
			}
			s.responseFault = func(phase string) error {
				if phase == "raw-unlinked" {
					return errors.New("owned second-volume checkpoint")
				}
				return nil
			}
			if err := s.Purge(id); err == nil {
				t.Fatal("purge checkpoint not reached")
			}
			s.responseFault = nil
			if _, err := s.Recover(); err != nil {
				t.Fatal(err)
			}
			requirePurgeRecord(t, c, id, true)
			if restored {
				requirePurgePayload(t, original, "safe")
			} else if _, err := os.Lstat(original); !os.IsNotExist(err) {
				t.Fatal("purge published original path", err)
			}
		})
	}
}
