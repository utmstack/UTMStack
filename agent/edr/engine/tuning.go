package engine

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
	FixedMaxRecursion = 16
	FixedMaxFiles     = 10000
	FixedMaxScanTime  = 120000 // ms

	mib = 1 << 20
	gib = 1 << 30

	// Resident daemon needs the full signature DB in RAM (>1GB) plus headroom.
	dbResidentRAM = 1300 * mib // ~1.3 GB signature set resident
	minResidentRAM = 1600 * mib // below this, a resident daemon risks OOM
	reloadSpikeRAM = 3 * gib    // concurrent reload briefly needs ~2x the DB
	osBaselineRAM  = 512 * mib  // OS + agent headroom to leave free

	fdLimit = 1024 // conservative RLIMIT_NOFILE assumption (§8.2)
)

// Tuning is the derived, host-specific engine configuration.
type Tuning struct {
	Tier             Tier
	ResidentViable   bool // false => don't run a resident daemon (§8.4)
	MaxThreads       int
	MaxQueue         int
	MaxScanSizeMB    int
	MaxFileSizeMB    int
	MaxRecursion     int // always FixedMaxRecursion
	MaxFiles         int // always FixedMaxFiles
	ConcurrentReload bool
	Reason           string // human-readable note on the decision / any downgrade
}

// DeriveTuning computes the resource knobs from host RAM, CPU cores, and free
// temp space, honoring the fixed security bounds and the viability gate (§8.4).
// override (optional: "constrained"|"standard"|"server") forces a tier; the
// viability gate and security bounds still apply.
func DeriveTuning(ramBytes uint64, cores int, freeTempBytes uint64, override string) Tuning {
	if cores < 1 {
		cores = 1
	}

	t := Tuning{
		MaxRecursion: FixedMaxRecursion,
		MaxFiles:     FixedMaxFiles,
	}

	// --- Tier selection (§8.3) ---
	tier := selectTier(ramBytes, cores)
	switch Tier(override) {
	case TierConstrained, TierStandard, TierServer:
		tier = Tier(override)
		t.Reason = "tier forced by override; "
	}
	t.Tier = tier

	// --- Viability gate (§8.4) ---
	if ramBytes < minResidentRAM {
		t.ResidentViable = false
		t.Tier = TierConstrained
		t.MaxThreads = 1
		t.MaxQueue = 2
		t.MaxScanSizeMB = 25
		t.MaxFileSizeMB = 25
		t.ConcurrentReload = false
		t.Reason += "RAM below resident-daemon threshold; scanning layer downgraded (on-demand/skip)"
		return t
	}
	t.ResidentViable = true

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

	// --- Size caps (§8.3), within the 2GB hard file limit ---
	switch tier {
	case TierConstrained:
		t.MaxScanSizeMB, t.MaxFileSizeMB = 50, 25
	case TierStandard:
		t.MaxScanSizeMB, t.MaxFileSizeMB = 200, 100
	default: // Server
		t.MaxScanSizeMB, t.MaxFileSizeMB = 400, 200
	}

	// --- Concurrent reload only where RAM absorbs the ~2x spike (§7, §8.1) ---
	t.ConcurrentReload = ramBytes >= reloadSpikeRAM

	// --- MaxQueue = 2*threads, subject to the fd constraint (§8.2) ---
	t.MaxQueue = 2 * t.MaxThreads
	for t.MaxThreads*t.MaxRecursion+t.MaxQueue+6 >= fdLimit && t.MaxQueue > t.MaxThreads {
		t.MaxQueue--
	}

	// --- Clamp so the worst-case unpack footprint fits free temp ---
	// worst-case ≈ MaxThreads × MaxFileSize × MaxRecursion (§8.2); keep it under
	// half of free temp so unpacking cannot fill the partition. Reduce file size
	// first, then shed threads (the other multiplier), down to a safe floor.
	const minFileMB = 10
	if freeTempBytes > 0 {
		clamped := false
		for {
			worst := uint64(t.MaxThreads) * uint64(t.MaxFileSizeMB) * mib * uint64(t.MaxRecursion)
			if worst <= freeTempBytes/2 {
				break
			}
			switch {
			case t.MaxFileSizeMB > minFileMB:
				t.MaxFileSizeMB -= 5
				if t.MaxScanSizeMB > t.MaxFileSizeMB*2 {
					t.MaxScanSizeMB = t.MaxFileSizeMB * 2
				}
			case t.MaxThreads > 1:
				t.MaxThreads--
				t.MaxQueue = 2 * t.MaxThreads
			default:
				worst = 0 // floor reached; accept the minimum footprint
			}
			clamped = true
			if worst == 0 {
				break
			}
		}
		if clamped {
			t.Reason += "resource knobs reduced to fit free temp space; "
		}
	}

	if t.Reason == "" {
		t.Reason = "derived from host RAM/cores within security bounds"
	}
	return t
}

// SafeDefaultTuning is the failsafe used when the host cannot be profiled
// (RAM/core/temp probe failed or returned nonsense). It is the leanest
// *functional* resident configuration: a single worker, small size caps, and
// NO concurrent reload (avoiding the ~2x signature-RAM spike). It keeps
// detection working while being very unlikely to strain even a low-spec host.
func SafeDefaultTuning() Tuning {
	return Tuning{
		Tier:             TierConstrained,
		ResidentViable:   true,
		MaxThreads:       1,
		MaxQueue:         2,
		MaxScanSizeMB:    50,
		MaxFileSizeMB:    25,
		MaxRecursion:     FixedMaxRecursion,
		MaxFiles:         FixedMaxFiles,
		ConcurrentReload: false,
		Reason:           "failsafe: host profiling unavailable — conservative low-spec defaults",
	}
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
