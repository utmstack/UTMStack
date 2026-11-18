package procwatch

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/behavioral"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/guard"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"github.com/utmstack/UTMStack/shared/logger"
)

const maxConcurrentGuards = 64

type Dispatch struct {
	guardSlots    chan struct{}
	guardOverflow atomic.Uint64
	guardPanics   atomic.Uint64
	table         *proctable.Table
	guard         *guard.Guard
	spool         *event.Spool            // optional: process-creation telemetry (§4.9)
	skip          func(image string) bool // optional: skip telemetry+guard for these images
	onProc        func(ProcStart)         // optional: ransomware T1490 sensor hook
	onSnapshot    func(ProcStart)         // independent current-state command observation
}

// SetSnapshotHook is configured before Run. Snapshot authority comes from the
// newly captured identity; it must never masquerade as the queued original exec.
func (d *Dispatch) SetSnapshotHook(fn func(ProcStart)) { d.onSnapshot = fn }

func NewDispatch(tab *proctable.Table, g *guard.Guard, sp *event.Spool, skip func(string) bool, onProc func(ProcStart)) *Dispatch {
	return &Dispatch{guardSlots: make(chan struct{}, maxConcurrentGuards), table: tab, guard: g, spool: sp, skip: skip, onProc: onProc}
}

func (d *Dispatch) OnProcStart(ps ProcStart) {
	p := proctable.Proc{PID: ps.PID, PPID: ps.PPID, Image: ps.Image, Cmdline: ps.Cmdline, StartTS: time.Now().UnixNano(), OSIdentity: ps.OSIdentity, IdentityRequired: ps.IdentityRequired}
	d.table.Add(p)

	// Ransomware T1490 sensor: match recovery-tampering commands on the raw
	// process start. Runs before the skip filter so vssadmin/wbadmin under
	// %SystemRoot% are still evaluated (the encryptor is their parent).
	if d.onProc != nil && ps.EventKind != "fork" && ps.EventKind != "snapshot" && ps.EventKind != "inventory" {
		d.onProc(ps)
	}
	if d.onSnapshot != nil && ps.EventKind == "snapshot" {
		d.onSnapshot(ps)
	}

	// Skip the EDR's own working tree and explicitly trusted images.
	if d.skip != nil && d.skip(p.Image) {
		return
	}

	// Behavioral telemetry: forward the process-creation event (double duty with
	// the responder — one watcher, two consumers).
	if d.spool != nil {
		ev := behavioral.ProcTelemetry(p)
		if ps.EventKind == "snapshot" {
			ev.Action = "process_snapshot"
		}
		if ps.EventKind == "fork" {
			ev.Action = "process_fork"
		}
		if ps.EventKind == "inventory" {
			ev.Action = "process_inventory"
		}
		if js, err := ev.ToJSON(); err == nil {
			_ = d.spool.Append(js)
		}
	}

	if d.guard != nil && ps.EventKind != "fork" && ps.EventKind != "inventory" {
		// Admission precedes goroutine creation: scanner admission alone cannot
		// bound the memory of goroutines waiting to reach it.
		select {
		case d.guardSlots <- struct{}{}:
		default:
			d.guardOverflow.Add(1)
			return
		}
		// Run the guard off the event loop so scan latency never blocks WMI, with
		// panic recovery so one bad launch can't crash the service.
		go func() {
			defer func() { <-d.guardSlots }()
			defer func() {
				if r := recover(); r != nil {
					d.guardPanics.Add(1)
					logger.Error("UTMStack EDR: panic in guard.OnStart: %v", r)
				}
			}()
			d.guard.OnStart(p)
		}()
	}
}

// OnProcExit removes only the observed instance, preserving a concurrently
// registered replacement that reused the numeric PID.
func (d *Dispatch) OnProcExit(pid int, identity int64) { d.table.RemoveObserved(pid, identity) }

// Overflow remains visible until a new dispatcher/service run; later idle
// capacity cannot recover a launch whose expedited guard was never admitted.
func (d *Dispatch) OverflowCount() uint64 { return d.guardOverflow.Load() }
func (d *Dispatch) Health() error {
	if overflow, panics := d.guardOverflow.Load(), d.guardPanics.Load(); overflow > 0 || panics > 0 {
		return fmt.Errorf("process guard coverage lost: %d admission overflows, %d panics", overflow, panics)
	}
	return d.table.Health()
}
