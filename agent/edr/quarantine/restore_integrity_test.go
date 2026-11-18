package quarantine

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

func payloadHash(data string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(data))) }

func restoreFixture(t *testing.T) (*Store, *cache.Cache, string, string, string) {
	t.Helper()
	dir := t.TempDir()
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	s, err := New(filepath.Join(dir, "quarantine"), c)
	if err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(dir, "document.txt")
	if err := os.WriteFile(original, []byte("payload"), 0600); err != nil {
		t.Fatal(err)
	}
	id, err := s.Quarantine(original, payloadHash("payload"), "Safe.Restore.Fixture")
	if err != nil {
		t.Fatal(err)
	}
	return s, c, id, original, filepath.Join(s.dir, id+".quarantined")
}

func TestRestoreRejectsChangedStoredContent(t *testing.T) {
	for _, body := range []string{"altered", "pay", "payload appended", ""} {
		t.Run(fmt.Sprintf("bytes-%d", len(body)), func(t *testing.T) {
			s, c, id, original, stored := restoreFixture(t)
			if err := os.WriteFile(stored, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			if err := s.Restore(id); err == nil {
				t.Fatal("changed quarantine content was restored")
			}
			if _, err := os.Lstat(original); !os.IsNotExist(err) {
				t.Fatal("failed restore published an original path", err)
			}
			if data, err := os.ReadFile(stored); err != nil || string(data) != body {
				t.Fatal("changed source was lost", err)
			}
			if row, found, err := c.GetQuarantine(id); err != nil || !found || row.Restored {
				t.Fatal("failed integrity check changed the record", err)
			}
		})
	}
}

func TestRestoreRejectsUntrustedRecordMetadata(t *testing.T) {
	for _, kind := range []string{"missing-hash", "invalid-hash", "different-hash", "not-restorable", "purged"} {
		t.Run(kind, func(t *testing.T) {
			s, c, id, original, stored := restoreFixture(t)
			r, _, err := c.GetQuarantine(id)
			if err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "missing-hash":
				r.SHA256 = ""
			case "invalid-hash":
				r.SHA256 = strings.Repeat("z", 64)
			case "different-hash":
				r.SHA256 = payloadHash("different body")
			case "not-restorable":
				r.Restorable = false
			case "purged":
				r.Purged = true
			}
			if err := c.StoreQuarantine(r); err != nil {
				t.Fatal(err)
			}
			if err := s.Restore(id); err == nil {
				t.Fatal("untrusted record was restored")
			}
			if _, err := os.Lstat(original); !os.IsNotExist(err) {
				t.Fatal("failed record validation published a destination", err)
			}
			if data, err := os.ReadFile(stored); err != nil || string(data) != "payload" {
				t.Fatal("stored source lost", err)
			}
			if r, _, err := c.GetQuarantine(id); err != nil || r.Restored {
				t.Fatal("rejected record was marked restored", err)
			}
		})
	}
}

func TestRestoreRejectsStoredSymlinksAndDirectories(t *testing.T) {
	for _, kind := range []string{"symlink", "directory"} {
		t.Run(kind, func(t *testing.T) {
			s, _, id, original, stored := restoreFixture(t)
			if err := os.Remove(stored); err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(filepath.Dir(original), "unrelated.txt")
			if err := os.WriteFile(target, []byte("payload"), 0600); err != nil {
				t.Fatal(err)
			}
			if kind == "directory" {
				if err := os.Mkdir(stored, 0700); err != nil {
					t.Fatal(err)
				}
			} else if err := os.Symlink(target, stored); err != nil {
				t.Skipf("symlinks unavailable: %v", err)
			}
			if err := s.Restore(id); err == nil {
				t.Fatal("non-regular quarantine entry was restored")
			}
			if _, err := os.Lstat(original); !os.IsNotExist(err) {
				t.Fatal("non-regular source published a destination", err)
			}
			if _, err := os.Lstat(stored); err != nil {
				t.Fatal("rejected stored entry disappeared", err)
			}
			if data, err := os.ReadFile(target); err != nil || string(data) != "payload" {
				t.Fatal("symlink target changed", err)
			}
		})
	}
}

func TestRestoreAcceptsUppercaseSHA256(t *testing.T) {
	s, c, id, original, _ := restoreFixture(t)
	r, _, err := c.GetQuarantine(id)
	if err != nil {
		t.Fatal(err)
	}
	r.SHA256 = strings.ToUpper(r.SHA256)
	if err := c.StoreQuarantine(r); err != nil {
		t.Fatal(err)
	}
	if err := s.Restore(id); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(original); err != nil || string(data) != "payload" {
		t.Fatal("valid content lost", err)
	}
}

func TestRestoreAndPurgeSerializeAcrossStores(t *testing.T) {
	s, c, id, original, stored := restoreFixture(t)
	other, err := New(s.dir, c)
	if err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	s.restoreBeforePublish = func() error { close(entered); <-release; return nil }
	finished := make(chan error, 1)
	go func() { finished <- s.Restore(id) }()
	select {
	case <-entered:
	case err := <-finished:
		t.Fatalf("restore never reached commit: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("restore stalled")
	}
	// Always release the worker before the database and test directory close.
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	for _, operation := range []func(string) error{other.Restore, other.Purge} {
		if err := operation(id); !errors.Is(err, ErrQuarantineBusy) {
			t.Errorf("competing operation did not honor item lock: %v", err)
		}
	}
	close(release)
	select {
	case err := <-finished:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("restore did not finish")
	}
	if body, err := os.ReadFile(original); err != nil || string(body) != "payload" {
		t.Fatal("winning restore lost payload", err)
	}
	if _, err := os.Stat(stored); !os.IsNotExist(err) {
		t.Fatal("stored input not retired", err)
	}
	if err := other.Purge(id); err != nil {
		t.Fatalf("lock not released: %v", err)
	}
	if body, err := os.ReadFile(original); err != nil || string(body) != "payload" {
		t.Fatal("purge touched restored file", err)
	}
}

func TestQuarantineOperationsRejectNonordinaryIDs(t *testing.T) {
	s, _, _, _, _ := restoreFixture(t)
	for _, id := range []string{"", "../outside", "a/b", `a\b`, "a:b", ".", "..", strings.Repeat("a", 129)} {
		for _, op := range []func(string) error{s.Restore, s.Purge} {
			if err := op(id); err == nil {
				t.Fatalf("invalid ID accepted: %q", id)
			}
		}
	}
}

func TestQuarantinePinsRelativePathsAtAdmission(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	s, err := New("quarantine", c)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("relative.txt", []byte("payload"), 0600); err != nil {
		t.Fatal(err)
	}
	id, err := s.Quarantine("relative.txt", payloadHash("payload"), "Safe.Relative.Fixture")
	if err != nil {
		t.Fatal(err)
	}
	if rec, _, err := c.GetQuarantine(id); err != nil || rec.OriginalPath != filepath.Join(dir, "relative.txt") {
		t.Fatalf("relative path not pinned: %+v %v", rec, err)
	}
	t.Chdir(t.TempDir())
	if err := s.Restore(id); err != nil {
		t.Fatal(err)
	}
	if body, err := os.ReadFile(filepath.Join(dir, "relative.txt")); err != nil || string(body) != "payload" {
		t.Fatal("restore depended on current working directory", err)
	}
}
