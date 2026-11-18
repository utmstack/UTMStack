package quarantine

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFreshCancellationCrashWorker(t *testing.T) {
	if os.Getenv("OPENEDR_FRESH_CANCEL_CRASH_WORKER") != "1" {
		t.Skip("owned fresh cancellation crash worker")
	}
	body, err := os.ReadFile(os.Getenv("OPENEDR_FRESH_CANCEL_INSPECTION"))
	if err != nil {
		t.Fatal(err)
	}
	var r FreshCancellationInspection
	if err = json.Unmarshal(body, &r); err != nil {
		t.Fatal(err)
	}
	_, err = cancelFreshRestore(os.Getenv("OPENEDR_FRESH_CANCEL_ROOT"), r.QuarantineID, r.OperationID, r, "retain exact incomplete fresh operation", func(at string) error {
		if at == os.Getenv("OPENEDR_FRESH_CANCEL_PHASE") {
			os.Exit(97)
		}
		return nil
	})
	t.Fatal("fresh cancellation crash checkpoint did not exit", err)
}

func TestFreshCancellationActualProcessCrashes(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"receipt-created", "receipt-written", "approved", "before-move", "archived", "synced"} {
		t.Run(phase, func(t *testing.T) {
			s, id, destination, stored, a, body := freshExecutionFixture(t)
			_, err := executeFreshFixture(s, body, func(at string) error {
				if at == "fresh-execution-publication-intent" {
					return errors.New("owned publication intent")
				}
				return nil
			})
			if err == nil {
				t.Fatal("publication checkpoint not reached")
			}
			r, err := InspectFreshCancellation(s.dir, id, a.OperationID)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(filepath.Dir(s.dir), "owned-fresh-cancel-inspection.json")
			saved, err := json.Marshal(r)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(path, saved, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(self, "-test.run=^TestFreshCancellationCrashWorker$", "-test.count=1")
			cmd.Env = append(os.Environ(), "OPENEDR_FRESH_CANCEL_CRASH_WORKER=1", "OPENEDR_FRESH_CANCEL_ROOT="+s.dir, "OPENEDR_FRESH_CANCEL_INSPECTION="+path, "OPENEDR_FRESH_CANCEL_PHASE=fresh-cancellation-"+phase)
			output, err := cmd.CombinedOutput()
			if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 97 {
				t.Fatalf("fresh cancellation did not exit97: %v %s", err, output)
			}
			_, err = CancelFreshRestore(s.dir, id, a.OperationID, r, "retain exact incomplete fresh operation")
			if (err != nil) != (phase == "receipt-created") {
				t.Fatal("fresh cancellation crash adopted incomplete history or refused exact complete receipt", err)
			}
			requirePurgePayload(t, stored, "payload")
			requirePurgePayload(t, filepath.Join(filepath.Dir(destination), ".edr-restore-"+a.OperationID, "snapshot"), "payload")
			if _, err = os.Lstat(destination); !os.IsNotExist(err) {
				t.Fatal("fresh cancellation crash published a destination", err)
			}
			if err = s.Restore(id); !errors.Is(err, ErrRecoveryHeld) {
				t.Fatal("fresh cancellation crash released original fence", err)
			}
		})
	}
}
