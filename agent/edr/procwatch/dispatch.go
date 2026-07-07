package procwatch

import (
	"time"

	"github.com/utmstack/UTMStack/agent/edr/behavioral"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/guard"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"github.com/utmstack/UTMStack/shared/logger"
)

type Dispatch struct {
	table  *proctable.Table
	guard  *guard.Guard
	spool  *event.Spool            // optional: process-creation telemetry (§4.9)
	skip   func(image string) bool // optional: skip telemetry+guard for these images
	onProc func(ProcStart)         // optional: ransomware T1490 sensor hook
}

func NewDispatch(tab *proctable.Table, g *guard.Guard, sp *event.Spool, skip func(string) bool, onProc func(ProcStart)) *Dispatch {
	return &Dispatch{table: tab, guard: g, spool: sp, skip: skip, onProc: onProc}
}

func (d *Dispatch) OnProcStart(ps ProcStart) {
	p := proctable.Proc{PID: ps.PID, PPID: ps.PPID, Image: ps.Image, Cmdline: ps.Cmdline, StartTS: time.Now().UnixNano()}
	d.table.Add(p)

	// Ransomware T1490 sensor: match recovery-tampering commands on the raw
	// process start. Runs before the skip filter so vssadmin/wbadmin under
	// %SystemRoot% are still evaluated (the encryptor is their parent).
	if d.onProc != nil {
		d.onProc(ps)
	}

	// Skip the EDR's own engine helper processes (clamd/freshclam/clamdscan):
	// scanning them is wasteful, and their image/command line carry the engine
	// name — which must never reach an emitted event (branding rule).
	if d.skip != nil && d.skip(p.Image) {
		return
	}

	// Behavioral telemetry: forward the process-creation event (double duty with
	// the responder — one watcher, two consumers).
	if d.spool != nil {
		if js, err := behavioral.ProcTelemetry(p).ToJSON(); err == nil {
			_ = d.spool.Append(js)
		}
	}

	if d.guard != nil {
		// Run the guard off the event loop so scan latency never blocks WMI, with
		// panic recovery so one bad launch can't crash the service.
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("UTMStack EDR: panic in guard.OnStart: %v", r)
				}
			}()
			d.guard.OnStart(p)
		}()
	}
}
