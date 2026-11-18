package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/quarantine"
)

func TestQuarantineFreshRealMainCancellationAndExactRetry(t *testing.T) {
	program := os.Getenv("OPENEDR_CURRENT_QUARANTINE_PROGRAM")
	if program == "" {
		t.Skip("requires the pinned current quarantine test program")
	}
	root := t.TempDir()
	control := filepath.Join(root, "fixture.json")
	setup := exec.Command(program, "-test.run=^TestFreshCLIFixtureWorker$", "-test.count=1")
	setup.Env = append(os.Environ(), "OPENEDR_FRESH_CLI_FIXTURE_WORKER=1", "OPENEDR_FRESH_CLI_FIXTURE_MODE=cancel", "OPENEDR_FRESH_CLI_FIXTURE_CONTROL="+control, "TMPDIR="+root, "TMP="+root, "TEMP="+root, "SystemTemp="+root+string(os.PathSeparator))
	if output, err := setup.CombinedOutput(); err != nil {
		t.Fatalf("owned fresh cancellation CLI fixture: %v %s", err, output)
	}
	var fixture struct {
		Root, ID, Operation, Destination string
		Inspection                       quarantine.FreshRestoreInspection
	}
	body, err := os.ReadFile(control)
	if err != nil || json.Unmarshal(body, &fixture) != nil {
		t.Fatal("invalid cancellation fixture", err)
	}
	relative, err := filepath.Rel(root, fixture.Root)
	if err != nil || !filepath.IsLocal(relative) {
		t.Fatal("fixture escaped the owned temporary root", err)
	}
	cfg, _ := json.Marshal(map[string]string{"quarantine_dir": fixture.Root})
	if err = os.WriteFile(filepath.Join(root, "edr.json"), cfg, 0600); err != nil {
		t.Fatal(err)
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	run := func(args []string) []byte {
		t.Helper()
		raw, e := json.Marshal(args)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(root, "command.json"), raw, 0600); e != nil {
			t.Fatal(e)
		}
		cmd := exec.Command(self, "-test.run=^TestQuarantineFreshMainWorker$", "-test.count=1")
		cmd.Env = append(os.Environ(), "OPENEDR_FRESH_MAIN_WORKER=1", "OPENEDR_FRESH_MAIN_ROOT="+root, "OPENEDR_FRESH_MAIN_DATABASE="+filepath.Join(root, "must-not-open-db"))
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if e = cmd.Run(); e != nil || stderr.Len() != 0 {
			t.Fatalf("standalone fresh cancellation main failed: %v stdout=%s stderr=%s", e, &stdout, &stderr)
		}
		for _, name := range []string{"must-not-open-db", "must-not-create-log"} {
			if _, e = os.Lstat(filepath.Join(root, name)); !os.IsNotExist(e) {
				t.Fatal("retention command initialized database or log", name, e)
			}
		}
		return stdout.Bytes()
	}
	saved := run([]string{"inspect-fresh-cancel", fixture.ID, "--operation", fixture.Operation})
	var inspection quarantine.FreshCancellationInspection
	if err = json.Unmarshal(saved, &inspection); err != nil || inspection.OperationID != fixture.Operation || inspection.Journal.Status != "captured" {
		t.Fatal("invalid CLI cancellation inspection", err)
	}
	reviewPath := filepath.Join(root, "saved-fresh-cancellation.json")
	if err = os.WriteFile(reviewPath, saved, 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"cancel-fresh", fixture.ID, "--operation", fixture.Operation, "--inspection", reviewPath, "--reason", "independent CLI fresh cancellation"}
	var first quarantine.FreshRestoreCancellation
	for attempt := 0; attempt < 2; attempt++ {
		var result quarantine.FreshRestoreCancellation
		if err = json.Unmarshal(run(args), &result); err != nil || result.OperationID != fixture.Operation || result.Kind != "fresh-restore-cancellation" {
			t.Fatal("invalid CLI fresh cancellation", err)
		}
		if attempt == 0 {
			first = result
		} else if !reflect.DeepEqual(first, result) {
			t.Fatal("CLI cancellation retry changed its decision")
		}
	}
	archive := filepath.Join(fixture.Root, ".fresh-restores", fixture.ID, fixture.Operation, "restore")
	if body, err = os.ReadFile(filepath.Join(archive, "unknown-owned-cli-evidence")); err != nil || string(body) != "preserve CLI evidence" {
		t.Fatal("CLI cancellation lost unknown bytes", err)
	}
	if _, err = os.Lstat(fixture.Destination); !os.IsNotExist(err) {
		t.Fatal("CLI cancellation published a destination", err)
	}
	report, err := quarantine.InspectRecovery(fixture.Root, fixture.ID)
	if err != nil || report.Journals[1].Status != "absent" || report.Journals[2].ManifestID != fixture.Inspection.Hold.ManifestID || report.Journals[2].SHA256 != fixture.Inspection.Hold.SHA256 {
		t.Fatal("CLI cancellation changed original fence", err)
	}
}
