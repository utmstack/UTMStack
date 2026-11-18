//go:build linux || darwin

package quarantine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRestoreJournalPreservesNonUTF8Destination(t *testing.T) {
	s, c, id, original, stored := restoreFixture(t)
	r, _, err := c.GetQuarantine(id)
	if err != nil {
		t.Fatal(err)
	}
	parent := filepath.Join(filepath.Dir(original), string([]byte{'d', 0xff, 0x80}))
	if err = os.Mkdir(parent, 0700); err != nil {
		t.Skipf("filesystem does not support byte filename: %v", err)
	}
	r.OriginalPath = filepath.Join(parent, string([]byte{'f', 0xfe, 0x81}))
	if err = c.StoreQuarantine(r); err != nil {
		t.Fatal(err)
	}
	s.responseFault = func(p string) error {
		if p == "restore-published-fs" {
			return errors.New("safe interruption")
		}
		return nil
	}
	if err = s.Restore(id); err == nil {
		t.Fatal("checkpoint not reached")
	}
	s.responseFault = nil
	m := restoreFixtureManifest(t, s, id)
	if string(m.ParentPath) != parent || m.Record.OriginalPath != r.OriginalPath {
		t.Fatal("journal lost destination bytes")
	}
	results, err := s.Recover()
	if err != nil || len(results) != 1 || results[0].Status != "restored" || results[0].OriginalPath != r.OriginalPath {
		t.Fatalf("byte path not recovered: %+v %v", results, err)
	}
	requirePurgePayload(t, r.OriginalPath, "payload")
	if _, err = os.Lstat(stored); !os.IsNotExist(err) {
		t.Fatal("backup not retired", err)
	}
}
