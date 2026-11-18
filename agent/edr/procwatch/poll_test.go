package procwatch

import (
	"github.com/utmstack/UTMStack/agent/edr/linuxproc"
	"testing"
)

func TestProcessChangesUsesCreationIdentityAndExecImage(t *testing.T) {
	first := []linuxproc.Process{{PID: 42, PPID: 7, StartTicks: 100, Image: "/opt/inert"}}
	state, events := processChanges(nil, first, true)
	if len(events) != 0 {
		t.Fatal("baseline reported as new launch")
	}
	state, events = processChanges(state, first, false)
	if len(events) != 0 {
		t.Fatal("duplicate launch")
	}
	reused := []linuxproc.Process{{PID: 42, PPID: 9, StartTicks: 101, Image: "/opt/inert"}}
	state, events = processChanges(state, reused, false)
	if len(events) != 1 || events[0].OSIdentity != 101 {
		t.Fatalf("PID reuse lost: %+v", events)
	}
	execed := []linuxproc.Process{{PID: 42, PPID: 9, StartTicks: 101, Image: "/opt/other"}}
	state, events = processChanges(state, execed, false)
	if len(events) != 1 || events[0].Image != "/opt/other" {
		t.Fatal(events)
	}
	state, _ = processChanges(state, nil, false)
	if len(state) != 0 {
		t.Fatal("exited process retained")
	}
}
