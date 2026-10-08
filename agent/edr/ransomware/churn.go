package ransomware

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// commonExts are extensions that normal software writes to; an acquisition of
// one of these is not a churn signal. Deliberately broad — FP cost is tuned in
// H9 via the config knobs, not by trimming this set. A rename counts regardless
// of target extension (a mass renamer is abnormal on its own), so this set
// only gates the Linux write-approximation path.
var commonExts = map[string]bool{
	"": true, ".tmp": true, ".temp": true, ".log": true, ".txt": true,
	".md": true, ".json": true, ".csv": true, ".xml": true, ".yaml": true,
	".yml": true, ".db": true, ".sqlite": true, ".sqlite3": true, ".bin": true,
	".dat": true, ".bak": true, ".old": true, ".part": true, ".swp": true,
	".swo": true, ".pyc": true, ".zip": true, ".gz": true, ".tar": true,
	".7z": true, ".rar": true, ".iso": true, ".img": true, ".cache": true,
	".ini": true, ".cfg": true, ".conf": true, ".pid": true, ".sock": true,
	".crdownload": true, ".download": true, ".partial": true,
}

// ChurnSensor detects a PID moving many distinct files onto new, unusual
// extensions within a short window — the on-disk signature of an encryptor
// appending .locked/.crypt/etc.
//
// Fire condition (both must hold, per the task spec):
//   - in-window acquisition events >= Behavior.ChurnMinRenames
//   - in-window distinct files     >= Behavior.ChurnMinDistinct
//
// Input by platform:
//   - Windows: OpRename is first-class; every rename is an acquisition event
//     (renames are abnormal regardless of target extension).
//   - Linux: fanotify delivers OpWrite only; a write to a file whose
//     extension is outside commonExts is an acquisition event, deduped per
//     (PID,path) so one big file rewritten many times counts once.
type ChurnSensor struct {
	cfg       config.EDRConfig
	now       func() time.Time
	mu        sync.Mutex
	acqs      map[int][]churnAcq   // pid → in-window acquisition events
	fireUntil map[int]time.Time    // pid → cooldown expiry
	written   map[int]map[string]bool // pid → path already counted (write dedupe)
}

type churnAcq struct {
	path string
	ts   time.Time
}

// NewChurnSensor builds the sensor. A nil now falls back to time.Now.
func NewChurnSensor(cfg config.EDRConfig, now func() time.Time) *ChurnSensor {
	if now == nil {
		now = time.Now
	}
	return &ChurnSensor{
		cfg: cfg, now: now,
		acqs:      map[int][]churnAcq{},
		fireUntil: map[int]time.Time{},
		written:   map[int]map[string]bool{},
	}
}

// Record ingests one file op and returns at most one evidence.
func (c *ChurnSensor) Record(ev FileEvent) *Evidence {
	if ev.Op != OpRename && ev.Op != OpWrite {
		return nil
	}
	// The write-approximation only counts unusual extensions; renames count
	// as-is. A common extension (.txt/.log/...) written en masse is normal
	// software, not an encryptor.
	if ev.Op == OpWrite {
		ext := strings.ToLower(filepath.Ext(ev.Path))
		if commonExts[ext] {
			return nil
		}
	}
	b := c.cfg.Ransomware.Behavior
	window := time.Duration(b.ChurnWindowMs) * time.Millisecond
	cooldown := time.Duration(b.CooldownMs) * time.Millisecond
	now := c.now()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Prune the window and de-duplicate write events per (PID,path).
	list := c.acqs[ev.PID]
	kept := list[:0]
	distinct := map[string]bool{}
	for _, a := range list {
		if now.Sub(a.ts) <= window {
			kept = append(kept, a)
			distinct[a.path] = true
		}
	}
	addEvent := true
	if ev.Op == OpWrite {
		// A first write to a new (pid,path) is an acquisition; repeats are not.
		w := c.written[ev.PID]
		if w == nil {
			w = map[string]bool{}
			c.written[ev.PID] = w
		}
		if w[ev.Path] {
			addEvent = false
		} else {
			w[ev.Path] = true
		}
	}
	if addEvent {
		kept = append(kept, churnAcq{path: ev.Path, ts: now})
		distinct[ev.Path] = true
	}
	c.acqs[ev.PID] = kept

	if len(kept) < b.ChurnMinRenames || len(distinct) < b.ChurnMinDistinct {
		return nil
	}
	// One fire per burst; suppressed while the cooldown is active.
	if until, ok := c.fireUntil[ev.PID]; ok && now.Before(until) {
		return nil
	}
	c.fireUntil[ev.PID] = now.Add(cooldown)
	// Reset counters so the next burst is measured fresh.
	c.acqs[ev.PID] = nil
	c.written[ev.PID] = map[string]bool{}

	return &Evidence{
		PID:    ev.PID,
		Kind:   KindExtChurn,
		Weight: float64(c.cfg.Ransomware.FuzzyWeight("ext_churn")),
		Detail: fmt.Sprintf("%d files in %dms", len(distinct), b.ChurnWindowMs),
		TS:     now,
	}
}
