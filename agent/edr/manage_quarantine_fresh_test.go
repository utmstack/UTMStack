package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/quarantine"
)

func TestQuarantineFreshMainWorker(t *testing.T) {
	if os.Getenv("OPENEDR_FRESH_MAIN_WORKER") != "1" {
		t.Skip("owned fresh approval/execution main worker")
	}
	root := os.Getenv("OPENEDR_FRESH_MAIN_ROOT")
	config.ConfigFile = filepath.Join(root, "edr.json")
	config.LogFile = filepath.Join(root, "must-not-create-log")
	config.DBFile = os.Getenv("OPENEDR_FRESH_MAIN_DATABASE")
	var args []string
	body, err := os.ReadFile(filepath.Join(root, "command.json"))
	if err != nil || json.Unmarshal(body, &args) != nil {
		t.Fatal("invalid owned CLI command", err)
	}
	os.Args = append([]string{"utmstack_edr", "quarantine"}, args...)
	main()
	t.Fatal("fresh main must exit via its standalone route")
}

func TestQuarantineFreshRealMainApprovalExecutionAndExactRetry(t *testing.T) {
	program := os.Getenv("OPENEDR_CURRENT_QUARANTINE_PROGRAM")
	if program == "" {
		t.Skip("requires the pinned current quarantine test program")
	}
	root := t.TempDir()
	control := filepath.Join(root, "fixture.json")
	setup := exec.Command(program, "-test.run=^TestFreshCLIFixtureWorker$", "-test.count=1")
	setup.Env = append(os.Environ(), "OPENEDR_FRESH_CLI_FIXTURE_WORKER=1", "OPENEDR_FRESH_CLI_FIXTURE_CONTROL="+control, "TMPDIR="+root, "TMP="+root, "TEMP="+root, "SystemTemp="+root+string(os.PathSeparator))
	if output, err := setup.CombinedOutput(); err != nil {
		t.Fatalf("owned CLI fixture worker: %v %s", err, output)
	}
	var fixture struct {
		Root        string                            `json:"root"`
		Database    string                            `json:"database"`
		ID          string                            `json:"id"`
		Destination string                            `json:"destination"`
		Inspection  quarantine.FreshRestoreInspection `json:"inspection"`
	}
	body, err := os.ReadFile(control)
	if err != nil || json.Unmarshal(body, &fixture) != nil {
		t.Fatal("invalid owned CLI fixture", err)
	}
	if !filepath.IsAbs(fixture.Root) || !filepath.IsAbs(fixture.Database) || !filepath.IsAbs(fixture.Destination) {
		t.Fatal("fixture paths are not absolute")
	}
	inspection, _ := json.Marshal(fixture.Inspection)
	cfg, _ := json.Marshal(map[string]string{"quarantine_dir": fixture.Root})
	for name, body := range map[string][]byte{"inspection.json": inspection, "edr.json": cfg} {
		if err = os.WriteFile(filepath.Join(root, name), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	run := func(args []string, database string) []byte {
		t.Helper()
		body, e := json.Marshal(args)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(root, "command.json"), body, 0600); e != nil {
			t.Fatal(e)
		}
		cmd := exec.Command(self, "-test.run=^TestQuarantineFreshMainWorker$", "-test.count=1")
		cmd.Env = append(os.Environ(), "OPENEDR_FRESH_MAIN_WORKER=1", "OPENEDR_FRESH_MAIN_ROOT="+root, "OPENEDR_FRESH_MAIN_DATABASE="+database)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if e = cmd.Run(); e != nil || stderr.Len() != 0 {
			t.Fatalf("real fresh main failed: %v stdout=%s stderr=%s", e, &stdout, &stderr)
		}
		if _, e = os.Lstat(filepath.Join(root, "must-not-create-log")); !os.IsNotExist(e) {
			t.Fatal("standalone fresh command initialized logging", e)
		}
		return stdout.Bytes()
	}
	operation := uuid.NewString()
	approved := run([]string{"approve-fresh", fixture.ID, "--operation", operation, "--inspection", filepath.Join(root, "inspection.json"), "--reason", "independent CLI review"}, filepath.Join(root, "must-not-open-db"))
	if _, err = os.Lstat(filepath.Join(root, "must-not-open-db")); !os.IsNotExist(err) {
		t.Fatal("approval opened a database", err)
	}
	var approval quarantine.FreshRestoreApproval
	if err = json.Unmarshal(approved, &approval); err != nil || approval.OperationID != operation {
		t.Fatal("invalid CLI approval", err)
	}
	if err = os.WriteFile(filepath.Join(root, "approval.json"), approved, 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"execute-fresh", fixture.ID, "--approval", filepath.Join(root, "approval.json"), "--reason", "execute independent CLI review"}
	var first quarantine.FreshRestoreResult
	for attempt := 0; attempt < 2; attempt++ {
		var result quarantine.FreshRestoreResult
		if err = json.Unmarshal(run(args, fixture.Database), &result); err != nil || result.OperationID != operation || result.Phase != "committed" || string(result.Destination) != fixture.Destination || result.OutputID == "" {
			t.Fatal("invalid CLI execution result", result, err)
		}
		if attempt == 0 {
			first = result
		} else if !reflect.DeepEqual(first, result) {
			t.Fatal("lost-return CLI retry changed publication proof")
		}
	}
	if payload, err := os.ReadFile(fixture.Destination); err != nil || string(payload) != "payload" {
		t.Fatal("CLI did not publish approved contents", err)
	}
	report, err := quarantine.InspectRecovery(fixture.Root, fixture.ID)
	if err != nil || report.Journals[2].ManifestID != fixture.Inspection.Hold.ManifestID || report.Journals[2].SHA256 != fixture.Inspection.Hold.SHA256 {
		t.Fatal("CLI changed original retention fence", err)
	}
}

func TestQuarantineFreshCLIRejectsIncompleteArguments(t *testing.T) {
	for _, action := range []func([]string, *bytes.Buffer, *bytes.Buffer) int{
		func(a []string, o, d *bytes.Buffer) int { return runQuarantineFreshApproval(a, o, d) },
		func(a []string, o, d *bytes.Buffer) int { return runQuarantineFreshExecution(a, o, d) },
		func(a []string, o, d *bytes.Buffer) int { return runQuarantineFreshCancellationInspection(a, o, d) },
		func(a []string, o, d *bytes.Buffer) int { return runQuarantineFreshCancellation(a, o, d) },
		func(a []string, o, d *bytes.Buffer) int { return runQuarantineRetirementInspection(a, o, d) },
		func(a []string, o, d *bytes.Buffer) int { return runQuarantineRetirement(a, o, d) },
	} {
		for _, args := range [][]string{nil, {"id"}, {"id", "--wrong", "x", "--reason", "text"}} {
			var out, diagnostic bytes.Buffer
			if code := action(args, &out, &diagnostic); code != 1 || out.Len() != 0 || diagnostic.Len() == 0 {
				t.Fatal("invalid fresh command arguments accepted", args, code)
			}
		}
	}
}
