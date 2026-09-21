package event

import (
	"strings"
	"testing"
)

func TestNewProcessActionBrandedWithProc(t *testing.T) {
	e := NewProcessAction(ActionKilled, SourceProcessWatcher, ProcInfo{PID: 4321, PPID: 100, Image: `C:\a.exe`}, "malicious", "Win.Test.EICAR")
	js, err := e.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(js, `"action":"killed"`) || !strings.Contains(js, `"pid":4321`) {
		t.Fatalf("bad event: %s", js)
	}
	if !strings.Contains(js, `"product":"UTMStack EDR"`) {
		t.Fatalf("missing branding: %s", js)
	}
	if strings.Contains(strings.ToLower(js), "clam") {
		t.Fatalf("leaked engine name: %s", js)
	}
}
