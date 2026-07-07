//go:build windows

package behavioral

import (
	"context"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/shared/exec"
	"github.com/utmstack/UTMStack/shared/logger"
)

type PSLogReader struct {
	spool     *event.Spool
	lastRecID int64
}

func NewPSLogReader(sp *event.Spool) *PSLogReader { return &PSLogReader{spool: sp} }

func (r *PSLogReader) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.poll()
		}
	}
}

func (r *PSLogReader) poll() {
	// Query the last 20 PowerShell script-block (4104) events as XML.
	out, err := exec.RunWithOutput("wevtutil", ".",
		"qe", "Microsoft-Windows-PowerShell/Operational",
		"/q:*[System[EventID=4104]]", "/c:20", "/rd:true", "/f:xml")
	if err != nil {
		logger.Debug(100, "UTMStack EDR behavioral: wevtutil: %v", err)
		return
	}
	// wevtutil emits events newest-first; forward only records past our cursor.
	max := r.lastRecID
	for _, ev := range parse4104(out) {
		if ev.recordID <= r.lastRecID {
			continue
		}
		if ev.recordID > max {
			max = ev.recordID
		}
		te := ScriptBlockTelemetry(ev.scriptText, "PowerShell")
		if js, err := te.ToJSON(); err == nil {
			_ = r.spool.Append(js)
		}
	}
	r.lastRecID = max
}
