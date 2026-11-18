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

	d.OnProcStart(ProcStart{PID: 77, PPID: 4, Image: `C:\enc.exe`, Cmdline: "enc.exe", OSIdentity: 12345, IdentityRequired: true})

	if got.PID != 77 {
		t.Fatalf("onProc not invoked with the ProcStart: %+v", got)
	}
	p, ok := tab.Get(77)
	if !ok || p.StartTS == 0 || p.OSIdentity != 12345 || !p.IdentityRequired {
		t.Fatalf("StartTS not stamped in table: %+v ok=%v", p, ok)
	}
}

func TestDispatchExitCannotDeleteReusedPID(t *testing.T) {
	tab := proctable.New()
	d := NewDispatch(tab, nil, nil, nil, nil)
	d.OnProcStart(ProcStart{PID: 42, OSIdentity: 100, IdentityRequired: true})
	d.OnProcStart(ProcStart{PID: 42, OSIdentity: 101, IdentityRequired: true})
	d.OnProcExit(42, 100)
	if p, ok := tab.Get(42); !ok || p.OSIdentity != 101 {
		t.Fatalf("reused PID deleted: %+v %v", p, ok)
	}
	d.OnProcExit(42, 101)
	if _, ok := tab.Get(42); ok {
		t.Fatal("exited instance retained")
	}
}

func TestCurrentSnapshotsAndForksCannotTriggerOriginalCommandResponse(t *testing.T) {
	tab := proctable.New()
	calls := 0
	snapshots := 0
	d := NewDispatch(tab, nil, nil, nil, func(ProcStart) { calls++ })
	d.SetSnapshotHook(func(ps ProcStart) {
		if ps.EventKind != "snapshot" || ps.OSIdentity != 100 {
			t.Fatal(ps)
		}
		snapshots++
	})
	for _, kind := range []string{"fork", "snapshot", "inventory"} {
		d.OnProcStart(ProcStart{PID: 42, OSIdentity: 100, IdentityRequired: true, EventKind: kind})
		if _, ok := tab.Get(42); !ok {
			t.Fatal("current identity was not retained")
		}
	}
	if calls != 0 {
		t.Fatal("unattributed original command response", calls)
	}
	if snapshots != 1 {
		t.Fatal("current observation was not delivered independently", snapshots)
	}
	d.OnProcStart(ProcStart{PID: 42, OSIdentity: 100, IdentityRequired: true, EventKind: "exec"})
	if calls != 1 {
		t.Fatal("bound exec was not evaluated")
	}
}
