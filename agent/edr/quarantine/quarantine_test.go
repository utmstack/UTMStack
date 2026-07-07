package quarantine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

func TestQuarantineMovesFileAndRecords(t *testing.T) {
	dir := t.TempDir()
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	store, err := New(filepath.Join(dir, "quarantine"), c)
	if err != nil {
		t.Fatal(err)
	}

	src := filepath.Join(dir, "a.exe")
	if err := os.WriteFile(src, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	id, err := store.Quarantine(src, "hash", "Win.Test.EICAR")
	if err != nil {
		t.Fatalf("Quarantine: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatal("original file should be gone after quarantine")
	}
	rec, found, _ := c.GetQuarantine(id)
	if !found || rec.OriginalPath != src || rec.Engine != "UTMStack EDR" {
		t.Fatalf("record wrong: %+v found=%v", rec, found)
	}

	// The quarantined copy must preserve the original bytes (move, not delete).
	stored := filepath.Join(dir, "quarantine", id+".quarantined")
	if b, err := os.ReadFile(stored); err != nil || string(b) != "payload" {
		t.Fatalf("quarantined copy content = %q err=%v, want original bytes", string(b), err)
	}

	if err := store.Restore(id); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if b, err := os.ReadFile(src); err != nil || string(b) != "payload" {
		t.Fatalf("restore did not put the original bytes back: %q err=%v", string(b), err)
	}
	if _, err := os.Stat(stored); !os.IsNotExist(err) {
		t.Fatal("quarantined copy should be gone after restore")
	}
	// Re-restoring an already-restored id must fail cleanly, not corrupt anything.
	if err := store.Restore(id); err == nil {
		t.Fatal("re-restore of an already-restored id should error")
	} else if !strings.Contains(err.Error(), "already restored") {
		t.Fatalf("re-restore error = %q, want 'already restored'", err)
	}
}

func TestPurgeAndPurgeExpired(t *testing.T) {
	dir := t.TempDir()
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	store, err := New(filepath.Join(dir, "quarantine"), c)
	if err != nil {
		t.Fatal(err)
	}

	// Two quarantined items.
	mk := func(name string) string {
		src := filepath.Join(dir, name)
		_ = os.WriteFile(src, []byte("x"), 0o644)
		id, err := store.Quarantine(src, "h", "Win.Test")
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	idA := mk("a.exe")
	idB := mk("b.exe")

	// Manual purge of A: stored file gone, record marked purged, re-purge errors.
	if err := store.Purge(idA); err != nil {
		t.Fatalf("purge: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "quarantine", idA+".quarantined")); !os.IsNotExist(err) {
		t.Fatal("purged file should be deleted")
	}
	rec, _, _ := c.GetQuarantine(idA)
	if !rec.Purged {
		t.Fatal("record not marked purged")
	}
	if err := store.Purge(idA); err == nil {
		t.Fatal("re-purge should error")
	}

	// PurgeExpired with a 0-day retention is a no-op (keep forever semantics).
	if n, _ := store.PurgeExpired(0, time.Now()); n != 0 {
		t.Fatalf("retention 0 must be a no-op, purged %d", n)
	}
	// Back-date B and sweep with a 1-day retention → B is purged, A already was.
	if err := c.StoreQuarantine(cache.QuarantineRecord{
		QuarantineID: idB, OriginalPath: filepath.Join(dir, "b.exe"), Engine: "UTMStack EDR",
		QuarantinedAt: time.Now().Add(-48 * time.Hour),
	}); err != nil {
		t.Fatal(err)
	}
	n, err := store.PurgeExpired(1, time.Now())
	if err != nil || n != 1 {
		t.Fatalf("PurgeExpired = %d, %v; want 1 purged", n, err)
	}
}
