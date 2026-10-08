//go:build windows || linux

package ransomware

import (
	"fmt"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestChurn_RenamesFire(t *testing.T) {
	base := time.Unix(1000, 0)
	cur := base
	c := NewChurnSensor(cfgWithBehavior(config.DefaultBehaviorTuning()), func() time.Time { return cur })

	var fired int
	// 12 renames across 6 distinct files (two per file — rename to a unique
	// temp then to the final name), all to .locked, inside the window.
	for i := 0; i < 6; i++ {
		for j := 0; j < 2; j++ {
			path := fmt.Sprintf("/data/f%d.locked", i)
			if ev := c.Record(FileEvent{PID: 3, Path: path, Op: OpRename}); ev != nil {
				fired++
				if ev.Kind != KindExtChurn {
					t.Fatalf("bad kind: %+v", ev)
				}
			}
			cur = cur.Add(200 * time.Millisecond)
		}
	}
	if fired != 1 {
		t.Fatalf("fired %d times, want 1", fired)
	}
}

func TestChurn_SameExtensionRewritesNeverFire(t *testing.T) {
	base := time.Unix(1000, 0)
	cur := base
	c := NewChurnSensor(cfgWithBehavior(config.DefaultBehaviorTuning()), func() time.Time { return cur })

	// 40 writes to the same .txt extension for one PID: first write is the
	// only acquisition, the rest are repeats → no fire.
	for i := 0; i < 40; i++ {
		path := fmt.Sprintf("/data/f%d.txt", i%10)
		if ev := c.Record(FileEvent{PID: 4, Path: path, Op: OpWrite}); ev != nil {
			t.Fatalf("same-extension rewrites fired: %+v", ev)
		}
		cur = cur.Add(50 * time.Millisecond)
	}
}

func TestChurn_LinuxWriteApproxFire(t *testing.T) {
	base := time.Unix(1000, 0)
	cur := base
	c := NewChurnSensor(cfgWithBehavior(config.DefaultBehaviorTuning()), func() time.Time { return cur })

	var fired int
	// 12 distinct files, each acquiring a new unusual extension (.enc), via
	// the write-approximation path used on Linux. One write per file, all
	// inside the 5s window → fires once.
	for i := 0; i < 12; i++ {
		p := fmt.Sprintf("/data/d%d.enc", i)
		if ev := c.Record(FileEvent{PID: 5, Path: p, Op: OpWrite}); ev != nil {
			fired++
			if ev.Kind != KindExtChurn {
				t.Fatalf("bad kind: %+v", ev)
			}
		}
		cur = cur.Add(300 * time.Millisecond)
	}
	if fired != 1 {
		t.Fatalf("write-approx fired %d times, want 1", fired)
	}
}

func TestChurn_CooldownSuppressesRefire(t *testing.T) {
	base := time.Unix(1000, 0)
	cur := base
	c := NewChurnSensor(cfgWithBehavior(config.DefaultBehaviorTuning()), func() time.Time { return cur })

	fire := func() *Evidence {
		for i := 0; i < 12; i++ {
			if ev := c.Record(FileEvent{PID: 6, Path: fmt.Sprintf("/data/g%d.locked", i), Op: OpRename}); ev != nil {
				return ev
			}
			cur = cur.Add(100 * time.Millisecond)
		}
		return nil
	}
	if fire() == nil {
		t.Fatal("first burst did not fire")
	}
	// Immediately after: counters were reset and cooldown is active → the
	// next 12 renames rebuild acquisitions but must not emit.
	if ev := fire(); ev != nil {
		t.Fatalf("refired inside cooldown: %+v", ev)
	}
}
