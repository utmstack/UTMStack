package engine

import "github.com/utmstack/UTMStack/agent/edr/config"

// Host-adaptive engine tuning per the ClamAV Configuration & Tuning Guide (§8).
// Feature/security settings are static (applied identically everywhere); only
// the resource knobs below are derived from the host, and derived values may
// never cross the fixed security bounds.

type Tier string

const (
	TierConstrained Tier = "constrained"
	TierStandard    Tier = "standard"
	TierServer      Tier = "server"
)

// Fixed security bounds (§6, §8.1) — NEVER scaled up for a bigger host.
const (
	FixedMaxRecursion  = config.MaxRecursion
	FixedMaxFiles      = config.MaxFiles
	FixedMaxPartitions = config.MaxPartitions
	FixedMaxScanTime   = config.MaxScanTimeMS // ms

	mib = 1 << 20
	gib = 1 << 30
)

// Tuning is the derived, host-specific engine configuration.
type Tuning struct {
	Tier          Tier   `json:"tier"`
	MaxThreads    int    `json:"max_threads"`
	MaxScanSizeMB int    `json:"max_scan_size_mb"`
	MaxFileSizeMB int    `json:"max_file_size_mb"`
	MaxRecursion  int    `json:"max_recursion"`    // at most FixedMaxRecursion
	MaxFiles      int    `json:"max_files"`        // at most FixedMaxFiles
	MaxPartitions int    `json:"max_partitions"`   // default 50, at most FixedMaxPartitions
	MaxScanTimeMS int    `json:"max_scan_time_ms"` // includes waiting for a native-engine worker
	Reason        string `json:"reason"`           // human-readable note on the decision / any downgrade
}

// DeriveTuning computes the resource knobs from host RAM and CPU cores,
// honoring the fixed security bounds. The native
// engine analyzes in memory and never extracts to disk, so free disk space
// does not bound it. override (optional: "constrained"|"standard"|"server")
// forces a tier; small-memory and security bounds still apply.
func DeriveTuning(ramBytes uint64, cores int, override string) Tuning {
	if cores < 1 {
		cores = 1
	}

	t := Tuning{
		MaxRecursion:  FixedMaxRecursion,
		MaxFiles:      FixedMaxFiles,
		MaxPartitions: 50,
		MaxScanTimeMS: FixedMaxScanTime,
	}

	// --- Tier selection (§8.3) ---
	tier := selectTier(ramBytes, cores)
	switch Tier(override) {
	case TierConstrained, TierStandard, TierServer:
		tier = Tier(override)
		t.Reason = "tier forced by override; "
	}
	t.Tier = tier

	// --- Threads (§8.2): min(cores,16), capped tighter on smaller tiers ---
	switch tier {
	case TierConstrained:
		t.MaxThreads = minInt(cores, 2)
	case TierStandard:
		t.MaxThreads = minInt(cores, 4)
	default: // Server
		t.MaxThreads = minInt(cores, 16)
	}
	if t.MaxThreads < 1 {
		t.MaxThreads = 1
	}

	// --- Native size defaults, within the shared hard resource ceilings ---
	switch tier {
	case TierConstrained:
		t.MaxScanSizeMB, t.MaxFileSizeMB = 50, 25
	case TierStandard:
		t.MaxScanSizeMB, t.MaxFileSizeMB = 200, 100
	default: // Server
		t.MaxScanSizeMB, t.MaxFileSizeMB = 400, 200
	}

	// Input budgets limit work only. They do not promise that the compiled
	// signature set or overlapping reload generations fit the host's RAM.
	if ramBytes < 2*gib {
		t.MaxThreads = 1
		t.MaxFileSizeMB, t.MaxScanSizeMB = 25, 25
		t.Reason += "small-memory host; single worker and small input budgets; "
	}

	t.Reason += "derived from host RAM/cores within security bounds"
	return t
}

// SafeDefaultTuning is the failsafe used when the host cannot be profiled
// (RAM/core probe failed or returned nonsense). It is the leanest
// *functional* planning configuration: a single worker and small size caps.
// Signature-database and atomic-reload memory are separate from
// these scan budgets, and no amount of RAM sufficiency is guaranteed here.
func SafeDefaultTuning() Tuning {
	return Tuning{
		Tier:          TierConstrained,
		MaxThreads:    1,
		MaxScanSizeMB: 50,
		MaxFileSizeMB: 25,
		MaxRecursion:  FixedMaxRecursion,
		MaxFiles:      FixedMaxFiles,
		MaxPartitions: 50,
		MaxScanTimeMS: FixedMaxScanTime,
		Reason:        "failsafe: host profiling unavailable — conservative low-spec defaults",
	}
}

// ApplyScanLimits applies explicit analysis budgets to a derived tier. It also
// applies scan_concurrency to the engine itself, so AMSI and on-demand callers
// share the same ceiling as the real-time scanner. Config validation reports
// invalid input, but direct Go callers cannot bypass the hard ceilings here.
// If host profiling failed, overrides may only reduce the failsafe budgets.
func ApplyScanLimits(t Tuning, cfg config.EDRConfig, profileAvailable bool) Tuning {
	limits := cfg.ScanLimits
	if cfg.ScanConcurrency > 0 {
		t.MaxThreads = minInt(t.MaxThreads, minInt(cfg.ScanConcurrency, config.MaxScanConcurrency))
	}
	selectLimit := func(current, requested, ceiling int) int {
		current = minInt(current, ceiling)
		if requested > 0 {
			requested = minInt(requested, ceiling)
			if profileAvailable || requested < current {
				return requested
			}
		}
		return current
	}
	t.MaxFileSizeMB = selectLimit(t.MaxFileSizeMB, limits.MaxFileSizeMB, config.MaxFileSizeMB)
	t.MaxScanSizeMB = selectLimit(t.MaxScanSizeMB, limits.MaxScanSizeMB, config.MaxScanSizeMB)
	t.MaxRecursion = selectLimit(t.MaxRecursion, limits.MaxRecursion, FixedMaxRecursion)
	t.MaxFiles = selectLimit(t.MaxFiles, limits.MaxFiles, FixedMaxFiles)
	t.MaxPartitions = selectLimit(t.MaxPartitions, limits.MaxPartitions, FixedMaxPartitions)
	t.MaxScanTimeMS = selectLimit(t.MaxScanTimeMS, limits.MaxScanTimeMS, FixedMaxScanTime)
	// An explicitly smaller aggregate budget also limits each object. An
	// explicit file size with an automatic aggregate budget gets enough total
	// budget for at least one full input; callers can set both for precision.
	if t.MaxFileSizeMB > t.MaxScanSizeMB {
		if limits.MaxScanSizeMB == 0 && profileAvailable {
			t.MaxScanSizeMB = t.MaxFileSizeMB
		} else {
			t.MaxFileSizeMB = t.MaxScanSizeMB
		}
	}
	if limits != (config.ScanLimits{}) {
		t.Reason += "; explicit scan budgets within hard security ceilings"
	}
	if cfg.ScanConcurrency > 0 {
		t.Reason += "; scan_concurrency applies to all engine callers"
	}
	return t
}

func selectTier(ramBytes uint64, cores int) Tier {
	switch {
	case ramBytes >= 8*gib && cores >= 8:
		return TierServer
	case ramBytes >= 4*gib:
		return TierStandard
	default:
		return TierConstrained
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
