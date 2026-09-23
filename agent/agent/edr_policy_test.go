package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/config"
	edrconfig "github.com/utmstack/UTMStack/agent/edr/config"
)

// policyResp is the parsed shape of the policy_set EdrResult payload. Raw is
// the exact JSON bytes returned, so tests can assert [] vs null directly.
type policyResp struct {
	Applied         bool            `json:"applied"`
	Version         string          `json:"version"`
	AppliedLive     json.RawMessage `json:"applied_live"`
	RequiresRestart json.RawMessage `json:"requires_restart"`
	Drift           json.RawMessage `json:"drift"`
	Raw             []byte
}

// runPolicySet executes one policy_set dispatch and parses the response into
// its documented fields. The binary path is irrelevant for policy_set (no CLI
// or service is consulted through it); missingBin keeps it obviously inert.
func runPolicySet(t *testing.T, e *edrEnv, payload string) *policyResp {
	t.Helper()
	ok, errStr, data := edrDispatch("policy_set", payload, missingBin(t), &fakeEDRCLI{})
	if !ok || errStr != "" || data == "" {
		t.Fatalf("policy_set got ok=%v err=%q data=%q", ok, errStr, data)
	}
	var r policyResp
	if err := json.Unmarshal([]byte(data), &r); err != nil {
		t.Fatalf("policy_set payload is not JSON: %v (%q)", err, data)
	}
	r.Raw = []byte(data)
	return &r
}

// asStrings decodes a JSON array of strings; a null (absent list) counts as
// empty so assertions stay lenient about [] vs null.
func asStrings(t *testing.T, raw json.RawMessage) []string {
	t.Helper()
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("key %s is not a JSON array of strings: %v (%s)", string(raw), err, raw)
	}
	return out
}

func wantEmptyArray(t *testing.T, key string, raw json.RawMessage) {
	t.Helper()
	// The contract requires an explicit [] — null is a different JSON token.
	if len(raw) == 0 || string(raw) == "null" {
		t.Fatalf("%s=%s want [] (the arrays must never be null)", key, raw)
	}
	if len(asStrings(t, raw)) != 0 {
		t.Fatalf("%s=%s want []", key, raw)
	}
}

// isJSONArray reports whether a value decoded from map[string]any is a JSON
// array (any non-array type, or null, fails).
func isJSONArray(v any) bool {
	if v == nil {
		return false
	}
	_, ok := v.([]any)
	return ok
}

// mustRaw unwraps a value decoded from map[string]any into its raw JSON bytes.
func mustRaw(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal %v: %v", v, err)
	}
	return b
}

// writeEDRJSONState persists the given config and, when fp != nil, its
// post-merge doc (minus policy_version) as the fingerprint — i.e. the exact
// files edrPolicySet produces after a successful central apply.
func writeEDRJSONState(t *testing.T, dir string, c edrconfig.EDRConfig, fp *map[string]json.RawMessage) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "edr.json"), mustJSON(t, c), 0o600); err != nil {
		t.Fatal(err)
	}
	if fp != nil {
		delete(*fp, "policy_version")
		if err := os.WriteFile(filepath.Join(dir, "edr.policy-applied.json"), mustJSON(t, *fp), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func docOf(t *testing.T, c edrconfig.EDRConfig) map[string]json.RawMessage {
	t.Helper()
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(mustJSON(t, c), &doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// stubEDRFileEnv redirects every policy_set seam to real file-backed I/O under
// a temp dir: the config load/save round-trips edr.json, and the fingerprint
// and drift files are real files. Service control is captured, not executed.
// This lets the overlay, live/restart classification, and drift comparison run
// against genuine persisted state (hermetic — no install dir is touched).
// e.fileDir is exposed so callers know where the files live.
func stubEDRFileEnv(t *testing.T, e *edrEnv) {
	t.Helper()
	e.fileDir = t.TempDir()
	oldS, oldT, oldA, oldR := edrSvcStart, edrSvcStop, edrSvcActive, edrSvcRestart
	oldL, oldV, oldFP, oldDrift := edrCfgLoad, edrCfgSave, edrPolicyFingerprintFile, edrPolicyDriftFile
	edrSvcStart = func() error { e.started = true; return e.startErr }
	edrSvcStop = func() error { e.stopped = true; return e.stopErr }
	edrSvcActive = func() (bool, error) { return e.active, nil }
	edrSvcRestart = func() error { e.restarted = true; return e.restartErr }
	edrCfgLoad = func() (edrconfig.EDRConfig, error) {
		var c edrconfig.EDRConfig
		if b, err := os.ReadFile(filepath.Join(e.fileDir, "edr.json")); err == nil {
			_ = json.Unmarshal(b, &c)
		} else if !os.IsNotExist(err) {
			return edrconfig.Default(), err
		}
		return c, e.loadErr
	}
	edrCfgSave = func(c edrconfig.EDRConfig) error {
		e.saved = &c
		if err := os.WriteFile(filepath.Join(e.fileDir, "edr.json"), mustJSON(t, c), 0o600); err != nil {
			return err
		}
		return e.saveErr
	}
	edrPolicyFingerprintFile = filepath.Join(e.fileDir, "edr.policy-applied.json")
	edrPolicyDriftFile = filepath.Join(e.fileDir, "edr.policy-drift.json")
	t.Cleanup(func() {
		edrSvcStart, edrSvcStop, edrSvcActive, edrSvcRestart = oldS, oldT, oldA, oldR
		edrCfgLoad, edrCfgSave, edrPolicyFingerprintFile, edrPolicyDriftFile = oldL, oldV, oldFP, oldDrift
	})
}

func TestEdrPolicyLiveVsRestart(t *testing.T) {
	// base is the current effective state on disk; policy is the central
	// document to apply.
	base := edrconfig.Default()
	base.Enabled = true
	base.Allowlist.Paths = []string{`C:\base`}
	base.Allowlist.Processes = []string{"svchost.exe"}
	base.Ransomware.ResponseMode = "suspend"
	base.Sensors.FileWatcher = boolPtr(true)

	cases := []struct {
		name            string
		policy          map[string]any
		active          bool
		baseMod         func(*edrconfig.EDRConfig)
		wantLive        []string
		wantRestart     []string
		wantStart       bool
		wantStop        bool
		wantRestartCall bool
	}{
		{
			name:     "allowlist-only change is live, no restart",
			policy:   map[string]any{"allowlist": map[string]any{"paths": []string{`C:\new`}}},
			active:   false,
			wantLive: []string{"allowlist"},
		},
		{
			name:            "sensor change requires restart, service active",
			policy:          map[string]any{"sensors": map[string]any{"file_watcher": false}},
			active:          true,
			wantRestart:     []string{"sensors"},
			wantRestartCall: true,
		},
		{
			name:        "sensor change, service inactive — reported but not restarted",
			policy:      map[string]any{"sensors": map[string]any{"amsi": false}},
			active:      false,
			wantRestart: []string{"sensors"},
		},
		{
			name:      "enabled false to true starts the service, reported live",
			policy:    map[string]any{"enabled": true},
			baseMod:   func(c *edrconfig.EDRConfig) { c.Enabled = false },
			wantLive:  []string{"enabled"},
			wantStart: true,
		},
		{
			name:     "enabled true to false stops the service, reported live",
			policy:   map[string]any{"enabled": false},
			active:   true,
			wantLive: []string{"enabled"},
			wantStop: true,
		},
		{
			name:     "blocklist change is live",
			policy:   map[string]any{"blocklist": map[string]any{"refresh_hours": 12}},
			active:   false,
			wantLive: []string{"blocklist"},
		},
		{
			name:     "ransomware response_mode only is live",
			policy:   map[string]any{"ransomware": map[string]any{"response_mode": "kill"}},
			active:   false,
			wantLive: []string{"ransomware"},
		},
		{
			name:            "ransomware threshold change requires restart",
			policy:          map[string]any{"ransomware": map[string]any{"kill_threshold": 5}},
			active:          true,
			wantRestart:     []string{"ransomware"},
			wantRestartCall: true,
		},
		{
			name: "mixed live and restart keys are split correctly",
			policy: map[string]any{
				"allowlist": map[string]any{"commands": []string{"backup.exe"}},
				"sensors":   map[string]any{"behavioral": false},
			},
			active:          true,
			wantLive:        []string{"allowlist"},
			wantRestart:     []string{"sensors"},
			wantRestartCall: true,
		},
		{
			name:   "unchanged policy reports nothing",
			policy: map[string]any{"clamd_addr": "127.0.0.1:3310"},
			active: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := &edrEnv{active: tc.active}
			stubEDRFileEnv(t, e)
			b := base
			if tc.baseMod != nil {
				tc.baseMod(&b)
			}
			writeEDRJSONState(t, e.fileDir, b, nil)

			payload := map[string]any{"policy": tc.policy, "version": "vN"}
			r := runPolicySet(t, e, string(mustJSON(t, payload)))

			if len(tc.wantLive) > 0 {
				// Map iteration order is not guaranteed; compare as a set.
				wantSetEqual(t, "applied_live", asStrings(t, r.AppliedLive), tc.wantLive)
			} else {
				wantEmptyArray(t, "applied_live", r.AppliedLive)
			}
			if len(tc.wantRestart) > 0 {
				wantSetEqual(t, "requires_restart", asStrings(t, r.RequiresRestart), tc.wantRestart)
			} else {
				wantEmptyArray(t, "requires_restart", r.RequiresRestart)
			}
			wantEmptyArray(t, "drift", r.Drift)

			if tc.wantStart {
				if !e.started {
					t.Fatalf("service start was not called")
				}
			} else if e.started {
				t.Fatalf("service start was called unexpectedly")
			}
			if tc.wantStop {
				if !e.stopped {
					t.Fatalf("service stop was not called")
				}
			} else if e.stopped {
				t.Fatalf("service stop was called unexpectedly")
			}
			if tc.wantRestartCall {
				if !e.restarted {
					t.Fatalf("service restart was not called")
				}
			} else if e.restarted {
				t.Fatalf("service restart was called unexpectedly")
			}
		})
	}
}

// wantSetEqual compares two string slices as unordered sets.
func wantSetEqual(t *testing.T, key string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s=%v want %v", key, got, want)
	}
	seen := map[string]bool{}
	for _, g := range got {
		seen[g] = true
	}
	for _, w := range want {
		if !seen[w] {
			t.Fatalf("%s=%v want %v", key, got, want)
		}
	}
}

func boolPtr(b bool) *bool { return &b }

func TestEdrPolicyDriftLocalEditWinsCentral(t *testing.T) {
	e := &edrEnv{}
	stubEDRFileEnv(t, e)
	dir := e.fileDir

	// Last central policy: scan_concurrency=4, fail_mode=open.
	central := edrconfig.Default()
	central.SigUpdateHours = 8
	fp := docOf(t, central)
	writeEDRJSONState(t, dir, central, &fp)

	// A local hand-edit since then: someone raised the concurrency.
	local := central
	local.ScanConcurrency = 16
	if err := os.WriteFile(filepath.Join(dir, "edr.json"), mustJSON(t, local), 0o600); err != nil {
		t.Fatal(err)
	}

	// New central policy that does not touch scan_concurrency: the drifted
	// key's central value must come from the fingerprint, and the local edit
	// must be traced while the central document is still applied.
	r := runPolicySet(t, e, `{"policy":{"clamd_addr":"10.0.0.1:3310"},"version":"v9"}`)

	var drift []map[string]json.RawMessage
	if err := json.Unmarshal(r.Drift, &drift); err != nil {
		t.Fatalf("drift not a JSON array: %v (%s)", err, r.Drift)
	}
	if len(drift) != 1 {
		t.Fatalf("drift=%v want exactly 1 record (the local scan_concurrency edit)", drift)
	}
	var k string
	_ = json.Unmarshal(drift[0]["key"], &k)
	if k != "scan_concurrency" {
		t.Fatalf("drift key=%q want scan_concurrency: %v", k, drift)
	}
	var lw, ce any
	_ = json.Unmarshal(drift[0]["local_was"], &lw)
	_ = json.Unmarshal(drift[0]["central"], &ce)
	if lw != float64(16) || ce != float64(central.ScanConcurrency) {
		t.Fatalf("scan_concurrency drift local_was=%v central=%v want 16 / %d", lw, ce, central.ScanConcurrency)
	}
	if e.saved == nil || e.saved.ClamdAddr != "10.0.0.1:3310" {
		t.Fatalf("central policy was not applied to the saved config: %+v", e.saved)
	}
	if e.saved == nil || e.saved.ScanConcurrency != 16 {
		t.Fatalf("scan_concurrency=%d want the local value 16 (not central)", e.saved.ScanConcurrency)
	}

	// The fingerprint is overwritten with the new post-merge state.
	var newFP map[string]json.RawMessage
	if b, err := os.ReadFile(edrPolicyFingerprintFile); err != nil {
		t.Fatalf("fingerprint not overwritten: %v", err)
	} else if err := json.Unmarshal(b, &newFP); err != nil {
		t.Fatalf("fingerprint not overwritten: %v", err)
	}
	var fpClamd string
	_ = json.Unmarshal(newFP["clamd_addr"], &fpClamd)
	if fpClamd != "10.0.0.1:3310" {
		t.Fatalf("fingerprint clamd_addr=%q want the new central value", fpClamd)
	}
	if _, ok := newFP["policy_version"]; ok {
		t.Fatalf("fingerprint must not carry policy_version: %v", newFP)
	}

	// The drift trace is persisted next to the fingerprint.
	if _, err := os.Stat(edrPolicyDriftFile); err != nil {
		t.Fatalf("drift trace file missing: %v", err)
	}
}

// TestEdrPolicyArraysNeverNull pins the JSON contract: when nothing is
// reported, the raw payload must carry explicit [] for applied_live,
// requires_restart and drift — never null.
func TestEdrPolicyArraysNeverNull(t *testing.T) {
	e := &edrEnv{}
	stubEDRFileEnv(t, e)
	dir := e.fileDir
	base := edrconfig.Default()
	writeEDRJSONState(t, dir, base, nil)

	r := runPolicySet(t, e, `{"policy":{"clamd_addr":"127.0.0.1:3310"},"version":"v1"}`)
	if !strings.Contains(string(r.Raw), `"applied_live":[]`) ||
		!strings.Contains(string(r.Raw), `"requires_restart":[]`) ||
		!strings.Contains(string(r.Raw), `"drift":[]`) {
		t.Fatalf("arrays must be [], not null: %s", r.Raw)
	}
}

func TestEdrPolicyFirstApplyNoDrift(t *testing.T) {
	e := &edrEnv{}
	stubEDRFileEnv(t, e)
	dir := e.fileDir
	base := edrconfig.Default()
	writeEDRJSONState(t, dir, base, nil)

	r := runPolicySet(t, e, `{"policy":{"clamd_addr":"10.0.0.9:3310"},"version":"v1"}`)
	wantEmptyArray(t, "drift", r.Drift)
	if e.saved == nil || e.saved.PolicyVersion != "v1" {
		t.Fatalf("first policy_set did not persist the version: %+v", e.saved)
	}
	// A fingerprint now exists for the next comparison.
	if _, err := os.Stat(edrPolicyFingerprintFile); err != nil {
		t.Fatalf("fingerprint was not created: %v", err)
	}
}

func TestEdrPolicyValidationErrors(t *testing.T) {
	cases := []struct {
		name    string
		payload string
	}{
		{"empty policy object", `{"policy":{},"version":"v1"}`},
		{"version only", `{"version":"v1"}`},
		{"policy not an object", `{"policy":"nope"}`},
		{"payload not JSON", `not json`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := &edrEnv{}
			stubEDREnv(t, e)
			ok, errStr, data := edrDispatch("policy_set", tc.payload, missingBin(t), &fakeEDRCLI{})
			if ok || errStr != "validation_error" || data != "" {
				t.Fatalf("got ok=%v err=%q data=%q want validation_error", ok, errStr, data)
			}
			if e.saved != nil {
				t.Fatalf("no config should be saved for an invalid payload")
			}
			// No fingerprint or drift trace is written on validation failure.
			if _, err := os.Stat(edrPolicyFingerprintFile); !os.IsNotExist(err) {
				t.Fatalf("fingerprint should not exist: %v", err)
			}
		})
	}
}

func TestEdrPolicyRestartOnlyWhenActive(t *testing.T) {
	cases := []struct {
		name       string
		active     bool
		wantCalled bool
	}{
		{"service active", true, true},
		{"service inactive", false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := &edrEnv{active: tc.active}
			stubEDRFileEnv(t, e)
			dir := e.fileDir
			base := edrconfig.Default()
			writeEDRJSONState(t, dir, base, nil)

			r := runPolicySet(t, e, `{"policy":{"sensors":{"file_watcher":false}},"version":"v2"}`)
			wantSetEqual(t, "requires_restart", asStrings(t, r.RequiresRestart), []string{"sensors"})
			if e.restarted != tc.wantCalled {
				t.Fatalf("restarted=%v want %v", e.restarted, tc.wantCalled)
			}
		})
	}
}

// captureEDRSender is defined in edr_control_test.go; this test pins the
// processor path: a policy_set through edrCommandProcessor carries the full
// split in the EdrResult payload.
func TestEdrPolicyCommandProcessorPayload(t *testing.T) {
	e := &edrEnv{active: true}
	stubEDRFileEnv(t, e)
	dir := e.fileDir
	base := edrconfig.Default()
	writeEDRJSONState(t, dir, base, nil)

	sender := &captureEDRSender{}
	cnf := &config.Config{AgentID: 7}
	cmd := &EdrCommand{CmdId: "cmd-policy", Action: "policy_set", Payload: `{"policy":{"sensors":{"amsi":false},"allowlist":{"commands":["x.exe"]}},"version":"v5"}`}

	edrCommandProcessor(sender, cnf, cmd)

	if len(sender.sent) != 1 {
		t.Fatalf("sent %d streams, want 1", len(sender.sent))
	}
	res := sender.sent[0].GetEdrResult()
	if res == nil || !res.Ok {
		t.Fatalf("EdrResult missing or not ok: %v", res)
	}
	var r policyResp
	if err := json.Unmarshal([]byte(res.Payload), &r); err != nil {
		t.Fatalf("payload not JSON: %v (%q)", err, res.Payload)
	}
	if r.Version != "v5" {
		t.Fatalf("version=%q want v5", r.Version)
	}
	wantSetEqual(t, "applied_live", asStrings(t, r.AppliedLive), []string{"allowlist"})
	wantSetEqual(t, "requires_restart", asStrings(t, r.RequiresRestart), []string{"sensors"})
	wantEmptyArray(t, "drift", r.Drift)
	if !e.restarted {
		t.Fatalf("restart was not called")
	}
}
