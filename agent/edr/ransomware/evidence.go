package ransomware

import "time"

// SignalKind identifies a ransomware detector's output. v1 ships the two
// high-fidelity kinds; the *Placeholder kinds reserve weights for v2 fuzzy
// sensors and keep the scorer exercised with sub-max evidence today.
type SignalKind string

const (
	KindCanary SignalKind = "canary"
	KindT1490  SignalKind = "t1490"

	KindEntropyPlaceholder  SignalKind = "entropy"   // v2
	KindExtChurnPlaceholder SignalKind = "ext_churn" // v2
)

// DefaultWeights maps a kind to its evidence weight. Canary and T1490 are
// max-weight (a single hit crosses the default kill threshold of 100).
var DefaultWeights = map[SignalKind]float64{
	KindCanary:              100,
	KindT1490:               100,
	KindEntropyPlaceholder:  40,
	KindExtChurnPlaceholder: 40,
}

// Evidence is one detector observation about a process. Gen is the process
// generation (proctable StartTS) used to invalidate a recycled PID.
type Evidence struct {
	PID    int
	Gen    int64
	Kind   SignalKind
	Weight float64
	Detail string
	TS     time.Time
}
