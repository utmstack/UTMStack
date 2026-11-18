package ransomware

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

func fillScorer(t *testing.T, s *Scorer) {
	t.Helper()
	for pid := 1; pid <= maxScorerStates; pid++ {
		if d := s.Add(Evidence{PID: pid, Gen: 1, Kind: KindCanary, Weight: 1}); d.Err != nil {
			t.Fatalf("capacity admission %d: %v", pid, d.Err)
		}
	}
}

func TestScorerCapacityNeverEvictsActiveIdentityAndFailureStaysVisible(t *testing.T) {
	_, clock := fixedClock(time.Unix(1000, 0))
	s := NewScorer(50, 100, 10, clock)
	fillScorer(t, s)
	d := s.Add(Evidence{PID: maxScorerStates + 1, Gen: 1, Kind: KindCanary, Weight: 100})
	if !errors.Is(d.Err, ErrScorerCapacity) || d.Escalation != EscNone || len(s.states) != maxScorerStates || s.states[1].gen != 1 {
		t.Fatal("capacity silently evicted active state or hid rejected evidence", d)
	}
	s.Forget(1)
	if d := s.Add(Evidence{PID: maxScorerStates + 1, Gen: 1, Kind: KindCanary, Weight: 100}); d.Err != nil || d.Escalation != EscKill {
		t.Fatal("explicitly released capacity was not usable", d)
	}
	if !errors.Is(s.Health(), ErrScorerCapacity) {
		t.Fatal("lost evidence became healthy after capacity recovered")
	}
	if len(s.slots) > maxScorerStates || len(s.free) > maxScorerStates {
		t.Fatal("slot bookkeeping grew beyond bound")
	}
}

func TestScorerPrunesExpiredStateWithBoundedWorkAndRefusesStaleInput(t *testing.T) {
	current, clock := fixedClock(time.Unix(1000, 0))
	s := NewScorer(50, 100, 10, clock)
	fillScorer(t, s)
	*current = current.Add(scorerWindow + time.Second)
	before := len(s.states)
	if err := s.Health(); err != nil {
		t.Fatal(err)
	}
	if removed := before - len(s.states); removed != scorerPruneBudget {
		t.Fatalf("prune work is not bounded: %d", removed)
	}
	for i := 0; i < maxScorerStates/scorerPruneBudget; i++ {
		_ = s.Health()
	}
	if len(s.states) != 0 {
		t.Fatal("expired entries were starved by pruning cursor")
	}
	d := s.Add(Evidence{PID: 700, Gen: 1, Kind: KindCanary, Weight: 100, TS: current.Add(-scorerWindow - time.Second)})
	if !errors.Is(d.Err, ErrStaleEvidence) || len(s.states) != 0 {
		t.Fatal("stale input filled state", d)
	}
	if d := s.Add(Evidence{PID: 700, Gen: 1, Kind: KindCanary, Weight: 100}); d.Err != nil || d.Escalation != EscKill {
		t.Fatal("fresh evidence not admitted after expiry", d)
	}
}

func TestScorerOldGenerationCannotEraseNewOccupantScore(t *testing.T) {
	_, clock := fixedClock(time.Unix(1000, 0))
	s := NewScorer(50, 100, 10, clock)
	s.Add(Evidence{PID: 700, Gen: 10, Kind: KindCanary, Weight: 100})
	s.Add(Evidence{PID: 700, Gen: 11, Kind: KindCanary, Weight: 40})
	for _, gen := range []int64{10, 0} {
		d := s.Add(Evidence{PID: 700, Gen: gen, Kind: KindCanary, Weight: 100})
		if !errors.Is(d.Err, ErrStaleEvidence) || s.states[700].gen != 11 || s.states[700].score != 40 {
			t.Fatal("old/unknown identity replaced newer state", d)
		}
	}
	d := s.Add(Evidence{PID: 700, Gen: 11, Kind: KindCanary, Weight: 60})
	if d.Err != nil || d.Escalation != EscKill {
		t.Fatal("new occupant lost its accumulated score", d)
	}
}

func TestScorerBoundsEvidenceKindsDetailsAndWeights(t *testing.T) {
	_, clock := fixedClock(time.Unix(1000, 0))
	s := NewScorer(50, 100, 10, clock)
	d := s.Add(Evidence{PID: 700, Gen: 1, Kind: KindCanary, Weight: 100, Detail: strings.Repeat("x", 65536)})
	if d.Err != nil || len(s.states[700].topDetail) != maxEvidenceDetail {
		t.Fatal("evidence detail retained unbounded bytes", d)
	}
	for _, ev := range []Evidence{
		{PID: 701, Kind: SignalKind(strings.Repeat("unknown", 10000)), Weight: 100},
		{PID: 701, Kind: KindCanary, Weight: math.NaN()},
		{PID: 701, Kind: KindCanary, Weight: math.Inf(1)},
		{PID: 701, Kind: KindCanary, Weight: -1},
		{PID: 701, Kind: KindCanary, Weight: 1001},
	} {
		if d := s.Add(ev); !errors.Is(d.Err, ErrInvalidEvidence) || len(s.states) != 1 {
			t.Fatal("invalid evidence retained or silently treated as ordinary verdict", d)
		}
	}
	if !errors.Is(s.Health(), ErrInvalidEvidence) {
		t.Fatal("invalid admitted-source evidence did not degrade health")
	}
}

func TestScorerCurrentMaxWeightSignalSurvivesNormalDispatchLatency(t *testing.T) {
	base := time.Unix(1000, 0)
	s := NewScorer(50, 100, 10, func() time.Time { return base.Add(500 * time.Microsecond) })
	d := s.Add(Evidence{PID: 700, Gen: 1, Kind: KindCanary, Weight: 100, TS: base})
	if d.Err != nil || d.Escalation != EscKill || d.Score != 100 {
		t.Fatal("normal dispatch latency reduced a max-weight signal below containment threshold", d)
	}
}

type countRansomwareSpool struct{ n int }

func (c *countRansomwareSpool) Append(string) error { c.n++; return nil }

type countRansomwareIncidents struct{ n int }

func (c *countRansomwareIncidents) StoreIncident(cache.RansomwareIncident) error { c.n++; return nil }

type healthyScorerTestFeed struct{}

func (healthyScorerTestFeed) Run(ctx context.Context, _ func(FileEvent)) error {
	<-ctx.Done()
	return ctx.Err()
}
func (healthyScorerTestFeed) Health() error { return nil }

func TestGuardAlertModeStateIsBoundedAndCoverageHealthIsNotFeedHealth(t *testing.T) {
	g, resp, _, _, _, _ := newTestGuard("alert", map[string]bool{"/canary": true}, 50, 100)
	spool, inc := &countRansomwareSpool{}, &countRansomwareIncidents{}
	g.deps.Spool = spool
	g.deps.Incidents = inc
	g.liveFeed = healthyScorerTestFeed{}
	g.feedHealthy.Store(true)
	for pid := 1; pid <= maxScorerStates+10; pid++ {
		g.OnFileEvent(FileEvent{PID: pid, Path: "/canary", Op: OpWrite})
	}
	if len(g.scorer.states) != maxScorerStates || spool.n != maxScorerStates || inc.n != maxScorerStates || resp.killCount() != 0 {
		t.Fatal("alert-mode history was unbounded or performed a response")
	}
	if !g.FeedHealthy() || !errors.Is(g.Health(), ErrScorerCapacity) {
		t.Fatal("live feed concealed lost scorer evidence", g.Health())
	}
}
