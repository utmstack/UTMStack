package procwatch

import (
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

func TestDispatchAddsToTable(t *testing.T) {
	tab := proctable.New()
	d := NewDispatch(tab, nil, nil, nil, nil) // nil guard/spool/onProc => dispatch only updates the table
	d.OnProcStart(ProcStart{PID: 909, PPID: 4, Image: `C:\a.exe`})

	// table update is synchronous
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if _, ok := tab.Get(909); ok {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("process 909 was not added to the table")
}

func TestDispatch_InvokesOnProcAndStampsStartTS(t *testing.T) {
	tab := proctable.New()
	var got ProcStart
	d := NewDispatch(tab, nil, nil, nil, func(ps ProcStart) { got = ps })

	d.OnProcStart(ProcStart{PID: 77, PPID: 4, Image: `C:\enc.exe`, Cmdline: "enc.exe"})

	if got.PID != 77 {
		t.Fatalf("onProc not invoked with the ProcStart: %+v", got)
	}
	p, ok := tab.Get(77)
	if !ok || p.StartTS == 0 {
		t.Fatalf("StartTS not stamped in table: %+v ok=%v", p, ok)
	}
}
