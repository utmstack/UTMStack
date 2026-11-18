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

func TestReviewCancelCrashWorker(t *testing.T) {
	if os.Getenv("OPENEDR_CANCEL_WORKER") != "1" {
		t.Skip("owned cancellation process-crash worker")
	}
	var a RecoveryReleaseApproval
	if err := json.Unmarshal([]byte(os.Getenv("OPENEDR_CANCEL_APPROVAL")), &a); err != nil {
		t.Fatal(err)
	}
	_, err := cancelRecovery(os.Getenv("OPENEDR_CANCEL_ROOT"), os.Getenv("OPENEDR_CANCEL_ID"), a, "owned explicit cancellation", func(at string) error {
		if at == os.Getenv("OPENEDR_CANCEL_PHASE") {
			os.Exit(97)
		}
		return nil
	})
	t.Fatalf("cancellation crash checkpoint not reached: %v", err)
}

func TestReviewCancelActualProcessCrashes(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"cancel-directory-created", "cancel-receipt-written", "cancel-approved", "cancel-before-move", "cancel-archived", "cancel-synced"} {
		t.Run(phase, func(t *testing.T) {
			s, id, original, stored, a := reviewReleaseFixture(t, "publication-intent")
			before, err := InspectRecovery(s.dir, id)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.cache.Close(); err != nil {
				t.Fatal(err)
			}
			encoded, _ := json.Marshal(a)
			cmd := exec.Command(self, "-test.run=^TestReviewCancelCrashWorker$", "-test.count=1")
			cmd.Env = append(os.Environ(), "OPENEDR_CANCEL_WORKER=1", "OPENEDR_CANCEL_ROOT="+s.dir, "OPENEDR_CANCEL_ID="+id, "OPENEDR_CANCEL_APPROVAL="+string(encoded), "OPENEDR_CANCEL_PHASE="+phase)
			output, err := cmd.CombinedOutput()
			status, ok := err.(*exec.ExitError)
			if !ok || status.ExitCode() != 97 {
				t.Fatalf("cancellation worker did not exit at checkpoint: %v %s", err, output)
			}
			after, err := InspectRecovery(s.dir, id)
			moved := phase == "cancel-archived" || phase == "cancel-synced"
			if err != nil || !reflect.DeepEqual(before.Journals[2], after.Journals[2]) || (after.Journals[1].Status == "absent") != moved {
				t.Fatal("cancellation lost fence or changed the wrong boundary", after, err)
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
			if err := recovered.Restore(id); !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("actual process exit permitted restore", err)
			}
			m, err := CancelRecovery(s.dir, id, a, "owned explicit cancellation")
			if phase == "cancel-directory-created" {
				if err == nil {
					t.Fatal("bare directory adopted as cancellation approval")
				}
			} else {
				if err != nil {
					t.Fatal("exact cancellation could not resume", err)
				}
				if again, err := CancelRecovery(s.dir, id, a, m.Reason); err != nil || !reflect.DeepEqual(m, again) {
					t.Fatal("resumed cancellation did not stabilize", err)
				}
			}
			requirePurgePayload(t, stored, "payload")
			if _, err := os.Lstat(original); !os.IsNotExist(err) {
				t.Fatal("cancellation exit published the destination", err)
			}
			t.Logf("actual exit 97 at %s; archived=%t; fence unchanged", phase, moved)
		})
	}
}
