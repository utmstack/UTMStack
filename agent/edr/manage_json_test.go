package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

// ---- test helpers ----

// captureStdout runs fn and returns whatever it writes to os.Stdout, restoring
// the original stream afterwards.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()
	fn()
	w.Close()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// withConfigFile points config.ConfigFile at path for the duration of the test,
// restoring it on cleanup. Keeps config Load/Save hermetic.
func withConfigFile(t *testing.T, path string) {
	t.Helper()
	old := config.ConfigFile
	config.ConfigFile = path
	t.Cleanup(func() { config.ConfigFile = old })
}

func setJSONMode(t *testing.T, on bool) {
	t.Helper()
	old := jsonMode
	jsonMode = on
	t.Cleanup(func() { jsonMode = old })
}

// ---- 1. flag parsing ----

func TestParseJSONFlag(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantOn  bool
		wantOut []string
	}{
		{"absent", []string{"config", "show"}, false, []string{"config", "show"}},
		{"first", []string{"--json", "config", "show"}, true, []string{"config", "show"}},
		{"last", []string{"config", "show", "--json"}, true, []string{"config", "show"}},
		{"mid", []string{"config", "--json", "show"}, true, []string{"config", "show"}},
		{
			// must not swallow the value that follows it
			"no_value_eaten", []string{"set", "key", "--json", "value"}, true, []string{"set", "key", "value"},
		},
		{"twice", []string{"--json", "scan", "--json", "/x"}, true, []string{"scan", "/x"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			on, out := parseJSONFlag(tc.args)
			if on != tc.wantOn {
				t.Fatalf("on = %v, want %v", on, tc.wantOn)
			}
			if len(out) != len(tc.wantOut) {
				t.Fatalf("out = %v, want %v", out, tc.wantOut)
			}
			for i := range out {
				if out[i] != tc.wantOut[i] {
					t.Fatalf("out = %v, want %v", out, tc.wantOut)
				}
			}
		})
	}
}

// ---- 2. envelope shape ----

func TestEmitJSONSuccessHasNoErrorKey(t *testing.T) {
	setJSONMode(t, true)
	out := captureStdout(t, func() {
		emitJSON(true, "", map[string]interface{}{"a": 1})
	})
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("not JSON: %v (%q)", err, out)
	}
	if _, has := m["error"]; has {
		t.Fatalf("success envelope must omit error key, got %v", m)
	}
	if m["ok"] != true {
		t.Fatalf("ok = %v, want true", m["ok"])
	}
	if len(m) != 2 {
		t.Fatalf("success envelope must have exactly ok+data, got %v", m)
	}
	data, isObj := m["data"].(map[string]interface{})
	if !isObj || data["a"] != float64(1) {
		t.Fatalf("data = %v, want {a:1}", m["data"])
	}
}

func TestEmitJSONFailureHasErrorAndNonNullData(t *testing.T) {
	setJSONMode(t, true)
	out := captureStdout(t, func() {
		emitJSON(false, "boom", nil)
	})
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("not JSON: %v (%q)", err, out)
	}
	if m["ok"] != false {
		t.Fatalf("ok = %v, want false", m["ok"])
	}
	if m["error"] != "boom" {
		t.Fatalf("error = %v, want boom", m["error"])
	}
	if len(m) != 3 {
		t.Fatalf("failure envelope must have exactly ok+error+data, got %v", m)
	}
	if _, isObj := m["data"].(map[string]interface{}); !isObj {
		t.Fatalf("data must be a non-null object, got %v", m["data"])
	}
}

func TestEmitJSONNoopWhenFlagAbsent(t *testing.T) {
	setJSONMode(t, false)
	out := captureStdout(t, func() {
		emitJSON(true, "", map[string]interface{}{"a": 1})
	})
	if out != "" {
		t.Fatalf("emitJSON must be silent without --json, got %q", out)
	}
}

// ---- 3. config group ----

func TestApplySetSignatureFallback(t *testing.T) {
	for _, v := range []string{"", "cdn", "none"} {
		c := config.Default()
		if err := applySet(&c, "signature_fallback", v); err != nil {
			t.Fatalf("signature_fallback=%q: %v", v, err)
		}
		if c.SignatureFallback != v {
			t.Fatalf("SignatureFallback = %q, want %q", c.SignatureFallback, v)
		}
	}
	c := config.Default()
	if err := applySet(&c, "signature_fallback", "bogus"); err == nil {
		t.Fatal("bogus signature_fallback must error")
	}
	// A URL must be rejected — mirror URLs belong to signature_mirror.
	if err := applySet(&c, "signature_fallback", "http://x/y"); err == nil {
		t.Fatal("URL signature_fallback must error")
	}
}

func TestConfigGetDataKnownAndUnknown(t *testing.T) {
	withConfigFile(t, filepath.Join(t.TempDir(), "edr.json")) // absent -> defaults

	data, err := configGetData("signature_fallback")
	if err != nil {
		t.Fatalf("known key: %v", err)
	}
	m, _ := data.(map[string]interface{})
	if m["key"] != "signature_fallback" || m["value"] != "cdn" {
		t.Fatalf("data = %v, want key=signature_fallback value=cdn", m)
	}

	if _, err := configGetData("does.not.exist"); err == nil {
		t.Fatal("unknown key must error")
	}
}

func TestConfigSetDataApplies(t *testing.T) {
	dir := t.TempDir()
	withConfigFile(t, filepath.Join(dir, "edr.json"))

	data, err := configSetData("signature_fallback", "none")
	if err != nil {
		t.Fatal(err)
	}
	m, _ := data.(map[string]interface{})
	if m["key"] != "signature_fallback" || m["value"] != "none" || m["applied"] != true {
		t.Fatalf("data = %v", m)
	}
	// persisted to disk
	cfg, _ := config.Load()
	if cfg.SignatureFallback != "none" {
		t.Fatalf("saved SignatureFallback = %q, want none", cfg.SignatureFallback)
	}

	if _, err := configSetData("signature_fallback", "bogus"); err == nil {
		t.Fatal("bad value must error")
	}
}

// ---- 4. allow group ----

func TestAllowAddRemoveListData(t *testing.T) {
	dir := t.TempDir()
	withConfigFile(t, filepath.Join(dir, "edr.json"))

	d, err := allowAddData("path", "/tmp/foo")
	if err != nil {
		t.Fatal(err)
	}
	m, _ := d.(map[string]interface{})
	if m["count"] != 1 {
		t.Fatalf("count = %v, want 1", m["count"])
	}
	if _, dup := m["already_present"]; dup {
		t.Fatalf("first add must not flag already_present: %v", m)
	}

	d, err = allowAddData("path", "/tmp/foo")
	if err != nil {
		t.Fatal(err)
	}
	m, _ = d.(map[string]interface{})
	if m["already_present"] != true {
		t.Fatalf("second add must flag already_present: %v", m)
	}
	if m["count"] != 1 {
		t.Fatalf("count = %v, want 1 (no dup)", m["count"])
	}

	d, err = allowListData("path")
	if err != nil {
		t.Fatal(err)
	}
	m, _ = d.(map[string]interface{})
	if m["count"] != 1 {
		t.Fatalf("list count = %v, want 1", m["count"])
	}
	entries, _ := m["entries"].([]string)
	if len(entries) != 1 || entries[0] != "/tmp/foo" {
		t.Fatalf("entries = %v, want [/tmp/foo]", m["entries"])
	}

	d, err = allowRemoveData("path", "/tmp/foo")
	if err != nil {
		t.Fatal(err)
	}
	m, _ = d.(map[string]interface{})
	if m["count"] != 0 {
		t.Fatalf("count after remove = %v, want 0", m["count"])
	}

	if _, err := allowRemoveData("path", "/tmp/foo"); err == nil {
		t.Fatal("removing missing entry must error")
	}
}

func TestAllowNetworkValidation(t *testing.T) {
	dir := t.TempDir()
	withConfigFile(t, filepath.Join(dir, "edr.json"))

	if _, err := allowAddData("network", "not-an-ip"); err == nil {
		t.Fatal("non-IP network entry must error")
	}
	d, err := allowAddData("network", "10.0.0.0/8")
	if err != nil {
		t.Fatal(err)
	}
	m, _ := d.(map[string]interface{})
	if m["category"] != "network" {
		t.Fatalf("category = %v, want network", m["category"])
	}
}

// ---- 5. quarantine record -> JSON mapping ----

func TestQuarItemDataStateMapping(t *testing.T) {
	base := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	cases := []struct {
		name  string
		rec   cache.QuarantineRecord
		state string
	}{
		{"held", cache.QuarantineRecord{
			QuarantineID: "id1", OriginalPath: "/p/1", Detection: "mal", QuarantinedAt: base,
		}, "held"},
		{"restored", cache.QuarantineRecord{
			QuarantineID: "id2", OriginalPath: "/p/2", Detection: "mal", QuarantinedAt: base, Restored: true,
		}, "restored"},
		{"purged", cache.QuarantineRecord{
			QuarantineID: "id3", OriginalPath: "/p/3", Detection: "mal", QuarantinedAt: base, Purged: true,
		}, "purged"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := quarItemData(tc.rec)
			if m["state"] != tc.state {
				t.Fatalf("state = %v, want %s", m["state"], tc.state)
			}
			if m["id"] != tc.rec.QuarantineID {
				t.Fatalf("id = %v", m["id"])
			}
			if m["path"] != tc.rec.OriginalPath {
				t.Fatalf("path = %v", m["path"])
			}
			if m["date"] != base.Format("2006-01-02 15:04:05") {
				t.Fatalf("date = %v", m["date"])
			}
		})
	}
}

// ---- 6. human output unchanged ----

func TestHumanAllowListEmptyUnchanged(t *testing.T) {
	setJSONMode(t, false)
	withConfigFile(t, filepath.Join(t.TempDir(), "edr.json"))
	out := captureStdout(t, func() { runAllow([]string{"path", "list"}) })
	if out != "(no user path entries; built-in defaults still apply)\n" {
		t.Fatalf("human allow list output changed: %q", out)
	}
}

func TestHumanConfigGetUnchanged(t *testing.T) {
	setJSONMode(t, false)
	withConfigFile(t, filepath.Join(t.TempDir(), "edr.json"))
	out := captureStdout(t, func() { runConfig([]string{"get", "enabled"}) })
	if out != "false\n" {
		t.Fatalf("human config get output changed: %q", out)
	}
}

func TestHumanStatusNotRunningUnchanged(t *testing.T) {
	setJSONMode(t, false)
	old := config.StatusFile
	config.StatusFile = filepath.Join(t.TempDir(), "status.json") // absent
	defer func() { config.StatusFile = old }()
	out := captureStdout(t, runStatus)
	if out != "UTMStack EDR: not running\n" {
		t.Fatalf("human status output changed: %q", out)
	}
}

// ---- status / help / lifecycle data shapes ----

func TestStatusDataMissingFile(t *testing.T) {
	old := config.StatusFile
	config.StatusFile = filepath.Join(t.TempDir(), "status.json") // absent
	defer func() { config.StatusFile = old }()

	data, err := statusData()
	if err != nil {
		t.Fatal(err)
	}
	m, _ := data.(map[string]interface{})
	if m["running"] != false {
		t.Fatalf("data = %v, want {running:false}", m)
	}
}

func TestStatusDataParsesJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "status.json")
	if err := os.WriteFile(p, []byte(`{"running":true,"uptime":"3d"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	old := config.StatusFile
	config.StatusFile = p
	defer func() { config.StatusFile = old }()

	data, err := statusData()
	if err != nil {
		t.Fatal(err)
	}
	m, _ := data.(map[string]interface{})
	if m["running"] != true || m["uptime"] != "3d" {
		t.Fatalf("data = %v", m)
	}
}

func TestHelpDataUsageText(t *testing.T) {
	data, err := helpData()
	if err != nil {
		t.Fatal(err)
	}
	m, _ := data.(map[string]interface{})
	u, _ := m["usage"].(string)
	if u == "" || u[:len("UTMStack EDR — management CLI")] != "UTMStack EDR — management CLI" {
		t.Fatalf("usage = %q", u)
	}
}

func TestLifecycleDataShapes(t *testing.T) {
	install, _ := installData()
	installMap, _ := install.(map[string]interface{})
	if installMap["action"] != "install" {
		t.Fatalf("install data = %v", install)
	}
	uninstall, _ := uninstallData()
	uninstallMap, _ := uninstall.(map[string]interface{})
	if uninstallMap["action"] != "uninstall" {
		t.Fatalf("uninstall data = %v", uninstall)
	}
}
