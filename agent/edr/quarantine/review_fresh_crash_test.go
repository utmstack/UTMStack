package quarantine

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestFreshApprovalCrashWorker(t *testing.T) {
	if os.Getenv("OPENEDR_FRESH_APPROVAL_CRASH_WORKER") != "1" {
		t.Skip("owned fresh approval crash worker")
	}
	body, err := os.ReadFile(os.Getenv("OPENEDR_FRESH_APPROVAL_INSPECTION"))
	if err != nil {
		t.Fatal(err)
	}
	var r FreshRestoreInspection
	if err = json.Unmarshal(body, &r); err != nil {
		t.Fatal(err)
	}
	_, err = approveFreshRestore(os.Getenv("OPENEDR_FRESH_APPROVAL_ROOT"), r.QuarantineID, os.Getenv("OPENEDR_FRESH_APPROVAL_OPERATION"), r, "approve exact new restore after independent review", func(at string) error {
		if at == os.Getenv("OPENEDR_FRESH_APPROVAL_PHASE") {
			os.Exit(97)
		}
		return nil
	})
	t.Fatal("fresh approval crash checkpoint did not exit", err)
}

func TestFreshApprovalActualProcessCrashes(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"fresh-approval-directory-created", "fresh-approval-written", "fresh-approval-synced"} {
		t.Run(phase, func(t *testing.T) {
			s, id, destination, stored, a := freshInspectionFixture(t)
			r, err := InspectFreshRestore(s.dir, id, a.HoldDecisionID, destination)
			if err != nil {
				t.Fatal(err)
			}
			body, err := json.Marshal(r)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(filepath.Dir(s.dir), "owned-fresh-inspection.json")
			if err = os.WriteFile(path, body, 0600); err != nil {
				t.Fatal(err)
			}
			operationID := uuid.NewString()
			cmd := exec.Command(self, "-test.run=^TestFreshApprovalCrashWorker$", "-test.count=1")
			cmd.Env = append(os.Environ(), "OPENEDR_FRESH_APPROVAL_CRASH_WORKER=1", "OPENEDR_FRESH_APPROVAL_ROOT="+s.dir, "OPENEDR_FRESH_APPROVAL_INSPECTION="+path, "OPENEDR_FRESH_APPROVAL_OPERATION="+operationID, "OPENEDR_FRESH_APPROVAL_PHASE="+phase)
			output, err := cmd.CombinedOutput()
			if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 97 {
				t.Fatalf("fresh approval checkpoint did not exit97: %v %s", err, output)
			}
			_, err = ApproveFreshRestore(s.dir, id, operationID, r, "approve exact new restore after independent review")
			if (err != nil) != (phase == "fresh-approval-directory-created") {
				t.Fatal("approval process exit adopted bare history or refused complete receipt", err)
			}
			if err = s.Restore(id); !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("process exit released normal recovery", err)
			}
			if err = s.Purge(id); !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("process exit released purge", err)
			}
			requirePurgePayload(t, stored, "payload")
			if _, err = os.Lstat(destination); !os.IsNotExist(err) {
				t.Fatal("process exit published the new destination", err)
			}
		})
	}
}
