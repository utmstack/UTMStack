package feed

import (
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/event"
)

// SignatureSource names the effective signature-update source for status and
// telemetry. Branded: never names the engine or the public database vendor.
func SignatureSource(cfg config.EDRConfig) string {
	if engine.ResolveMirrorURL(cfg) != "" {
		return "utmstack-mirror"
	}
	return "official-cdn"
}

func (f *Feed) SetSpool(sp *event.Spool) { f.spool = sp }

func (f *Feed) setLastSuccess(t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastSuccess = t
}

// LastSuccess returns the time of the last successful signature update (zero
// if none this process lifetime). Safe for cross-goroutine status reads.
func (f *Feed) LastSuccess() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.lastSuccess
}

// Stale reports whether the last success is older than max. A feed that has
// never succeeded has no baseline and is not reported stale.
func (f *Feed) Stale(now time.Time, max time.Duration) bool {
	ls := f.LastSuccess()
	if ls.IsZero() {
		return false
	}
	return now.Sub(ls) > max
}

// noteHealth emits one branded health event when staleness is first crossed;
// re-armed by the next success.
func (f *Feed) noteHealth(interval time.Duration) {
	if f.spool == nil {
		return
	}
	if !f.Stale(time.Now(), 3*interval) {
		return
	}
	if f.staleNotified {
		return
	}
	f.staleNotified = true
	ev := event.Event{
		Source:    event.SourceEngine,
		Action:    event.ActionHealth,
		Signature: "signature_updates_stale",
		Severity:  "warning",
	}
	if js, err := ev.ToJSON(); err == nil {
		_ = f.spool.Append(js)
	}
}
