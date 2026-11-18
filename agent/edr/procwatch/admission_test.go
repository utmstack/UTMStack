package procwatch

import (
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/guard"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"testing"
	"time"
)

type congestedScanner struct {
	entered chan struct{}
	release chan struct{}
}

func (s congestedScanner) ScanImage(string, string) (string, string, func() string, error) {
	s.entered <- struct{}{}
	<-s.release
	return "clean", "", nil, nil
}
func waitGuardSlots(t *testing.T, d *Dispatch, want int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for len(d.guardSlots) != want {
		if time.Now().After(deadline) {
			t.Fatalf("guard slots=%d want %d", len(d.guardSlots), want)
		}
		time.Sleep(time.Millisecond)
	}
}
func TestDispatchBoundsGuardsBeforeGoroutineCreationAndReportsLoss(t *testing.T) {
	sc := congestedScanner{make(chan struct{}, maxConcurrentGuards), make(chan struct{})}
	d := NewDispatch(proctable.New(), guard.New(config.Default(), sc, nil), nil, nil, nil)
	for pid := 1; pid <= maxConcurrentGuards; pid++ {
		d.OnProcStart(ProcStart{PID: pid, Image: "/opt/inert"})
	}
	for i := 0; i < maxConcurrentGuards; i++ {
		select {
		case <-sc.entered:
		case <-time.After(time.Second):
			t.Fatal("admitted scan did not start")
		}
	}
	d.OnProcStart(ProcStart{PID: 9999, Image: "/opt/inert"})
	if len(d.guardSlots) != maxConcurrentGuards || d.OverflowCount() != 1 || d.Health() == nil {
		t.Fatal("unbounded or silent guard congestion")
	}
	if _, ok := d.table.Get(9999); !ok {
		t.Fatal("overflow lost process telemetry/context")
	}
	close(sc.release)
	waitGuardSlots(t, d, 0)
	if d.Health() == nil {
		t.Fatal("lost launch disappeared from health")
	}
	d.OnProcStart(ProcStart{PID: 10000, Image: "/opt/inert"})
	waitGuardSlots(t, d, 0)
	if d.OverflowCount() != 1 {
		t.Fatal("released slot not reusable")
	}
}

type panicSuspender struct{}

func (panicSuspender) Suspend(int) error                   { panic("safe injected suspension failure") }
func (panicSuspender) Resume(int) error                    { return nil }
func (panicSuspender) KillTree(int, string) ([]int, error) { return nil, nil }
func TestDispatchReleasesGuardSlotAfterPanic(t *testing.T) {
	cfg := config.Default()
	cfg.SuspendOnLaunch = true
	d := NewDispatch(proctable.New(), guard.New(cfg, nil, panicSuspender{}), nil, nil, nil)
	d.OnProcStart(ProcStart{PID: 42, Image: "/opt/inert"})
	waitGuardSlots(t, d, 0)
	if d.guardPanics.Load() != 1 || d.Health() == nil {
		t.Fatal("guard panic did not release slot/report loss")
	}
}
