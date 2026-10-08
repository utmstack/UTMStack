package ransomware

import (
	"fmt"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// MassRateSensor tracks a sliding window of write ops per PID and emits one
// KindMassRate evidence when the in-window write count crosses the configured
// minimum. A per-PID cooldown bounds re-fire: one signal per burst, so a
// long-running encryptor does not flood the scorer with identical evidence.
// Gen is 0 on the emitted evidence: the scorer adopts the PID's real
// generation from the process table on a later signal (see
// TestAdoptGenWhenUnknown) and treats a 0→known transition as adoption, not
// a reset.
type MassRateSensor struct {
	cfg  config.EDRConfig
	now  func() time.Time
	mu   sync.Mutex
	// writes maps PID → in-window write timestamps (pruned on Record).
	writes map[int][]time.Time
	// lastFire maps PID → the timestamp of the last emitted evidence.
	lastFire map[int]time.Time
}

// NewMassRateSensor builds the sensor from the guard's config. The config is
// read at construction (structural), matching the Scorer's own handling of
// thresholds; a nil now falls back to time.Now.
func NewMassRateSensor(cfg config.EDRConfig, now func() time.Time) *MassRateSensor {
	if now == nil {
		now = time.Now
	}
	return &MassRateSensor{cfg: cfg, now: now, writes: map[int][]time.Time{}, lastFire: map[int]time.Time{}}
}

// Record ingests one file op. Only OpWrite counts (on Linux the whole fuzzy
// stack derives from the write stream; on Windows writes are a faithful
// subset of all mutating activity). It returns ready evidence exactly once
// per burst, or nil.
func (m *MassRateSensor) Record(ev FileEvent) *Evidence {
	if ev.Op != OpWrite {
		return nil
	}
	b := m.cfg.Ransomware.Behavior
	window := time.Duration(b.MassRateWindowMs) * time.Millisecond
	cooldown := time.Duration(b.CooldownMs) * time.Millisecond
	now := m.now()

	m.mu.Lock()
	defer m.mu.Unlock()

	ts := m.writes[ev.PID]
	pruned := ts[:0]
	for _, t := range ts {
		if now.Sub(t) <= window {
			pruned = append(pruned, t)
		}
	}
	pruned = append(pruned, now)
	m.writes[ev.PID] = pruned

	if len(pruned) < b.MassRateMinWrites {
		return nil
	}
	if last, ok := m.lastFire[ev.PID]; ok && now.Sub(last) < cooldown {
		return nil
	}
	m.lastFire[ev.PID] = now
	return &Evidence{
		PID:    ev.PID,
		Kind:   KindMassRate,
		Weight: float64(m.cfg.Ransomware.FuzzyWeight("mass_rate")),
		Detail: fmt.Sprintf("%d writes in %dms", len(pruned), b.MassRateWindowMs),
		TS:     now,
	}
}
