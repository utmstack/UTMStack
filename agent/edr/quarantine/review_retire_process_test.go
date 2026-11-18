package quarantine

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRetirementCrashWorker(t *testing.T) {
	if os.Getenv("OPENEDR_RETIRE_CRASH_WORKER") != "1" {
		t.Skip("owned retirement process-exit worker")
	}
	body, err := os.ReadFile(os.Getenv("OPENEDR_RETIRE_INSPECTION"))
	if err != nil {
		t.Fatal(err)
	}
	var r CancelledRetirementInspection
	if err = json.Unmarshal(body, &r); err != nil {
		t.Fatal(err)
	}
	_, err = retireCancelledRecovery(os.Getenv("OPENEDR_RETIRE_ROOT"), r, "retain every object after process exit", "", func(at string) error {
		if at == os.Getenv("OPENEDR_RETIRE_PHASE") {
			os.Exit(97)
		}
		return nil
	})
	t.Fatal("retirement process-exit checkpoint not reached", err)
}

func TestRetirementActualProcessExits(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"receipt-created", "marker-created", "ticket-created", "receipt-written", "approved", "ticket-prepared", "marker-prepared", "marker-published", "before-hold-move", "hold-archived", "synced"} {
		t.Run(phase, func(t *testing.T) {
			s, id, hold, stage, a := retirementInspectionFixture(t, "cancelled")
			r, err := InspectCancelledRetirement(s.dir, id, hold, stage)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(filepath.Dir(s.dir), "owned-retirement-inspection.json")
			body, err := json.Marshal(r)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(path, body, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(self, "-test.run=^TestRetirementCrashWorker$", "-test.count=1")
			cmd.Env = append(os.Environ(), "OPENEDR_RETIRE_CRASH_WORKER=1", "OPENEDR_RETIRE_ROOT="+s.dir, "OPENEDR_RETIRE_INSPECTION="+path, "OPENEDR_RETIRE_PHASE=retirement-"+phase)
			output, err := cmd.CombinedOutput()
			if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 97 {
				t.Fatalf("retirement did not exit97: %v %s", err, output)
			}
			beforeRetry := inspectionTree(t, filepath.Dir(s.dir))
			_, err = RetireCancelledRecovery(s.dir, r, "retain every object after process exit", "")
			partial := phase == "receipt-created" || phase == "marker-created" || phase == "ticket-created"
			if (err != nil) != partial {
				t.Fatal("retirement process retry adopted a partial decision or refused complete evidence", err)
			}
			if partial && !reflect.DeepEqual(beforeRetry, inspectionTree(t, filepath.Dir(s.dir))) {
				t.Fatal("partial decision retry changed retained evidence")
			}
			want := ErrRecoveryRetired
			if partial {
				want = ErrRecoveryHeld
			}
			if err = s.Restore(id); !errors.Is(err, want) {
				t.Fatal("process exit removed ordinary restore protection", err)
			}
			if err = s.Purge(id); !errors.Is(err, want) {
				t.Fatal("process exit authorized deletion", err)
			}
			requirePurgePayload(t, filepath.Join(s.dir, id+".quarantined"), "payload")
			requirePurgePayload(t, filepath.Join(stage, "snapshot"), "payload")
			requirePurgePayload(t, filepath.Join(string(a.Inspection.Destination.ParentPath), ".edr-restore-"+a.OperationID, "snapshot"), "payload")
			destination := filepath.Join(string(a.Inspection.Destination.ParentPath), string(a.Inspection.Destination.Leaf))
			if _, err = os.Lstat(destination); !os.IsNotExist(err) {
				t.Fatal("retirement process exit published a destination", err)
			}
		})
	}
}

func TestRetirementCLIFixtureWorker(t *testing.T) {
	if os.Getenv("OPENEDR_RETIRE_CLI_FIXTURE_WORKER") != "1" {
		t.Skip("owned retirement CLI fixture worker")
	}
	control := os.Getenv("OPENEDR_RETIRE_CLI_FIXTURE_CONTROL")
	if !filepath.IsAbs(control) {
		t.Fatal("owned CLI control must be absolute")
	}
	s, id, hold, stage, a := retirementInspectionFixture(t, "cancelled")
	if err := s.cache.Close(); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(string(a.Inspection.Destination.ParentPath), string(a.Inspection.Destination.Leaf))
	body, err := json.Marshal(map[string]any{"root": s.dir, "id": id, "hold": hold, "stage": stage, "operation": a.OperationID, "destination": destination})
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(control, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write(body); err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		t.Fatal(err)
	}
	// The parent owns the disposable temp root and its eventual removal.
	os.Exit(0)
}

func TestRetirementEarlierRuntimeKeepsPermanentBarrier(t *testing.T) {
	testRetirementEarlierRuntime(t, os.Getenv("OPENEDR_CANCEL_EARLIER_PROGRAM"))
}

func TestRetirementEarlier102RuntimeKeepsPermanentBarrier(t *testing.T) {
	testRetirementEarlierRuntime(t, os.Getenv("OPENEDR_RETIRE_EARLIER102_PROGRAM"))
}

func testRetirementEarlierRuntime(t *testing.T, program string) {
	t.Helper()
	if program == "" {
		t.Skip("requires the pinned earlier production cancellation API worker")
	}
	s, id, stage, a, _, _ := retireFixture(t, "cancelled")
	for _, path := range []string{filepath.Join(stage, "snapshot"), filepath.Join(string(a.Inspection.Destination.ParentPath), ".edr-restore-"+a.OperationID, "snapshot")} {
		before := inspectionTree(t, filepath.Dir(s.dir))
		cmd := exec.Command(program, "-test.run=^TestEarlierCancelledOperationWorker$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), "OPENEDR_EARLIER_CANCEL_WORKER=1", "OPENEDR_EARLIER_CANCEL_ROOT="+s.dir, "OPENEDR_EARLIER_CANCEL_ID="+id, "OPENEDR_EARLIER_CANCEL_STAGE="+path)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("earlier API ignored permanent retirement guard: %v %s", err, output)
		}
		if !reflect.DeepEqual(before, inspectionTree(t, filepath.Dir(s.dir))) {
			t.Fatal("earlier API changed retirement evidence")
		}
		requirePurgePayload(t, path, "payload")
	}
}
