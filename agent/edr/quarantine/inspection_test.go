package quarantine

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// Capture names, bytes and permissions, so inspection cannot quietly replay a
// recoverable transaction or create locks/security metadata during review.
func inspectionTree(t *testing.T, root string) map[string]string {
	t.Helper()
	result := make(map[string]string)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		value := info.Mode().String()
		if info.Mode().IsRegular() {
			body, e := os.ReadFile(path)
			if e != nil {
				return e
			}
			value += fmt.Sprintf(":%x", sha256.Sum256(body))
		}
		result[path] = value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestInspectRecoveryDoesNotReplayInterruptedOperations(t *testing.T) {
	for _, operation := range []string{"admission", "restore", "purge"} {
		t.Run(operation, func(t *testing.T) {
			var s *Store
			var id, original, stored string
			if operation == "admission" {
				fixture, c, path, _ := admissionFixture(t)
				s, original = fixture, path
				s.admissionFault = func(phase string) error {
					if phase == "claimed" {
						return errors.New("interrupted admission")
					}
					return nil
				}
				var err error
				id, err = s.Quarantine(path, payloadHash("malicious marker"), "Safe.Inspection.Control")
				if err == nil || id == "" {
					t.Fatalf("expected pending admission: %s %v", id, err)
				}
				if err := c.Close(); err != nil {
					t.Fatal(err)
				}
			} else {
				fixture, c, qid, path, backup := restoreFixture(t)
				s, id, original, stored = fixture, qid, path, backup
				checkpoint := "restore-published-fs"
				if operation == "purge" {
					checkpoint = "purge-intent"
				}
				s.responseFault = func(phase string) error {
					if phase == checkpoint {
						return errors.New("interrupted " + operation)
					}
					return nil
				}
				var err error
				if operation == "restore" {
					err = s.Restore(id)
				} else {
					err = s.Purge(id)
				}
				if err == nil {
					t.Fatal("fault checkpoint was not reached")
				}
				if err := c.Close(); err != nil {
					t.Fatal(err)
				}
			}
			before := inspectionTree(t, filepath.Dir(s.dir))
			for attempt := 0; attempt < 2; attempt++ {
				report, err := InspectRecovery(s.dir, id)
				if err != nil || report.RootID != s.rootID || len(report.Journals) != 3 {
					t.Fatalf("inspection: %+v %v", report, err)
				}
				for _, journal := range report.Journals {
					if journal.Operation != operation {
						if journal.Status != "absent" {
							t.Fatalf("unexpected journal: %+v", journal)
						}
						continue
					}
					if journal.Status != "captured" || journal.DirectoryID == "" || journal.ManifestID == "" || journal.SHA256 != payloadHash(string(journal.Manifest)) {
						t.Fatalf("missing exact evidence: %+v", journal)
					}
					if !json.Valid(journal.Manifest) {
						t.Fatal("fixture journal is not JSON")
					}
				}
			}
			if after := inspectionTree(t, filepath.Dir(s.dir)); !reflect.DeepEqual(before, after) {
				t.Fatalf("inspection changed private state: before=%v after=%v", before, after)
			}
			if operation == "restore" {
				requirePurgePayload(t, original, "payload")
				requirePurgePayload(t, stored, "payload")
			}
		})
	}
}

func TestInspectRecoveryRetainsMalformedAndIncompleteJournals(t *testing.T) {
	for _, kind := range []string{"missing", "malformed", "oversized", "directory", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			s, _, id, _, _ := restoreFixture(t)
			op := filepath.Join(s.dir, ".restores", id)
			root, err := s.openRoot()
			if err != nil {
				t.Fatal(err)
			}
			journals, err := admissionChild(root, ".restores")
			root.Close()
			if err != nil {
				t.Fatal(err)
			}
			operation, err := admissionChild(journals, id)
			journals.Close()
			if err != nil {
				t.Fatal(err)
			}
			operation.Close()
			manifest := filepath.Join(op, "manifest.json")
			body := []byte{'{', 0xff, 0, 'x'}
			switch kind {
			case "malformed":
				if err := os.WriteFile(manifest, body, 0600); err != nil {
					t.Fatal(err)
				}
			case "oversized":
				if err := os.WriteFile(manifest, make([]byte, maxAdmissionJournal+1), 0600); err != nil {
					t.Fatal(err)
				}
			case "directory":
				if err := os.Mkdir(manifest, 0700); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				target := filepath.Join(filepath.Dir(s.dir), "foreign-journal")
				if err := os.WriteFile(target, body, 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, manifest); err != nil {
					t.Skipf("symbolic links unavailable: %v", err)
				}
			}
			before := inspectionTree(t, s.dir)
			report, err := InspectRecovery(s.dir, id)
			entry := report.Journals[1]
			if kind == "malformed" {
				if err != nil || entry.Status != "captured" || !bytes.Equal(entry.Manifest, body) {
					t.Fatalf("malformed bytes not preserved: %+v %v", entry, err)
				}
				encoded, err := json.Marshal(report)
				if err != nil {
					t.Fatal(err)
				}
				var decoded RecoveryInspection
				if err := json.Unmarshal(encoded, &decoded); err != nil || !bytes.Equal(decoded.Journals[1].Manifest, body) {
					t.Fatal("JSON inspection lost original bytes", err)
				}
			} else if err == nil || entry.Status != "unreadable" || entry.Manifest != nil || entry.SHA256 != "" {
				t.Fatalf("incomplete evidence reported as captured: %+v %v", entry, err)
			}
			if after := inspectionTree(t, s.dir); !reflect.DeepEqual(before, after) {
				t.Fatal("inspection modified unresolved evidence")
			}
		})
	}
}

func TestInspectRecoveryDoesNotCreateStoreOrAcceptTraversal(t *testing.T) {
	root := t.TempDir()
	missing := filepath.Join(root, "missing-store")
	if _, err := InspectRecovery(missing, uuid.NewString()); !os.IsNotExist(err) {
		t.Fatalf("missing root: %v", err)
	}
	for _, id := range []string{"", "../foreign", uuid.NewString() + "/child"} {
		if _, err := InspectRecovery(missing, id); err == nil {
			t.Fatalf("accepted invalid id: %q", id)
		}
	}
	if entries, err := os.ReadDir(root); err != nil || len(entries) != 0 {
		t.Fatalf("inspection created storage: %v %v", entries, err)
	}
}

func TestInspectRecoveryEmptyStoreDoesNotCreateLocks(t *testing.T) {
	s, _, _, _ := admissionFixture(t)
	before := inspectionTree(t, s.dir)
	report, err := InspectRecovery(s.dir, uuid.NewString())
	if err != nil || len(report.Journals) != 3 {
		t.Fatalf("empty store: %+v %v", report, err)
	}
	for _, entry := range report.Journals {
		if entry.Status != "absent" {
			t.Fatalf("unexpected recovery entry: %+v", entry)
		}
	}
	if after := inspectionTree(t, s.dir); !reflect.DeepEqual(before, after) {
		t.Fatal("inspection created a lock or journal in an empty store")
	}
}
