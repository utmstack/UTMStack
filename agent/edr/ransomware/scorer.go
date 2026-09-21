package ransomware

import (
	"math"
	"sort"
	"sync"
	"time"
)

type Escalation int

const (
	EscNone Escalation = iota
	EscSuspend
	EscKill
)

// Decision is the scorer's verdict for one Add call.
type Decision struct {
	Escalation Escalation
	Score      float64
	TopSignal  string // "<kind>:<detail>" of the heaviest contributing evidence
	Kinds      []SignalKind
}

type pidState struct {
	gen        int64
	score      float64
	lastUpdate time.Time
	kinds      map[SignalKind]float64 // kind → heaviest weight seen (for diversity + top signal)
	topKind    SignalKind
	topDetail  string
	topWeight  float64
	maxEsc     Escalation // highest escalation already emitted (prevents re-firing)
}

// Scorer accumulates weighted, PID-tagged evidence with exponential time-decay
// and a kind-diversity bonus, mapping the running score to an escalation.
type Scorer struct {
	mu              sync.Mutex
	suspendT, killT float64
	halfLife        float64 // seconds
	now             func() time.Time
	states          map[int]*pidState
}

func NewScorer(suspendT, killT, halfLife float64, now func() time.Time) *Scorer {
	if now == nil {
		now = time.Now
	}
	return &Scorer{suspendT: suspendT, killT: killT, halfLife: halfLife, now: now, states: map[int]*pidState{}}
}

func (s *Scorer) Forget(pid int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.states, pid)
}

func (s *Scorer) Add(ev Evidence) Decision {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := ev.TS
	if t.IsZero() {
		t = s.now()
	}
	st := s.states[ev.PID]
	if st == nil || (ev.Gen != 0 && st.gen != ev.Gen) {
		st = &pidState{gen: ev.Gen, lastUpdate: t, kinds: map[SignalKind]float64{}}
		s.states[ev.PID] = st
	}
	// Decay the existing score to `t`.
	if s.halfLife > 0 {
		dt := t.Sub(st.lastUpdate).Seconds()
		if dt > 0 {
			st.score *= math.Pow(2, -dt/s.halfLife)
		}
	}
	st.lastUpdate = t
	st.score += ev.Weight
	if ev.Weight > st.kinds[ev.Kind] {
		st.kinds[ev.Kind] = ev.Weight
	}
	if ev.Weight >= st.topWeight {
		st.topWeight, st.topKind, st.topDetail = ev.Weight, ev.Kind, ev.Detail
	}

	// Diversity bonus: +25% per distinct kind beyond the first.
	mult := 1 + 0.25*float64(len(st.kinds)-1)
	eff := st.score * mult

	esc := EscNone
	switch {
	case eff >= s.killT:
		esc = EscKill
	case eff >= s.suspendT:
		esc = EscSuspend
	}
	// Only emit an escalation that is strictly higher than one already emitted.
	out := EscNone
	if esc > st.maxEsc {
		st.maxEsc = esc
		out = esc
	}
	return Decision{Escalation: out, Score: eff, TopSignal: signalString(st.topKind, st.topDetail), Kinds: sortedKinds(st.kinds)}
}

func signalString(k SignalKind, detail string) string {
	if detail == "" {
		return string(k)
	}
	return string(k) + ":" + detail
}

func sortedKinds(m map[SignalKind]float64) []SignalKind {
	ks := make([]SignalKind, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Slice(ks, func(i, j int) bool { return ks[i] < ks[j] })
	return ks
}
