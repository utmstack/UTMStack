package quarantine

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

// No environment-variable crash seam exists in the shipped executable. Only
// this test worker installs the per-store hook, then exits without defers.
func TestPurgeCrashWorker(t *testing.T) {
	if os.Getenv("OPENEDR_PURGE_CRASH_WORKER") != "1" {
		t.Skip("owned abrupt-exit worker, invoked by the process-crash matrix")
	}
	c, err := cache.Open(os.Getenv("OPENEDR_PURGE_CRASH_DB"))
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(os.Getenv("OPENEDR_PURGE_CRASH_ROOT"), c)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := os.Getenv("OPENEDR_PURGE_CRASH_PHASE")
	s.responseFault = func(phase string) error {
		if phase == checkpoint {
			os.Exit(97)
		}
		return nil
	}
	err = s.Purge(os.Getenv("OPENEDR_PURGE_CRASH_ID"))
	t.Fatalf("worker did not reach abrupt-exit checkpoint %q: %v", checkpoint, err)
}

func TestPurgeProcessCrashRecovery(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	observations := 0
	for _, restored := range []bool{false, true} {
		for _, checkpoint := range purgeCheckpoints {
			if restored && strings.HasPrefix(checkpoint, "stored-") {
				continue
			}
			t.Run(fmt.Sprintf("restored-%t/%s", restored, checkpoint), func(t *testing.T) {
				s, c, id, original, stored := restoreFixture(t)
				if restored {
					if err := s.Restore(id); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(original, []byte("benign replacement after restore"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				before, _, err := c.GetQuarantine(id)
				if err != nil {
					t.Fatal(err)
				}
				if err = c.Close(); err != nil {
					t.Fatal(err)
				}
				dbPath := filepath.Join(filepath.Dir(s.dir), "edr.db")
				cmd := exec.Command(self, "-test.run=^TestPurgeCrashWorker$", "-test.count=1")
				cmd.Env = append(os.Environ(), "OPENEDR_PURGE_CRASH_WORKER=1", "OPENEDR_PURGE_CRASH_DB="+dbPath, "OPENEDR_PURGE_CRASH_ROOT="+s.dir, "OPENEDR_PURGE_CRASH_ID="+id, "OPENEDR_PURGE_CRASH_PHASE="+checkpoint)
				output, err := cmd.CombinedOutput()
				status, isExit := err.(*exec.ExitError)
				if !isExit || status.ExitCode() != 97 {
					t.Fatalf("worker did not stop abruptly: %v %s", err, output)
				}
				reopened, err := cache.Open(dbPath)
				if err != nil {
					t.Fatal(err)
				}
				defer reopened.Close()
				recovered, err := New(s.dir, reopened)
				if err != nil {
					t.Fatal(err)
				}
				results := recovered.RecoveryStatus()
				if len(results) != 1 || results[0].Status == "needs_attention" || results[0].Operation != "purge" {
					t.Fatalf("actual crash was not reconciled: %+v", results)
				}
				purged := checkpoint != "purge-directory-created"
				r := requirePurgeRecord(t, reopened, id, purged)
				if r.Restored != restored {
					t.Fatal("purge changed historical restore state")
				}
				if purged {
					for _, path := range []string{stored, before.SourceCaptureDir} {
						if _, err = os.Lstat(path); !os.IsNotExist(err) {
							t.Fatalf("purged artifact remains after recovery: %s %v", path, err)
						}
					}
				} else {
					if !restored {
						requirePurgePayload(t, stored, "payload")
					}
					requirePurgePayload(t, filepath.Join(before.SourceCaptureDir, "source"), "payload")
				}
				if restored {
					requirePurgePayload(t, original, "benign replacement after restore")
				} else if _, err = os.Lstat(original); !os.IsNotExist(err) {
					t.Fatal("purge recovery published an original path", err)
				}
				if again, err := recovered.Recover(); err != nil || len(again) != 0 {
					t.Fatalf("second startup replay was not empty: %+v %v", again, err)
				}
				observations++
				t.Logf("actual process exit=97; recovered=%s; purged=%t; restored=%t", results[0].Status, r.Purged, r.Restored)
			})
		}
	}
	if observations != 36 {
		t.Fatalf("incomplete actual process-crash matrix: %d", observations)
	}
}

func TestPurgeUnreadableArtifacts(t *testing.T) {
	s, c, id, _, stored := restoreFixture(t)
	r, _, _ := c.GetQuarantine(id)
	for _, path := range []string{stored, filepath.Join(r.SourceCaptureDir, "source")} {
		if err := os.Chmod(path, 0); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Purge(id); err != nil {
		t.Fatalf("purge unnecessarily requires payload read permission: %v", err)
	}
	requirePurgeRecord(t, c, id, true)
}

// The native acceptance harness invokes this worker in an exclusively created
// disposable app directory. It does not install a seam in the shipped CLI.
func TestPurgeCLILabWorker(t *testing.T) {
	app := os.Getenv("OPENEDR_PURGE_CLI_APP")
	if app == "" {
		t.Skip("owned CLI acceptance fixture worker")
	}
	if !filepath.IsAbs(app) {
		t.Fatal("CLI fixture requires an absolute disposable directory")
	}
	if err := os.Mkdir(app, 0700); err != nil {
		t.Fatalf("refuse to take over an existing fixture directory: %v", err)
	}
	c, err := cache.Open(filepath.Join(app, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	s, err := New(filepath.Join(app, "quarantine"), c)
	if err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(app, "original.txt")
	if err := os.WriteFile(original, []byte("safe CLI purge fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	id, err := s.Quarantine(original, payloadHash("safe CLI purge fixture"), "Safe.Purge.CLILab")
	if err != nil {
		t.Fatal(err)
	}
	restored := os.Getenv("OPENEDR_PURGE_CLI_RESTORED") == "1"
	if restored {
		if err := s.Restore(id); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(original, []byte("benign restored replacement"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	r, _, err := c.GetQuarantine(id)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := cache.EncodeQuarantinePurgeRecord(r)
	if err != nil {
		t.Fatal(err)
	}
	config, _ := json.Marshal(map[string]any{"quarantine_dir": s.dir})
	for name, body := range map[string][]byte{"fixture.id": []byte(id), "fixture.record": encoded, "edr.json": config} {
		if err := os.WriteFile(filepath.Join(app, name), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
