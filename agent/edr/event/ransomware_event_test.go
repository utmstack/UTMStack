package event

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewRansomwareEvent_BrandedAndTagged(t *testing.T) {
	p := ProcInfo{PID: 1234, PPID: 10, Image: `C:\Users\x\enc.exe`, Cmdline: "enc.exe --all"}
	ev := NewRansomwareEvent(ActionRansomwareContained, p, "canary:00__accounts.xlsx", "critical")

	if ev.Source != SourceBehavioral {
		t.Fatalf("source = %q, want %q", ev.Source, SourceBehavioral)
	}
	if ev.Action != ActionRansomwareContained {
		t.Fatalf("action = %q", ev.Action)
	}
	if ev.Signature != "canary:00__accounts.xlsx" || ev.Severity != "critical" {
		t.Fatalf("signal/severity not set: %+v", ev)
	}
	if ev.Process == nil || ev.Process.PID != 1234 {
		t.Fatalf("process not carried: %+v", ev.Process)
	}
	js, err := ev.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	_ = json.Unmarshal([]byte(js), &m)
	if m["product"] != Product || m["engine"] != Product {
		t.Fatalf("branding missing: %v", m)
	}
	if strings.Contains(strings.ToLower(js), "clam") {
		t.Fatalf("engine name leaked into event: %s", js)
	}
}
