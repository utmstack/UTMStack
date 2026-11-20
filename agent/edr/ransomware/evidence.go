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

	// KindMassRate and KindRansomNote are the two remaining fuzzy (X1) write
	// signals; KindRegistryTamper is the Windows-only recovery-tamper registry
	// signal (T1490 via the Kernel-Registry provider). It is in the same weight
	// class as KindT1490 — recovery tampering is a high-fidelity, single-hit
	// signal, not one of the sub-suspend fuzzy sensors.
	KindMassRate       SignalKind = "mass_rate"
	KindRansomNote     SignalKind = "ransom_note"
	KindRegistryTamper SignalKind = "registry_tamper"
)

// DefaultWeights maps a kind to its evidence weight. Canary and T1490 are
// max-weight (a single hit crosses the default kill threshold of 100). Every
// fuzzy sensor is sub-suspend: one signal alone never escalates, and only the
// scorer's kind-diversity bonus lets several agree into a containment.
var DefaultWeights = map[SignalKind]float64{
	KindCanary:              100,
	KindT1490:               100,
	KindEntropyPlaceholder:  40,
	KindExtChurnPlaceholder: 40,
	KindMassRate:            35,
	KindRansomNote:          30,
	// Registry tampering is in the T1490 weight class: a single hit crosses the
	// kill threshold (100). It is not a fuzzy sensor.
	KindRegistryTamper: 100,
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
