package event

import (
	"encoding/json"
	"runtime"
	"testing"
)

func TestEventReportsActualOperatingSystem(t *testing.T) {
	raw, err := (Event{Source: SourceScheduledScan}).ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	var out Event
	if err = json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.OS != runtime.GOOS {
		t.Fatalf("reported OS %q want %s", out.OS, runtime.GOOS)
	}
}
