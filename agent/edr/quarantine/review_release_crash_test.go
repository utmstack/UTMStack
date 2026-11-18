package quarantine

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

func TestReviewReleaseCrashWorker(t *testing.T) {
	if os.Getenv("OPENEDR_RELEASE_WORKER") != "1" {
		t.Skip("owned release process-crash worker")
	}
	var a RecoveryReleaseApproval
	if err := json.Unmarshal([]byte(os.Getenv("OPENEDR_RELEASE_APPROVAL")), &a); err != nil {
		t.Fatal(err)
	}
	_, err := releaseRecovery(os.Getenv("OPENEDR_RELEASE_ROOT"), os.Getenv("OPENEDR_RELEASE_ID"), a, "owned explicit recovery release", func(at string) error {
		if at == os.Getenv("OPENEDR_RELEASE_PHASE") {
			os.Exit(97)
		}
		return nil
	})
	t.Fatalf("release crash checkpoint not reached: %v", err)
}

func TestReviewReleaseActualProcessCrashes(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"release-directory-created", "release-receipt-written", "release-approved", "release-before-move", "release-published", "release-synced"} {
		t.Run(phase, func(t *testing.T) {
			s, id, original, stored, a := reviewReleaseFixture(t, "verified")
			before, err := InspectRecovery(s.dir, id)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.cache.Close(); err != nil {
				t.Fatal(err)
			}
			encoded, _ := json.Marshal(a)
			cmd := exec.Command(self, "-test.run=^TestReviewReleaseCrashWorker$", "-test.count=1")
			cmd.Env = append(os.Environ(), "OPENEDR_RELEASE_WORKER=1", "OPENEDR_RELEASE_ROOT="+s.dir, "OPENEDR_RELEASE_ID="+id, "OPENEDR_RELEASE_APPROVAL="+string(encoded), "OPENEDR_RELEASE_PHASE="+phase)
			output, err := cmd.CombinedOutput()
			status, ok := err.(*exec.ExitError)
			if !ok || status.ExitCode() != 97 {
				t.Fatalf("worker did not exit at checkpoint: %v %s", err, output)
			}
			after, err := InspectRecovery(s.dir, id)
			if err != nil || !reflect.DeepEqual(before.Journals[1], after.Journals[1]) {
				t.Fatal("release itself replayed or changed original restore", err)
			}
			moved := phase == "release-published" || phase == "release-synced"
			if (after.Journals[2].Status == "absent") != moved {
				t.Fatal("unexpected active fence boundary", after.Journals[2])
			}
			c, err := cache.Open(filepath.Join(filepath.Dir(s.dir), "edr.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			recovered, err := New(s.dir, c)
			if err != nil {
				t.Fatal(err)
			}
			if !moved {
				for _, result := range recovered.RecoveryStatus() {
					if result.Status != "operator_held" {
						t.Fatal("unreleased hold allowed recovery", result)
					}
				}
			} else if _, err := os.Lstat(filepath.Join(s.dir, ".restores", id)); !os.IsNotExist(err) {
				t.Fatal("explicitly released original recovery was still fenced", err)
			}
			decision, err := ReleaseRecovery(s.dir, id, a, "owned explicit recovery release")
			if phase == "release-directory-created" {
				if err == nil {
					t.Fatal("bare history directory adopted as release approval")
				}
			} else {
				if err != nil {
					t.Fatal("identity-bound release could not resume", err)
				}
				again, err := ReleaseRecovery(s.dir, id, a, decision.Reason)
				if err != nil || !reflect.DeepEqual(decision, again) {
					t.Fatal("resumed release did not stabilize", err)
				}
			}
			rec, found, err := c.GetQuarantine(id)
			if err != nil || !found || rec.Purged || rec.Restored {
				t.Fatal("release claimed original completion", rec, err)
			}
			requirePurgePayload(t, stored, "payload")
			requirePurgePayload(t, filepath.Join(rec.SourceCaptureDir, "source"), "payload")
			if _, err := os.Lstat(original); !os.IsNotExist(err) {
				t.Fatal("release published original destination", err)
			}
			t.Logf("actual exit 97 at %s; moved=%t; no false original completion", phase, moved)
		})
	}
}
