package ransomware

import "time"

// SignalKind identifies a ransomware detector's output. Canary and T1490 are
// the two max-weight (deterministic) kinds; the fuzzy kinds (X1) are
// sub-suspend by default so no single one can escalate — the scorer's
// kind-diversity bonus is what lets several agree.
type SignalKind string

const (
	KindCanary    SignalKind = "canary"
	KindT1490     SignalKind = "t1490"
	KindEntropy   SignalKind = "entropy"   // X1: rewritten-file content randomness
	KindExtChurn  SignalKind = "ext_churn" // X1: many files taking a new unusual extension
	KindMassRate  SignalKind = "mass_rate" // X1: writes/PID faster than normal software
	KindRansomNote SignalKind = "ransom_note" // X1: ransom-note-named file write/create
	KindRegistry  SignalKind = "registry"  // X1 (Windows): recovery/VSS key tamper
)

// DefaultWeights maps a kind to its evidence weight. Canary and T1490 are
// max-weight (a single hit crosses the default kill threshold of 100). Every
// fuzzy weight is below the default suspend threshold (50). Config may lower
// any of them further (or raise below 50) without code changes (H9 tuning).
var DefaultWeights = map[SignalKind]float64{
	KindCanary:     100,
	KindT1490:      100,
	KindEntropy:    40,
	KindExtChurn:   40,
	KindMassRate:   35,
	KindRansomNote: 30,
	KindRegistry:   35,
}

// FuzzyKinds are the X1 behavioural kinds; each carries a tunable weight on
// RansomwareConfig that must stay below the suspend threshold.
var FuzzyKinds = []SignalKind{KindEntropy, KindExtChurn, KindMassRate, KindRansomNote, KindRegistry}

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
