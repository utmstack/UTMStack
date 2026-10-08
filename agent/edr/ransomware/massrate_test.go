//go:build windows || linux

package ransomware

import (
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func cfgWithBehavior(b config.BehaviorTuning) config.EDRConfig {
	c := config.Default()
	c.Ransomware.Behavior = b
	return c
}

func TestMassRate_FiresOncePerBurst(t *testing.T) {
	base := time.Unix(1000, 0)
	cur := base
	m := NewMassRateSensor(cfgWithBehavior(config.DefaultBehaviorTuning()), func() time.Time { return cur })

	var fired int
	for i := 0; i < 30; i++ {
		// 30 rapid writes, 100ms apart — inside the 5s window.
		if ev := m.Record(FileEvent{PID: 7, Path: "/data/f", Op: OpWrite}); ev != nil {
			fired++
			if ev.Kind != KindMassRate || ev.PID != 7 {
				t.Fatalf("bad evidence: %+v", ev)
			}
			if ev.Weight != float64(c2weight()) {
				t.Fatalf("weight = %v", ev.Weight)
			}
		}
		cur = cur.Add(100 * time.Millisecond)
	}
	if fired != 1 {
		t.Fatalf("fired %d times, want 1 (cooldown)", fired)
	}
}

// c2weight is a small indirection so the test asserts against the actual
// default config weight rather than a hardcoded copy of it.
func c2weight() int {
	return config.Default().Ransomware.FuzzyWeight("mass_rate")
}

func TestMassRate_SlowWritesNeverFire(t *testing.T) {
	base := time.Unix(1000, 0)
	cur := base
	m := NewMassRateSensor(cfgWithBehavior(config.DefaultBehaviorTuning()), func() time.Time { return cur })
	for i := 0; i < 30; i++ {
		// One write every 2s: the 5s window holds at most 3 → never 25.
		if ev := m.Record(FileEvent{PID: 8, Path: "/data/f", Op: OpWrite}); ev != nil {
			t.Fatalf("slow writes fired: %+v", ev)
		}
		cur = cur.Add(2 * time.Second)
	}
}

func TestMassRate_RestartsAfterCooldown(t *testing.T) {
	base := time.Unix(1000, 0)
	cur := base
	m := NewMassRateSensor(cfgWithBehavior(config.DefaultBehaviorTuning()), func() time.Time { return cur })

	for i := 0; i < 25; i++ {
		m.Record(FileEvent{PID: 9, Path: "/data/f", Op: OpWrite})
		cur = cur.Add(100 * time.Millisecond)
	}
	// Still inside the 10s cooldown → silence.
	if ev := m.Record(FileEvent{PID: 9, Path: "/data/f", Op: OpWrite}); ev != nil {
		t.Fatalf("fired inside cooldown: %+v", ev)
	}
	// Jump past the cooldown; the window is now empty (writes aged out) so
	// rebuild the burst.
	cur = cur.Add(15 * time.Second)
	var fired int
	for i := 0; i < 25; i++ {
		if ev := m.Record(FileEvent{PID: 9, Path: "/data/f", Op: OpWrite}); ev != nil {
			fired++
		}
		cur = cur.Add(100 * time.Millisecond)
	}
	if fired != 1 {
		t.Fatalf("second burst fired %d times, want 1", fired)
	}
}

func TestMassRate_IgnoresNonWrites(t *testing.T) {
	base := time.Unix(1000, 0)
	cur := base
	m := NewMassRateSensor(cfgWithBehavior(config.DefaultBehaviorTuning()), func() time.Time { return cur })
	for i := 0; i < 40; i++ {
		if ev := m.Record(FileEvent{PID: 10, Path: "/data/f", Op: OpRename}); ev != nil {
			t.Fatalf("rename counted as write: %+v", ev)
		}
		cur = cur.Add(50 * time.Millisecond)
	}
}
