package agent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/utmstack/UTMStack/agent/config"
	edrconfig "github.com/utmstack/UTMStack/agent/edr/config"
)

// fakeEDRCLI records the verb args it was invoked with and returns canned
// output. No real binary, subprocess, or service is ever touched.
type fakeEDRCLI struct {
	stdout string
	stderr string
	err    error
	calls  [][]string
}

func (f *fakeEDRCLI) run(ctx context.Context, args ...string) (string, string, error) {
	f.calls = append(f.calls, args)
	return f.stdout, f.stderr, f.err
}

// newFakeBin creates a temp file that passes fs.Exists so dispatch treats the
// module as installed.
func newFakeBin(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "utmstack_edr")
	if err := os.WriteFile(p, []byte("fake"), 0o700); err != nil {
		t.Fatal(err)
	}
	return p
}

// missingBin returns a path that does not exist.
func missingBin(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "does-not-exist")
}

// edrEnv captures the service-control and config-file calls so tests can
// assert on them.
type edrEnv struct {
	active   bool
	started  bool
	stopped  bool
	saved    *edrconfig.EDRConfig
	loadErr  error
	saveErr  error
	startErr error
	stopErr  error
}

// stubEDREnv redirects the service/config seams to the captured env and
// restores the originals on cleanup.
func stubEDREnv(t *testing.T, e *edrEnv) {
	t.Helper()
	oldS, oldT, oldA, oldL, oldV := edrSvcStart, edrSvcStop, edrSvcActive, edrCfgLoad, edrCfgSave
	edrSvcStart = func() error { e.started = true; return e.startErr }
	edrSvcStop = func() error { e.stopped = true; return e.stopErr }
	edrSvcActive = func() (bool, error) { return e.active, nil }
	edrCfgLoad = func() (edrconfig.EDRConfig, error) {
		return edrconfig.Default(), e.loadErr
	}
	edrCfgSave = func(c edrconfig.EDRConfig) error {
		e.saved = &c
		return e.saveErr
	}
	t.Cleanup(func() {
		edrSvcStart, edrSvcStop, edrSvcActive, edrCfgLoad, edrCfgSave = oldS, oldT, oldA, oldL, oldV
	})
}

func TestEDRDispatchCLIActions(t *testing.T) {
	cases := []struct {
		name     string
		action   string
		payload  string
		stdout   string
		wantArgs []string
		wantOK   bool
		wantData string
	}{
		{
			name:     "status",
			action:   "status",
			stdout:   `{"ok":true,"data":{"running":true}}`,
			wantArgs: []string{"status", "--json"},
			wantOK:   true,
			wantData: `{"running":true}`,
		},
		{
			name:     "quarantine_list",
			action:   "quarantine_list",
			stdout:   `{"ok":true,"data":{"items":[]}}`,
			wantArgs: []string{"quarantine", "list", "--json"},
			wantOK:   true,
			wantData: `{"items":[]}`,
		},
		{
			name:     "quarantine_restore",
			action:   "quarantine_restore",
			payload:  `{"id":"q123"}`,
			stdout:   `{"ok":true,"data":{"restored":"q123"}}`,
			wantArgs: []string{"quarantine", "restore", "q123", "--json"},
			wantOK:   true,
			wantData: `{"restored":"q123"}`,
		},
		{
			name:     "quarantine_purge_by_id",
			action:   "quarantine_purge",
			payload:  `{"id":"q123"}`,
			stdout:   `{"ok":true,"data":{"purged":"q123"}}`,
			wantArgs: []string{"quarantine", "purge", "q123", "--json"},
			wantOK:   true,
			wantData: `{"purged":"q123"}`,
		},
		{
			name:     "quarantine_purge_expired",
			action:   "quarantine_purge",
			payload:  `{"expired":true}`,
			stdout:   `{"ok":true,"data":{"purged":2}}`,
			wantArgs: []string{"quarantine", "purge", "--expired", "--json"},
			wantOK:   true,
			wantData: `{"purged":2}`,
		},
		{
			name:     "scan_path",
			action:   "scan_path",
			payload:  `{"path":"/tmp/x"}`,
			stdout:   `{"ok":true,"data":{"verdict":"clean"}}`,
			wantArgs: []string{"scan", "/tmp/x", "--json"},
			wantOK:   true,
			wantData: `{"verdict":"clean"}`,
		},
		{
			name:     "allow_add",
			action:   "allow_add",
			payload:  `{"category":"path","value":"/tmp/y"}`,
			stdout:   `{"ok":true,"data":{"added":"/tmp/y"}}`,
			wantArgs: []string{"allow", "path", "add", "/tmp/y", "--json"},
			wantOK:   true,
			wantData: `{"added":"/tmp/y"}`,
		},
		{
			name:     "allow_remove",
			action:   "allow_remove",
			payload:  `{"category":"network","value":"10.0.0.0/8"}`,
			stdout:   `{"ok":true,"data":{"removed":"10.0.0.0/8"}}`,
			wantArgs: []string{"allow", "network", "remove", "10.0.0.0/8", "--json"},
			wantOK:   true,
			wantData: `{"removed":"10.0.0.0/8"}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bin := newFakeBin(t)
			cli := &fakeEDRCLI{stdout: tc.stdout}
			ok, errStr, data := edrDispatch(tc.action, tc.payload, bin, cli)
			if ok != tc.wantOK || errStr != "" || data != tc.wantData {
				t.Fatalf("got ok=%v err=%q data=%q want ok=%v data=%q", ok, errStr, data, tc.wantOK, tc.wantData)
			}
			if len(cli.calls) != 1 {
				t.Fatalf("expected 1 CLI call, got %d (%v)", len(cli.calls), cli.calls)
			}
			if !reflect.DeepEqual(cli.calls[0], tc.wantArgs) {
				t.Fatalf("args=%v want=%v", cli.calls[0], tc.wantArgs)
			}
		})
	}
}

func TestEDRDispatchEnable(t *testing.T) {
	bin := newFakeBin(t)
	e := &edrEnv{}
	stubEDREnv(t, e)

	ok, errStr, _ := edrDispatch("enable", "", bin, &fakeEDRCLI{})
	if !ok || errStr != "" {
		t.Fatalf("enable got ok=%v err=%q", ok, errStr)
	}
	if !e.started {
		t.Fatalf("service start was not called")
	}
	if e.saved == nil || !e.saved.Enabled {
		t.Fatalf("config was not saved enabled")
	}
}

func TestEDRDispatchEnableServiceFails(t *testing.T) {
	bin := newFakeBin(t)
	e := &edrEnv{startErr: errors.New("systemctl exploded")}
	stubEDREnv(t, e)

	ok, errStr, _ := edrDispatch("enable", "", bin, &fakeEDRCLI{})
	if ok {
		t.Fatalf("enable should fail when service start fails")
	}
	if errStr != "systemctl exploded" {
		t.Fatalf("errStr=%q want the service error", errStr)
	}
	if !e.started {
		t.Fatalf("service start was not attempted")
	}
}

func TestEDRDispatchDisable(t *testing.T) {
	bin := newFakeBin(t)
	e := &edrEnv{}
	stubEDREnv(t, e)

	ok, errStr, _ := edrDispatch("disable", "", bin, &fakeEDRCLI{})
	if !ok || errStr != "" {
		t.Fatalf("disable got ok=%v err=%q", ok, errStr)
	}
	if !e.stopped {
		t.Fatalf("service stop was not called")
	}
	if e.saved == nil || e.saved.Enabled {
		t.Fatalf("config was not saved disabled")
	}
}

// A service that is not running is not an error: if the config save succeeds
// the disable still reports ok.
func TestEDRDispatchDisableServiceStopFailsButSaveOK(t *testing.T) {
	bin := newFakeBin(t)
	e := &edrEnv{stopErr: errors.New("unit not loaded")}
	stubEDREnv(t, e)

	ok, errStr, _ := edrDispatch("disable", "", bin, &fakeEDRCLI{})
	if !ok || errStr != "" {
		t.Fatalf("disable with stop error but ok save got ok=%v err=%q", ok, errStr)
	}
	if !e.stopped {
		t.Fatalf("service stop was not attempted")
	}
}

// If the config save also fails while the service was already down, the
// disable reports service_not_running.
func TestEDRDispatchDisableSaveFailsServiceDown(t *testing.T) {
	bin := newFakeBin(t)
	e := &edrEnv{active: false, saveErr: errors.New("no space left")}
	stubEDREnv(t, e)

	ok, errStr, _ := edrDispatch("disable", "", bin, &fakeEDRCLI{})
	if ok || errStr != "service_not_running" {
		t.Fatalf("disable got ok=%v err=%q want service_not_running", ok, errStr)
	}
}

// If the config save fails while the service WAS active, the disable reports
// the save error (the module could be disabled, only persistence failed).
func TestEDRDispatchDisableSaveFailsServiceActive(t *testing.T) {
	bin := newFakeBin(t)
	e := &edrEnv{active: true, saveErr: errors.New("no space left")}
	stubEDREnv(t, e)

	ok, errStr, _ := edrDispatch("disable", "", bin, &fakeEDRCLI{})
	if ok || errStr != "no space left" {
		t.Fatalf("disable got ok=%v err=%q want the save error", ok, errStr)
	}
}

func TestEDRDispatchPolicySet(t *testing.T) {
	e := &edrEnv{}
	stubEDREnv(t, e)

	ok, errStr, data := edrDispatch("policy_set", `{"policy":{"clamd_addr":"1.2.3.4:1234"},"version":"v2"}`, missingBin(t), &fakeEDRCLI{})
	if !ok || errStr != "" || data != `{"applied":true,"version":"v2"}` {
		t.Fatalf("policy_set got ok=%v err=%q data=%q", ok, errStr, data)
	}
	if e.saved == nil {
		t.Fatalf("config was not saved")
	}
	if e.saved.ClamdAddr != "1.2.3.4:1234" {
		t.Fatalf("ClamdAddr=%q want the overlaid value", e.saved.ClamdAddr)
	}
	// The policy document version is persisted so the status reporter can ship
	// it upstream for drift detection.
	if e.saved.PolicyVersion != "v2" {
		t.Fatalf("PolicyVersion=%q want v2", e.saved.PolicyVersion)
	}
	// A key absent from the policy must be left untouched (Default FailMode).
	if e.saved.FailMode != "open" {
		t.Fatalf("FailMode=%q want untouched default open", e.saved.FailMode)
	}
}

// A version without a policy document is not a valid payload.
func TestEDRDispatchPolicySetVersionOnly(t *testing.T) {
	e := &edrEnv{}
	stubEDREnv(t, e)

	ok, errStr, _ := edrDispatch("policy_set", `{"version":"v2"}`, missingBin(t), &fakeEDRCLI{})
	if ok || errStr != "validation_error" {
		t.Fatalf("policy_set version-only got ok=%v err=%q want validation_error", ok, errStr)
	}
	if e.saved != nil {
		t.Fatalf("no config should be saved for an invalid payload")
	}
}

// A non-object policy block is a validation error.
func TestEDRDispatchPolicySetPolicyNotObject(t *testing.T) {
	e := &edrEnv{}
	stubEDREnv(t, e)

	ok, errStr, _ := edrDispatch("policy_set", `{"policy":"nope"}`, missingBin(t), &fakeEDRCLI{})
	if ok || errStr != "validation_error" {
		t.Fatalf("policy_set non-object policy got ok=%v err=%q want validation_error", ok, errStr)
	}
}

func TestEDRDispatchPolicySetInvalidJSON(t *testing.T) {
	e := &edrEnv{}
	stubEDREnv(t, e)

	ok, errStr, _ := edrDispatch("policy_set", `not json`, missingBin(t), &fakeEDRCLI{})
	if ok || errStr != "validation_error" {
		t.Fatalf("policy_set bad json got ok=%v err=%q", ok, errStr)
	}
}

func TestEDRDispatchUnsupported(t *testing.T) {
	for _, action := range []string{
		"kill_process", "full_scan", "isolate", "release_isolation", "blocklist_refresh", "warp_drive", "",
	} {
		ok, errStr, _ := edrDispatch(action, "", missingBin(t), &fakeEDRCLI{})
		if ok || errStr != "unsupported_action" {
			t.Fatalf("action %q got ok=%v err=%q want unsupported_action", action, ok, errStr)
		}
	}
}

func TestEDRDispatchModuleNotInstalled(t *testing.T) {
	bin := missingBin(t)
	for _, action := range []string{"status", "quarantine_list", "scan_path", "enable", "disable"} {
		ok, errStr, _ := edrDispatch(action, `{"path":"/tmp/x","id":"q"}`, bin, &fakeEDRCLI{})
		if ok || errStr != "module_not_installed" {
			t.Fatalf("action %q got ok=%v err=%q want module_not_installed", action, ok, errStr)
		}
	}
}

func TestEDRDispatchValidationErrors(t *testing.T) {
	bin := newFakeBin(t)
	cases := []struct {
		name    string
		action  string
		payload string
	}{
		{"quarantine_restore_no_id", "quarantine_restore", `{}`},
		{"quarantine_restore_empty_id", "quarantine_restore", `{"id":""}`},
		{"scan_path_no_path", "scan_path", `{}`},
		{"allow_add_bad_category", "allow_add", `{"category":"bogus","value":"x"}`},
		{"allow_add_no_value", "allow_add", `{"category":"path"}`},
		{"quarantine_purge_neither", "quarantine_purge", `{}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cli := &fakeEDRCLI{}
			ok, errStr, _ := edrDispatch(tc.action, tc.payload, bin, cli)
			if ok || errStr != "validation_error" {
				t.Fatalf("got ok=%v err=%q want validation_error", ok, errStr)
			}
			if len(cli.calls) != 0 {
				t.Fatalf("no CLI call should be made, got %v", cli.calls)
			}
		})
	}
}

func TestEDRDispatchTimeout(t *testing.T) {
	bin := newFakeBin(t)
	cli := &fakeEDRCLI{err: context.DeadlineExceeded}

	ok, errStr, _ := edrDispatch("status", "", bin, cli)
	if ok || errStr != "timeout" {
		t.Fatalf("timeout got ok=%v err=%q want timeout", ok, errStr)
	}
}

// A non-zero exit is not authoritative: the envelope wins when parseable.
func TestEDRDispatchCLIEnvelopePassThrough(t *testing.T) {
	bin := newFakeBin(t)
	cli := &fakeEDRCLI{
		stdout: `{"ok":false,"error":"file already restored","data":{}}`,
		err:    errors.New("exit status 1"),
	}

	ok, errStr, _ := edrDispatch("quarantine_restore", `{"id":"q123"}`, bin, cli)
	if ok || errStr != "file already restored" {
		t.Fatalf("pass-through got ok=%v err=%q want the envelope error verbatim", ok, errStr)
	}
}

func TestEDRDispatchCLISuccessData(t *testing.T) {
	bin := newFakeBin(t)
	cli := &fakeEDRCLI{stdout: `{"ok":true,"data":{"restored":"q123"}}`}

	ok, errStr, data := edrDispatch("quarantine_restore", `{"id":"q123"}`, bin, cli)
	if !ok || errStr != "" || data != `{"restored":"q123"}` {
		t.Fatalf("success got ok=%v err=%q data=%q", ok, errStr, data)
	}
}

// Unparseable subprocess output with no stderr is reported as
// invalid_module_output.
func TestEDRDispatchCLIInvalidOutput(t *testing.T) {
	bin := newFakeBin(t)
	cli := &fakeEDRCLI{stdout: "this is not json", err: errors.New("exit status 1")}

	ok, errStr, _ := edrDispatch("status", "", bin, cli)
	if ok || errStr != "invalid_module_output" {
		t.Fatalf("invalid output got ok=%v err=%q want invalid_module_output", ok, errStr)
	}
}

// Unparseable subprocess output with stderr is still reported as
// invalid_module_output (the machine-readable token stays clean).
func TestEDRDispatchCLIInvalidOutputWithStderr(t *testing.T) {
	bin := newFakeBin(t)
	cli := &fakeEDRCLI{stdout: "garbage", stderr: "boom: kernel panic\nmore", err: errors.New("exit status 2")}

	ok, errStr, _ := edrDispatch("status", "", bin, cli)
	if ok || errStr != "invalid_module_output" {
		t.Fatalf("stderr present got ok=%v err=%q want invalid_module_output", ok, errStr)
	}
}

// captureEDRSender implements resultSender and records what was sent.
type captureEDRSender struct {
	sent []*BidirectionalStream
}

func (c *captureEDRSender) Send(m *BidirectionalStream) error {
	c.sent = append(c.sent, m)
	return nil
}

func TestEDRCommandProcessorSendsResult(t *testing.T) {
	sender := &captureEDRSender{}
	cnf := &config.Config{AgentID: 1}
	cmd := &EdrCommand{CmdId: "cmd-xyz", Action: "kill_process"}

	edrCommandProcessor(sender, cnf, cmd)

	if len(sender.sent) != 1 {
		t.Fatalf("sent %d streams, want 1", len(sender.sent))
	}
	res := sender.sent[0].GetEdrResult()
	if res == nil {
		t.Fatalf("sent stream was not an EdrResult: %v", sender.sent[0].StreamMessage)
	}
	if res.CmdId != "cmd-xyz" {
		t.Fatalf("cmd_id=%q want cmd-xyz", res.CmdId)
	}
	if res.AgentId != "1" {
		t.Fatalf("agent_id=%q want 1", res.AgentId)
	}
	if res.ExecutedAt == nil {
		t.Fatalf("executed_at is nil")
	}
	if res.Ok {
		t.Fatalf("ok should be false for an unsupported action")
	}
	if res.Error != "unsupported_action" {
		t.Fatalf("error=%q want unsupported_action", res.Error)
	}
}
