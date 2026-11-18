package quarantine

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

var purgeCheckpoints = []string{
	"purge-directory-created", "purge-intent", "stored-claim-intent", "stored-claimed-fs",
	"stored-claimed", "stored-delete-intent", "stored-unlinked", "stored-deleted",
	"raw-claim-intent", "raw-claimed-fs", "raw-claimed", "raw-delete-intent", "raw-unlinked", "raw-deleted",
	"capture-delete-intent", "capture-unlinked", "capture-deleted", "before-database-commit",
	"database-committed", "committed", "journal-cleanup",
}

func requirePurgePayload(t *testing.T, path, want string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil || string(b) != want {
		t.Fatalf("preserved payload %q: %q %v", path, b, err)
	}
}

func requirePurgeRecord(t *testing.T, c *cache.Cache, id string, purged bool) cache.QuarantineRecord {
	t.Helper()
	r, found, err := c.GetQuarantine(id)
	if err != nil || !found || r.Purged != purged || purged == r.PurgedAt.IsZero() {
		t.Fatalf("purge audit state: %+v %t %v", r, found, err)
	}
	return r
}

func TestPurgeRecoveryAtEveryBoundary(t *testing.T) {
	for _, checkpoint := range purgeCheckpoints {
		t.Run(checkpoint, func(t *testing.T) {
			s, c, id, original, stored := restoreFixture(t)
			before, _, _ := c.GetQuarantine(id)
			fired := false
			fault := errors.New("safe interruption")
			s.responseFault = func(phase string) error {
				if phase == checkpoint {
					fired = true
					return fault
				}
				return nil
			}
			err := s.Purge(id)
			var pending *PurgePendingError
			if !fired || !errors.As(err, &pending) || !errors.Is(err, fault) {
				t.Fatalf("missing durable interruption: fired=%t error=%v", fired, err)
			}
			if err = s.Restore(id); !errors.As(err, &pending) {
				t.Fatalf("restore bypassed the pending purge: %v", err)
			}
			s.responseFault = nil
			results, err := s.Recover()
			if err != nil || len(results) != 1 || results[0].Operation != "purge" || results[0].Status == "needs_attention" {
				t.Fatalf("purge replay: %+v %v", results, err)
			}
			purged := checkpoint != "purge-directory-created"
			r := requirePurgeRecord(t, c, id, purged)
			if purged {
				for _, path := range []string{stored, before.SourceCaptureDir} {
					if _, e := os.Lstat(path); !os.IsNotExist(e) {
						t.Fatalf("approved artifact remains: %s %v", path, e)
					}
				}
			} else {
				requirePurgePayload(t, stored, "payload")
				requirePurgePayload(t, filepath.Join(before.SourceCaptureDir, "source"), "payload")
			}
			if _, e := os.Lstat(original); !os.IsNotExist(e) {
				t.Fatal("purge published an original path", e)
			}
			if results, err = s.Recover(); err != nil || len(results) != 0 {
				t.Fatalf("purge replay was not idempotent: %+v %v", results, err)
			}
			after, _, _ := c.GetQuarantine(id)
			if !after.PurgedAt.Equal(r.PurgedAt) {
				t.Fatal("repeated recovery changed purge completion time")
			}
		})
	}
}

func TestPurgeRecoveryPreservesReoccupiedStoredName(t *testing.T) {
	for _, checkpoint := range []string{"stored-claim-intent", "stored-claimed", "stored-unlinked", "stored-deleted", "database-committed"} {
		t.Run(checkpoint, func(t *testing.T) {
			s, c, id, _, stored := restoreFixture(t)
			s.responseFault = func(phase string) error {
				if phase == checkpoint {
					return errors.New("safe interruption")
				}
				return nil
			}
			if err := s.Purge(id); err == nil {
				t.Fatal("interruption did not stop purge")
			}
			originalSaved := stored + ".saved"
			if checkpoint == "stored-claim-intent" {
				if err := os.Rename(stored, originalSaved); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(stored, []byte("unrelated replacement"), 0600); err != nil {
				t.Fatal(err)
			}
			s.responseFault = nil
			if results, err := s.Recover(); err == nil || len(results) != 1 || results[0].Status != "needs_attention" {
				t.Fatalf("replacement was not retained for attention: %+v %v", results, err)
			}
			requirePurgePayload(t, stored, "unrelated replacement")
			requirePurgeRecord(t, c, id, checkpoint == "database-committed")
			preserved := filepath.Join(filepath.Dir(s.dir), "benign-preserved")
			if err := os.Rename(stored, preserved); err != nil {
				t.Fatal(err)
			}
			if checkpoint == "stored-claim-intent" {
				if err := os.Rename(originalSaved, stored); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := s.Recover(); err != nil {
				t.Fatalf("authorized repair could not finish: %v", err)
			}
			requirePurgePayload(t, preserved, "unrelated replacement")
			requirePurgeRecord(t, c, id, true)
		})
	}
}

func TestPurgeMetadataFailureRetainsReplayAndRetentionError(t *testing.T) {
	s, c, id, _, stored := restoreFixture(t)
	dbPath := filepath.Join(filepath.Dir(s.dir), "edr.db")
	r, _, _ := c.GetQuarantine(id)
	r.QuarantinedAt = time.Now().Add(-48 * time.Hour)
	if err := c.StoreQuarantine(r); err != nil {
		t.Fatal(err)
	}
	s.responseFault = func(phase string) error {
		if phase == "stored-deleted" {
			return c.Close()
		}
		return nil
	}
	n, err := s.PurgeExpired(1, time.Now())
	var pending *PurgePendingError
	if n != 0 || !errors.As(err, &pending) {
		t.Fatalf("retention swallowed partial purge failure: %d %v", n, err)
	}
	if _, err = os.Lstat(stored); !os.IsNotExist(err) {
		t.Fatal("fixture never reached filesystem removal", err)
	}
	reopened, err := cache.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	recovered, err := New(s.dir, reopened)
	if err != nil || len(recovered.RecoveryStatus()) != 1 || recovered.RecoveryStatus()[0].Status != "purged" {
		t.Fatalf("startup did not finish approved purge: %v", err)
	}
	requirePurgeRecord(t, reopened, id, true)
}

func TestPurgeRejectsUnknownCaptureContentsBeforeDeleting(t *testing.T) {
	s, c, id, _, stored := restoreFixture(t)
	r, _, _ := c.GetQuarantine(id)
	foreign := filepath.Join(r.SourceCaptureDir, "unrelated")
	if err := os.WriteFile(foreign, []byte("benign"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Purge(id); err == nil {
		t.Fatal("unexpected capture contents were silently discarded")
	}
	requirePurgePayload(t, stored, "payload")
	requirePurgePayload(t, filepath.Join(r.SourceCaptureDir, "source"), "payload")
	requirePurgePayload(t, foreign, "benign")
	requirePurgeRecord(t, c, id, false)
}

func TestPurgeJournalRejectsChangedRecordAndMalformedData(t *testing.T) {
	for _, kind := range []string{"record", "nonce", "phase", "raw-identity", "capture-state", "missing-manifest", "oversized", "unknown-entry"} {
		t.Run(kind, func(t *testing.T) {
			s, c, id, _, stored := restoreFixture(t)
			before, _, _ := c.GetQuarantine(id)
			s.responseFault = func(phase string) error {
				if phase == "purge-intent" {
					return errors.New("safe interruption")
				}
				return nil
			}
			if err := s.Purge(id); err == nil {
				t.Fatal("missing interruption")
			}
			s.responseFault = nil
			root, _ := s.openRoot()
			journals, op, err := openPurgeOperation(root, id)
			if err != nil {
				t.Fatal(err)
			}
			m, err := readPurgeManifest(op, id)
			if err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "record":
				m.Record.Detection += ".changed"
			case "nonce":
				m.OperationID = "../foreign"
			case "phase":
				m.Phase = "unknown"
			case "raw-identity":
				m.Raw.ID = "other-object"
			case "capture-state":
				m.CaptureState = "deleted"
			}
			if err = writePurgeManifest(op, m); err != nil {
				t.Fatal(err)
			}
			manifest := filepath.Join(op.path, "manifest.json")
			switch kind {
			case "missing-manifest":
				if err = os.Rename(manifest, manifest+".retained"); err != nil {
					t.Fatal(err)
				}
			case "oversized":
				err = os.WriteFile(manifest, []byte(strings.Repeat("x", maxAdmissionJournal+1)), 0600)
			case "unknown-entry":
				// Unknown journal content blocks cleanup before destructive work.
				err = os.WriteFile(filepath.Join(op.path, "foreign"), []byte("benign"), 0600)
			}
			if err != nil {
				t.Fatal(err)
			}
			op.Close()
			journals.Close()
			root.Close()
			results, err := s.Recover()
			if err == nil || len(results) != 1 || results[0].Status != "needs_attention" {
				t.Fatalf("invalid journal was acted on: %+v %v", results, err)
			}
			requirePurgePayload(t, stored, "payload")
			requirePurgePayload(t, filepath.Join(before.SourceCaptureDir, "source"), "payload")
			requirePurgeRecord(t, c, id, false)
		})
	}
}

func TestPurgePendingJournalPreservesNonUTF8Paths(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("requires Linux filename bytes; the host APFS and Windows filesystems reject these names")
	}
	dir := filepath.Join(t.TempDir(), string([]byte{'p', 0xff}))
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	s, err := New(filepath.Join(dir, "quarantine"), c)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, string([]byte{'f', 0xfe}))
	if err = os.WriteFile(path, []byte("safe"), 0600); err != nil {
		t.Fatal(err)
	}
	id, err := s.Quarantine(path, payloadHash("safe"), "Safe.NonUTF8")
	if err != nil {
		t.Fatal(err)
	}
	s.responseFault = func(phase string) error {
		if phase == "purge-intent" {
			return errors.New("safe interruption")
		}
		return nil
	}
	if err = s.Purge(id); err == nil {
		t.Fatal("missing interruption")
	}
	s.responseFault = nil
	if results, err := s.Recover(); err != nil || len(results) != 1 || results[0].Status != "purged" || results[0].OriginalPath != path {
		t.Fatalf("purge journal lost pathname bytes: %+v %v", results, err)
	}
	requirePurgeRecord(t, c, id, true)
}

func FuzzPurgeJournal(f *testing.F) {
	dir := f.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		f.Fatal(err)
	}
	dir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		f.Fatal(err)
	}
	op, err := openAdmissionRoot(dir, true)
	if err != nil {
		f.Fatal(err)
	}
	f.Cleanup(func() { op.Close() })
	rec, err := cache.EncodeQuarantinePurgeRecord(cache.QuarantineRecord{QuarantineID: "fixture", QuarantinedAt: time.Unix(100, 0).UTC()})
	if err != nil {
		f.Fatal(err)
	}
	seed, err := json.Marshal(purgeManifest{Version: 1, OperationID: "e2a673a6-e55c-4adc-b25e-cb601f504d15", Phase: "prepared", ApprovedAt: time.Unix(100, 0).UTC(), RecordData: rec, Stored: purgeArtifact{State: "absent"}, Raw: purgeArtifact{State: "absent"}, CaptureState: "absent"})
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Add([]byte(`{"Version":1,"RecordData":"/w=="}`))
	f.Fuzz(func(t *testing.T, body []byte) {
		if len(body) > 32768 {
			t.Skip()
		}
		if err := os.WriteFile(filepath.Join(dir, "manifest.json"), body, 0600); err != nil {
			t.Fatal(err)
		}
		m, err := readPurgeManifest(op, "fixture")
		if err == nil && (m.Record.QuarantineID != "fixture" || m.Record.Purged || m.OperationID == "" || m.Version != 1) {
			t.Fatalf("accepted invalid purge identity: %s", fmt.Sprint(m))
		}
	})
}
