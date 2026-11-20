package ransomware

import (
	"errors"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	maxScorerStates   = 32768
	scorerPruneBudget = 64
	scorerWindow      = 10 * time.Minute
	maxEvidenceDetail = 256
)

var (
	ErrScorerCapacity  = errors.New("ransomware scorer capacity exceeded; evidence coverage lost")
	ErrStaleEvidence   = errors.New("ransomware evidence outside current identity/time window")
	ErrInvalidEvidence = errors.New("ransomware evidence invalid or outside bounded limits")
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
	Err        error // rejected/overflow evidence is not an ordinary no-escalation verdict
}

type pidState struct {
	slot       int
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
	slots           []int
	free            []int
	pruneCursor     int
	failure         error // sticky: later idle capacity does not recover lost evidence
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
	s.remove(pid)
}

// ForgetGeneration releases only evidence for the completed process instance.
// An unknown generation (zero) can clear only another unknown generation; it
// never authorizes removing evidence with an observed creation identity.
func (s *Scorer) ForgetGeneration(pid int, gen int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st := s.states[pid]; st != nil && st.gen == gen {
		s.remove(pid)
	}
}

func (s *Scorer) remove(pid int) {
	if st, ok := s.states[pid]; ok {
		delete(s.states, pid)
		s.slots[st.slot] = 0
		s.free = append(s.free, st.slot)
	}
}

func (s *Scorer) prune(t time.Time) {
	for checked := 0; checked < scorerPruneBudget && checked < len(s.slots); checked++ {
		if s.pruneCursor >= len(s.slots) {
			s.pruneCursor = 0
		}
		pid := s.slots[s.pruneCursor]
		s.pruneCursor++
		if st := s.states[pid]; st != nil && t.Sub(st.lastUpdate) > scorerWindow {
			s.remove(pid)
		}
	}
}

func (s *Scorer) Health() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prune(s.now())
	return s.failure
}

func boundedKind(k SignalKind) (SignalKind, bool) {
	switch k {
	case KindCanary:
		return KindCanary, true
	case KindT1490:
		return KindT1490, true
	case KindEntropyPlaceholder:
		return KindEntropyPlaceholder, true
	case KindExtChurnPlaceholder:
		return KindExtChurnPlaceholder, true
	case KindMassRate:
		return KindMassRate, true
	case KindRansomNote:
		return KindRansomNote, true
	case KindRegistryTamper:
		return KindRegistryTamper, true
	default:
		return "", false
	}
}

func (s *Scorer) Add(ev Evidence) Decision {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := s.now()
	s.prune(t)
	kind, known := boundedKind(ev.Kind)
	if ev.PID <= 0 || ev.Gen < 0 || !known || math.IsNaN(ev.Weight) || math.IsInf(ev.Weight, 0) || ev.Weight <= 0 || ev.Weight > 1000 {
		if s.failure == nil {
			s.failure = ErrInvalidEvidence
		}
		return Decision{Err: ErrInvalidEvidence}
	}
	ev.Kind = kind
	if !ev.TS.IsZero() {
		age := t.Sub(ev.TS)
		if age > scorerWindow || age < -time.Second {
			return Decision{Err: ErrStaleEvidence}
		}
		// Ordinary sensor dispatch latency must not turn the documented single
		// max-weight canary/T1490 signal into 99.999... below its kill threshold.
		if age > time.Second && s.halfLife > 0 {
			ev.Weight *= math.Pow(2, -age.Seconds()/s.halfLife)
		}
	}
	st := s.states[ev.PID]
	if st != nil && t.Sub(st.lastUpdate) > scorerWindow {
		s.remove(ev.PID)
		st = nil
	}
	if st != nil && ev.Gen != st.gen {
		if ev.Gen == 0 || ev.Gen < st.gen {
			return Decision{Err: ErrStaleEvidence}
		}
		// Explicitly newer observed process, not an arbitrary eviction to admit
		// another PID. Never let trailing old-generation events erase its score.
		s.remove(ev.PID)
		st = nil
	}
	if st == nil {
		if len(s.states) >= maxScorerStates {
			if s.failure == nil {
				s.failure = ErrScorerCapacity
			}
			return Decision{Err: ErrScorerCapacity}
		}
		var slot int
		if n := len(s.free); n > 0 {
			slot = s.free[n-1]
			s.free = s.free[:n-1]
		} else {
			slot = len(s.slots)
			s.slots = append(s.slots, 0)
		}
		s.slots[slot] = ev.PID
		st = &pidState{slot: slot, gen: ev.Gen, lastUpdate: t, kinds: map[SignalKind]float64{}}
		s.states[ev.PID] = st
	}
	// Decay the existing score to `t`.
	if s.halfLife > 0 {
		dt := t.Sub(st.lastUpdate).Seconds()
		if dt > 0 {
			st.score *= math.Pow(2, -dt/s.halfLife)
		}
	}
	if t.After(st.lastUpdate) {
		st.lastUpdate = t
	}
	st.score = math.Min(1e9, st.score+ev.Weight)
	if ev.Weight > st.kinds[ev.Kind] {
		st.kinds[ev.Kind] = ev.Weight
	}
	if ev.Weight >= st.topWeight {
		detail := ev.Detail
		if len(detail) > maxEvidenceDetail {
			detail = detail[:maxEvidenceDetail]
		}
		st.topWeight, st.topKind, st.topDetail = ev.Weight, ev.Kind, strings.Clone(detail)
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
