package event

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSanitizeNameReplacesVendorComponent(t *testing.T) {
	cases := map[string]string{
		"Clamav.Test.File-6":              "UTMStack.Test.File-6",
		"CLAMAV.Test.File-7":              "UTMStack.Test.File-7",
		"clamav.Test.File-6.UNOFFICIAL":   "UTMStack.Test.File-6.UNOFFICIAL",
		"Win.Test.ClamAV":                 "Win.Test.UTMStack",
		"Doc.ClamAV.Macro-1":              "Doc.UTMStack.Macro-1",
		"Clamav-Test-File-6":              "UTMStack-Test-File-6",
		"file_watcher:Clamav.Test.File-6": "file_watcher:UTMStack.Test.File-6",
		"Clamav.Test.File-6:Clamav":       "UTMStack.Test.File-6:UTMStack",
	}
	for in, want := range cases {
		if got := SanitizeName(in); got != want {
			t.Errorf("SanitizeName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSanitizeNameLeavesFamilyNamesAlone(t *testing.T) {
	for _, in := range []string{
		"",
		"clam",
		"Win.Test.EICAR_HDB-1",
		"Win.Trojan.Clampi-43",
		"Win.Trojan.Clamavir-1", // the letters continue: a different word
		"Win.Trojan.MyClamav-1", // not at the start of a component
		"Win.Trojan.Clamav2-1",
		"Win.Trojan.Clamavé-1",
		"Analysis.Incomplete.ResourceLimit",
		"canary:00__accounts.xlsx",
	} {
		if got := SanitizeName(in); got != in {
			t.Errorf("SanitizeName(%q) = %q, want it unchanged", in, got)
		}
	}
}

// eventFields decodes an event the way the platform receives it.
func eventFields(t *testing.T, e Event) map[string]any {
	t.Helper()
	js, err := e.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(js), &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestEveryEventPathSanitizesDetectionNames(t *testing.T) {
	const vendor = "Clamav.Test.File-6"
	const want = "UTMStack.Test.File-6"
	p := ProcInfo{PID: 7, PPID: 1, Image: `C:\x\a.exe`}
	cases := map[string]Event{
		"file":                NewDetection(`C:\x\a.exe`, "ab", vendor, SourceFileWatcher),
		"engine":              NewDetection(`C:\x\a.exe`, "ab", vendor, SourceEngine),
		"process":             NewProcessAction(ActionKilled, SourceProcessWatcher, p, "malicious", vendor),
		"process with source": NewProcessAction(ActionKilled, SourceProcessWatcher, p, "malicious", "file_watcher:"+vendor),
		"amsi":                NewAMSIEvent(ActionBlocked, "malicious", vendor, "PowerShell"),
		"ransomware":          NewRansomwareEvent(ActionRansomwareContained, p, vendor, "critical"),
		// Events a package builds as a literal (the incomplete-analysis and
		// feed-health reports) still pass through ToJSON.
		"literal": {Source: SourceEngine, Action: ActionHealth, Signature: vendor},
	}
	for name, e := range cases {
		m := eventFields(t, e)
		sig, _ := m["signature"].(string)
		if !strings.HasSuffix(sig, want) {
			t.Errorf("%s: signature = %q, want it to end in %q", name, sig, want)
		}
		js, _ := e.ToJSON()
		if strings.Contains(strings.ToLower(js), "clamav") {
			t.Errorf("%s: event carries the vendor name: %s", name, js)
		}
	}

	// A real family name passes through every path unchanged.
	fam := "Win.Trojan.Clampi-43"
	for name, e := range map[string]Event{
		"file":    NewDetection(`C:\x\a.exe`, "ab", fam, SourceFileWatcher),
		"process": NewProcessAction(ActionKilled, SourceProcessWatcher, p, "malicious", fam),
		"amsi":    NewAMSIEvent(ActionBlocked, "malicious", fam, "PowerShell"),
		"literal": {Source: SourceEngine, Action: ActionHealth, Signature: fam},
	} {
		if sig := eventFields(t, e)["signature"]; sig != fam {
			t.Errorf("%s: signature = %v, want %q unchanged", name, sig, fam)
		}
	}

	// Incomplete-analysis names are the product's own and stay as written.
	inc := NewAMSIEvent(ActionHealth, "unknown", "Analysis.Incomplete.Timeout", "PowerShell")
	if sig := eventFields(t, inc)["signature"]; sig != "Analysis.Incomplete.Timeout" {
		t.Errorf("incomplete-analysis signature = %v", sig)
	}
}

func TestScriptBlockTextIsNotRewritten(t *testing.T) {
	// A behavioral script-block event carries the captured script in Signature.
	// That is evidence, not a detection name, and must reach the platform as
	// the host ran it.
	script := "clamav.exe --version"
	e := Event{Source: SourceBehavioral, Action: "script_block", ObjectPath: "PowerShell", Signature: script}
	if sig := eventFields(t, e)["signature"]; sig != script {
		t.Fatalf("script text rewritten to %v", sig)
	}
}

func TestNetworkEventHasNoSignatureToRewrite(t *testing.T) {
	e := NewNetworkEvent(ActionBlocked, "outbound", "1.2.3.4", 443, "bad.example", "1.2.3.4", nil)
	m := eventFields(t, e)
	if _, ok := m["signature"]; ok {
		t.Fatalf("network event gained a signature: %v", m)
	}
	if m["domain"] != "bad.example" || m["indicator"] != "1.2.3.4" {
		t.Fatalf("network fields changed: %v", m)
	}
}

func TestRansomwareContainFailedAction(t *testing.T) {
	if ActionRansomwareContainFailed != "ransomware_contain_failed" {
		t.Fatalf("ActionRansomwareContainFailed = %q", ActionRansomwareContainFailed)
	}
	e := NewRansomwareEvent(ActionRansomwareContainFailed, ProcInfo{PID: 9}, "t1490:vssadmin_delete_shadows", "critical")
	if a := eventFields(t, e)["action"]; a != "ransomware_contain_failed" {
		t.Fatalf("action = %v", a)
	}
}
