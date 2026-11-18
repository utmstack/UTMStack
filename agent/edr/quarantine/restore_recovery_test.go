package quarantine

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

var restoreCheckpoints = []string{
	"directory-created", "intent", "stage-created", "output-created", "verified",
	"published-fs", "published", "before-database-commit", "database-committed", "committed",
	"stage-delete-intent", "stage-removed-fs", "stage-deleted",
	"source-claim-intent", "source-claimed-fs", "source-claimed", "source-delete-intent", "source-unlinked", "source-deleted", "journal-cleanup",
}

func restoreFixtureManifest(t *testing.T, s *Store, id string) restoreManifest {
	t.Helper()
	root, err := s.openRoot()
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	journals, op, err := openRestoreOperation(root, id)
	if err != nil {
		t.Fatal(err)
	}
	defer journals.Close()
	defer op.Close()
	m, err := readRestoreManifest(op, id)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestRestoreCrashWorker(t *testing.T) {
	if os.Getenv("OPENEDR_RESTORE_CRASH_WORKER") != "1" {
		t.Skip("owned process-crash worker; no crash seam exists in shipped executable")
	}
	c, err := cache.Open(os.Getenv("OPENEDR_RESTORE_DB"))
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(os.Getenv("OPENEDR_RESTORE_ROOT"), c)
	if err != nil {
		t.Fatal(err)
	}
	s.responseFault = func(phase string) error {
		if phase == "restore-"+os.Getenv("OPENEDR_RESTORE_PHASE") {
			os.Exit(97)
		}
		return nil
	}
	t.Fatalf("checkpoint not reached: %v", s.Restore(os.Getenv("OPENEDR_RESTORE_ID")))
}

func TestRestoreProcessCrashRecovery(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, checkpoint := range restoreCheckpoints {
		t.Run(checkpoint, func(t *testing.T) {
			s, c, id, original, stored := restoreFixture(t)
			rec, _, err := c.GetQuarantine(id)
			if err != nil {
				t.Fatal(err)
			}
			if err = c.Close(); err != nil {
				t.Fatal(err)
			}
			db := filepath.Join(filepath.Dir(s.dir), "edr.db")
			cmd := exec.Command(self, "-test.run=^TestRestoreCrashWorker$", "-test.count=1")
			cmd.Env = append(os.Environ(), "OPENEDR_RESTORE_CRASH_WORKER=1", "OPENEDR_RESTORE_DB="+db, "OPENEDR_RESTORE_ROOT="+s.dir, "OPENEDR_RESTORE_ID="+id, "OPENEDR_RESTORE_PHASE="+checkpoint)
			output, err := cmd.CombinedOutput()
			status, ok := err.(*exec.ExitError)
			if !ok || status.ExitCode() != 97 {
				t.Fatalf("worker did not exit abruptly: %v %s", err, output)
			}
			reopened, err := cache.Open(db)
			if err != nil {
				t.Fatal(err)
			}
			defer reopened.Close()
			recovered, err := New(s.dir, reopened)
			if err != nil {
				t.Fatal(err)
			}
			results := recovered.RecoveryStatus()
			if len(results) != 1 || results[0].Operation != "restore" || results[0].Status == "needs_attention" {
				t.Fatalf("actual crash not reconciled: %+v", results)
			}
			r, found, err := reopened.GetQuarantine(id)
			published := checkpoint != "directory-created" && checkpoint != "intent" && checkpoint != "stage-created" && checkpoint != "output-created" && checkpoint != "verified"
			if err != nil || !found || r.Restored != published || r.Purged {
				t.Fatalf("incorrect recovered record: %+v %v", r, err)
			}
			if published {
				requirePurgePayload(t, original, "payload")
				if _, err = os.Lstat(stored); !os.IsNotExist(err) {
					t.Fatal("backup not retired after receipt", err)
				}
			} else {
				requirePurgePayload(t, stored, "payload")
				if _, err = os.Lstat(original); !os.IsNotExist(err) {
					t.Fatal("unpublished crash restored original", err)
				}
			}
			requirePurgePayload(t, filepath.Join(rec.SourceCaptureDir, "source"), "payload")
			if again, err := recovered.Recover(); err != nil || len(again) != 0 {
				t.Fatalf("replay not empty: %+v %v", again, err)
			}
			t.Logf("actual process exit=97; restored=%t; status=%s", r.Restored, results[0].Status)
		})
	}
}

func TestRestoreRecoveryPreservesLaterDestination(t *testing.T) {
	for _, checkpoint := range []string{"published-fs", "published", "database-committed", "source-claimed"} {
		for _, replacement := range []string{"edit", "replace", "remove"} {
			t.Run(checkpoint+"/"+replacement, func(t *testing.T) {
				s, c, id, original, stored := restoreFixture(t)
				s.responseFault = func(phase string) error {
					if phase == "restore-"+checkpoint {
						return errors.New("safe interruption")
					}
					return nil
				}
				if err := s.Restore(id); err == nil {
					t.Fatal("checkpoint not reached")
				}
				s.responseFault = nil
				if replacement != "edit" {
					if err := os.Rename(original, original+"-moved"); err != nil {
						t.Fatal(err)
					}
				}
				if replacement != "remove" {
					if err := os.WriteFile(original, []byte("later benign content"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				results, err := s.Recover()
				// Before the durable publication entry, only the original output inode
				// proves publication. Equal bytes or a replacement pathname do not.
				ambiguous := checkpoint == "published-fs" && replacement != "edit"
				if ambiguous {
					if err == nil || len(results) != 1 || results[0].Status != "needs_attention" {
						t.Fatalf("ambiguous publication adopted: %+v %v", results, err)
					}
					requirePurgePayload(t, stored, "payload")
				} else {
					if err != nil || len(results) != 1 || results[0].Status != "restored" {
						t.Fatalf("known publication not reconciled: %+v %v", results, err)
					}
				}
				if replacement != "remove" {
					requirePurgePayload(t, original, "later benign content")
				} else if _, err = os.Lstat(original); !os.IsNotExist(err) {
					t.Fatal("removed destination recreated", err)
				}
				r, _, err := c.GetQuarantine(id)
				if err != nil || r.Restored == ambiguous {
					t.Fatal("incorrect historical restore state", err)
				}
			})
		}
	}
}

func TestRestoreUnidentifiedCreationWindowsRetainForAttention(t *testing.T) {
	for _, checkpoint := range []string{"stage-created-fs", "output-created-fs", "publication-intent"} {
		t.Run(checkpoint, func(t *testing.T) {
			s, c, id, original, stored := restoreFixture(t)
			s.responseFault = func(phase string) error {
				if phase == "restore-"+checkpoint {
					return errors.New("safe interruption")
				}
				return nil
			}
			if err := s.Restore(id); err == nil {
				t.Fatal("checkpoint not reached")
			}
			s.responseFault = nil
			results, err := s.Recover()
			if err == nil || len(results) != 1 || results[0].Status != "needs_attention" {
				t.Fatalf("missing creation/publication identity was guessed: %+v %v", results, err)
			}
			requirePurgePayload(t, stored, "payload")
			if _, err = os.Lstat(original); !os.IsNotExist(err) {
				t.Fatal("uncertain output was published", err)
			}
			if r, _, err := c.GetQuarantine(id); err != nil || r.Restored {
				t.Fatal("unproven restore marked successful", err)
			}
			if err := s.Purge(id); err == nil {
				t.Fatal("purge bypassed pending restore")
			}
		})
	}
}

func TestRestoreRecoveryPreservesChangedIdentitiesAndJournal(t *testing.T) {
	for _, kind := range []string{"parent", "stage", "output", "source", "record", "unknown-journal", "malformed-journal"} {
		t.Run(kind, func(t *testing.T) {
			s, c, id, original, stored := restoreFixture(t)
			phase := "verified"
			if kind == "source" {
				phase = "published"
			}
			s.responseFault = func(p string) error {
				if p == "restore-"+phase {
					return errors.New("safe interruption")
				}
				return nil
			}
			if err := s.Restore(id); err == nil {
				t.Fatal("checkpoint not reached")
			}
			s.responseFault = nil
			m := restoreFixtureManifest(t, s, id)
			stage := filepath.Join(string(m.ParentPath), m.stageName())
			op := filepath.Join(s.dir, ".restores", id)
			preserved := ""
			switch kind {
			case "parent":
				// The fixture parent also contains the database; move only a new
				// configured destination in the separate parent replacement test.
				m.ParentID += "-different"
			case "stage":
				if err := os.Rename(stage, stage+"-moved"); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(stage, 0700); err != nil {
					t.Fatal(err)
				}
				preserved = filepath.Join(stage, "sentinel")
			case "output":
				output := filepath.Join(stage, "snapshot")
				if err := os.Rename(output, output+"-moved"); err != nil {
					t.Fatal(err)
				}
				preserved = output
			case "source":
				if err := os.Rename(stored, stored+"-moved"); err != nil {
					t.Fatal(err)
				}
				preserved = stored
			case "record":
				r, _, err := c.GetQuarantine(id)
				if err != nil {
					t.Fatal(err)
				}
				r.Detection += "-changed"
				if err = c.StoreQuarantine(r); err != nil {
					t.Fatal(err)
				}
				// A published operation cannot commit against changed approval.
				m.Phase = "published"
			case "unknown-journal":
				preserved = filepath.Join(op, "foreign")
			case "malformed-journal":
				if err := os.WriteFile(filepath.Join(op, "manifest.json"), []byte("{"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if preserved != "" {
				if err := os.WriteFile(preserved, []byte("foreign preserved content"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if kind == "parent" || kind == "record" {
				root, err := s.openRoot()
				if err != nil {
					t.Fatal(err)
				}
				journals, handle, err := openRestoreOperation(root, id)
				if err != nil {
					t.Fatal(err)
				}
				err = writeRestoreManifest(handle, m)
				handle.Close()
				journals.Close()
				root.Close()
				if err != nil {
					t.Fatal(err)
				}
			}
			results, err := s.Recover()
			if err == nil || len(results) != 1 || results[0].Status != "needs_attention" {
				t.Fatalf("conflicting evidence accepted: %+v %v", results, err)
			}
			if preserved != "" {
				requirePurgePayload(t, preserved, "foreign preserved content")
			}
			if kind != "source" {
				requirePurgePayload(t, stored, "payload")
			}
			if kind == "source" {
				requirePurgePayload(t, original, "payload")
			}
			if _, err = os.Stat(op); err != nil {
				t.Fatal("recovery evidence lost", err)
			}
		})
	}
}

func TestRestoreExplicitRetryAfterUnpublishedInterruption(t *testing.T) {
	s, c, id, original, _ := restoreFixture(t)
	s.responseFault = func(p string) error {
		if p == "restore-verified" {
			return errors.New("safe interruption")
		}
		return nil
	}
	if err := s.Restore(id); err == nil {
		t.Fatal("checkpoint not reached")
	}
	s.responseFault = nil
	if err := s.Restore(id); err != nil {
		t.Fatal("explicit retry failed", err)
	}
	requirePurgePayload(t, original, "payload")
	if r, _, err := c.GetQuarantine(id); err != nil || !r.Restored {
		t.Fatal("retry returned success without restore", err)
	}
}

func FuzzRestoreJournal(f *testing.F) {
	f.Add([]byte{0xff, 0x80, 'a'}, []byte("{}"))
	f.Fuzz(func(t *testing.T, pathBytes, raw []byte) {
		if len(pathBytes) > 32768 || len(raw) > maxAdmissionJournal+1 {
			t.Skip()
		}
		// Exercise lossless metadata independently of whether a particular OS
		// accepts those bytes in a filename. Malformed manifests must not panic.
		r := cache.QuarantineRecord{QuarantineID: "fixture", OriginalPath: string(pathBytes), Detection: string(pathBytes), OriginalSecurity: string(pathBytes)}
		encoded, err := cache.EncodeQuarantinePurgeRecord(r)
		if err != nil {
			t.Fatal(err)
		}
		m := restoreManifest{RecordData: encoded, ParentPath: pathBytes, Leaf: pathBytes}
		body, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		var decoded restoreManifest
		if err = json.Unmarshal(body, &decoded); err != nil || !reflect.DeepEqual(m.ParentPath, decoded.ParentPath) || !reflect.DeepEqual(m.Leaf, decoded.Leaf) {
			t.Fatal("restore paths lost bytes", err)
		}
		roundtrip, err := cache.DecodeQuarantinePurgeRecord(decoded.RecordData)
		if err != nil || !reflect.DeepEqual(roundtrip.OriginalPath, r.OriginalPath) || roundtrip.OriginalSecurity != r.OriginalSecurity {
			t.Fatal("restore record lost bytes", err)
		}
		_, _ = decodeRestoreManifest(raw, "fixture")
	})
}
