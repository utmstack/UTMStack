package ransomware

import (
	"sync"
	"testing"
	"time"
)

func fixedClock(base time.Time) (*time.Time, func() time.Time) {
	t := base
	return &t, func() time.Time { return t }
}

func TestScorer_MaxWeightSignalKillsImmediately(t *testing.T) {
	now, clk := fixedClock(time.Unix(1000, 0))
	_ = now
	s := NewScorer(50, 100, 10, clk)
	d := s.Add(Evidence{PID: 7, Gen: 1, Kind: KindCanary, Weight: 100, Detail: "00__a.xlsx"})
	if d.Escalation != EscKill {
		t.Fatalf("canary touch must cross kill: got %v score=%v", d.Escalation, d.Score)
	}
	if d.TopSignal != "canary:00__a.xlsx" {
		t.Fatalf("top signal = %q", d.TopSignal)
	}
}

func TestScorer_EscalationDoesNotRefire(t *testing.T) {
	_, clk := fixedClock(time.Unix(1000, 0))
	s := NewScorer(50, 100, 10, clk)
	if s.Add(Evidence{PID: 7, Gen: 1, Kind: KindT1490, Weight: 100}).Escalation != EscKill {
		t.Fatal("first should kill")
	}
	// A second signal on the same PID+gen must not re-emit EscKill.
	if e := s.Add(Evidence{PID: 7, Gen: 1, Kind: KindCanary, Weight: 100}).Escalation; e != EscNone {
		t.Fatalf("re-fire = %v, want EscNone", e)
	}
}

func TestScorer_DecayLowersScore(t *testing.T) {
	base := time.Unix(1000, 0)
	cur := base
	s := NewScorer(50, 100, 10, func() time.Time { return cur })
	// One 60-weight signal → suspend but not kill.
	if s.Add(Evidence{PID: 5, Gen: 1, Kind: KindEntropyPlaceholder, Weight: 60}).Escalation != EscSuspend {
		t.Fatal("60 should suspend")
	}
	// Advance two half-lives (20s): 60 → ~15, well under suspend threshold.
	cur = base.Add(20 * time.Second)
	d := s.Add(Evidence{PID: 5, Gen: 1, Kind: KindEntropyPlaceholder, Weight: 1})
	if d.Score > 30 {
		t.Fatalf("score did not decay: %v", d.Score)
	}
}

func TestScorer_RecycledPIDResets(t *testing.T) {
	_, clk := fixedClock(time.Unix(1000, 0))
	s := NewScorer(50, 100, 10, clk)
	s.Add(Evidence{PID: 9, Gen: 1, Kind: KindCanary, Weight: 100})
	// Same PID, new generation → fresh score, a single small signal must not kill.
	d := s.Add(Evidence{PID: 9, Gen: 2, Kind: KindEntropyPlaceholder, Weight: 10})
	if d.Escalation != EscNone {
		t.Fatalf("recycled PID inherited old score: %v", d.Escalation)
	}
}

func TestScorer_KindDiversityBonus(t *testing.T) {
	_, clk := fixedClock(time.Unix(1000, 0))
	s := NewScorer(50, 100, 10, clk)
	// Two DIFFERENT 40-weight kinds: raw 80, with +25% diversity → 100 → kill.
	s.Add(Evidence{PID: 3, Gen: 1, Kind: KindEntropyPlaceholder, Weight: 40})
	d := s.Add(Evidence{PID: 3, Gen: 1, Kind: KindExtChurnPlaceholder, Weight: 40})
	if d.Escalation != EscKill {
		t.Fatalf("diversity bonus missing: score=%v esc=%v", d.Score, d.Escalation)
	}
}

// TestScorer_ConcurrentAddNoRace mimics production: the ETW file-activity feed
// and the WMI process dispatch both feed evidence into one Scorer from separate
// goroutines. Without internal locking, concurrent map writes panic
// ("concurrent map writes"). Run under -race to prove the mutex serializes them.
func TestScorer_ConcurrentAddNoRace(t *testing.T) {
	_, clk := fixedClock(time.Unix(1000, 0))
	s := NewScorer(50, 100, 10, clk)

	kinds := []SignalKind{KindCanary, KindT1490, KindEntropyPlaceholder, KindExtChurnPlaceholder}
	weights := []float64{10, 25, 40, 100}

	const adders = 50
	var wg sync.WaitGroup
	wg.Add(adders)
	for i := 0; i < adders; i++ {
		go func(i int) {
			defer wg.Done()
			// Vary deterministically by goroutine index: spread across a small
			// set of PIDs so writers collide on shared map entries, and rotate
			// kinds/weights so every code path (new state, decay, diversity) runs.
			pid := i%8 + 1
			for j := 0; j < 200; j++ {
				s.Add(Evidence{
					PID:    pid,
					Gen:    1,
					Kind:   kinds[(i+j)%len(kinds)],
					Weight: weights[(i+j)%len(weights)],
				})
			}
		}(i)
	}

	// A few concurrent Forgets on the same PID space exercise map deletion
	// racing against the inserts above.
	const forgetters = 6
	wg.Add(forgetters)
	for i := 0; i < forgetters; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				s.Forget(i%8 + 1)
			}
		}(i)
	}

	wg.Wait()
	// Reaching here without a "concurrent map writes" panic or a race-detector
	// report is the assertion.
}
