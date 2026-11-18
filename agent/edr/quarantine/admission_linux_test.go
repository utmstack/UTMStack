//go:build linux

package quarantine

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestAdmissionNativeCrossVolume(t *testing.T) {
	volume := os.Getenv("OPENEDR_RESTORE_TEST_VOLUME")
	if volume == "" {
		t.Skip("requires an isolated second writable volume")
	}
	s, c, unused, _ := admissionFixture(t)
	_ = os.Remove(unused)
	dir, err := os.MkdirTemp(volume, "admission-volume-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	p := filepath.Join(dir, "source")
	if err = os.WriteFile(p, []byte("cross volume marker"), 0640); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	storeInfo, err := os.Stat(s.dir)
	if err != nil {
		t.Fatal(err)
	}
	if before.Sys().(*syscall.Stat_t).Dev == storeInfo.Sys().(*syscall.Stat_t).Dev {
		t.Fatal("control is not on separate devices")
	}
	id, err := s.QuarantineVerified(p, payloadHash("cross volume marker"), "Safe.CrossVolume.Control", before)
	if err != nil {
		t.Fatal(err)
	}
	r, _, err := c.GetQuarantine(id)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.Stat(filepath.Join(r.SourceCaptureDir, "source"))
	if err != nil || !os.SameFile(before, raw) {
		t.Fatalf("original inode not preserved: %v", err)
	}
	copyInfo, err := os.Stat(filepath.Join(s.dir, id+".quarantined"))
	if err != nil || os.SameFile(raw, copyInfo) {
		t.Fatalf("snapshot is not independent: %v", err)
	}
	if err = s.Restore(id); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(p); err != nil || string(b) != "cross volume marker" {
		t.Fatalf("cross-volume restore: %q %v", b, err)
	}
	if err = s.Purge(id); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(r.SourceCaptureDir); !os.IsNotExist(err) {
		t.Fatalf("retained source not purged: %v", err)
	}
}
