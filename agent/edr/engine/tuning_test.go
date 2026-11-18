package engine

import (
	"errors"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestApplyConfiguredScanLimits(t *testing.T) {
	cfg := config.Default()
	cfg.ScanConcurrency = 1
	cfg.ScanLimits = config.ScanLimits{MaxFileSizeMB: 7, MaxScanSizeMB: 15, MaxRecursion: 3, MaxFiles: 42, MaxPartitions: 2, MaxScanTimeMS: 250}
	tn := ApplyScanLimits(DeriveTuning(64*gib, 32, ""), cfg, true)
	if tn.MaxThreads != 1 || tn.MaxFileSizeMB != 7 || tn.MaxScanSizeMB != 15 || tn.MaxRecursion != 3 || tn.MaxFiles != 42 || tn.MaxPartitions != 2 || tn.MaxScanTimeMS != 250 {
		t.Fatalf("configured budgets not effective: %+v", tn)
	}
	if !strings.Contains(tn.Reason, "explicit scan budgets") {
		t.Fatalf("reason hides explicit policy: %q", tn.Reason)
	}
}

func TestPlanTuningAppliesHostTierAndRequestedPolicy(t *testing.T) {
	cfg := config.Default()
	cfg.TierOverride = "standard"
	cfg.ScanConcurrency = 2
	cfg.ScanLimits = config.ScanLimits{MaxFileSizeMB: 7, MaxScanSizeMB: 15, MaxRecursion: 3, MaxFiles: 42, MaxScanTimeMS: 250}
	tn := planTuningWithProfiler(cfg, func() (uint64, int, error) { return 64 * gib, 32, nil })
	if tn.Tier != TierStandard || tn.MaxThreads != 2 || tn.MaxFileSizeMB != 7 || tn.MaxScanSizeMB != 15 || tn.MaxRecursion != 3 || tn.MaxFiles != 42 || tn.MaxScanTimeMS != 250 {
		t.Fatalf("host or configured policy lost: %+v", tn)
	}
	if !strings.Contains(tn.Reason, "tier forced by override") || !strings.Contains(tn.Reason, "explicit scan budgets") {
		t.Fatalf("tuning explanation lost: %q", tn.Reason)
	}
}

func TestConfiguredSizesRemainCoherent(t *testing.T) {
	base := DeriveTuning(2*gib, 2, "")
	for _, tc := range []struct {
		limits     config.ScanLimits
		file, scan int
	}{
		{config.ScanLimits{}, 25, 50},
		{config.ScanLimits{MaxFileSizeMB: 100}, 100, 100},
		{config.ScanLimits{MaxScanSizeMB: 10}, 10, 10},
		{config.ScanLimits{MaxFileSizeMB: 7, MaxScanSizeMB: 12}, 7, 12},
	} {
		cfg := config.Default()
		cfg.ScanLimits = tc.limits
		tn := ApplyScanLimits(base, cfg, true)
		if tn.MaxFileSizeMB != tc.file || tn.MaxScanSizeMB != tc.scan {
			t.Errorf("%+v => %d/%d, want %d/%d", tc.limits, tn.MaxFileSizeMB, tn.MaxScanSizeMB, tc.file, tc.scan)
		}
	}
}

func TestDirectConfigCannotExceedHardResourceCeilings(t *testing.T) {
	cfg := config.Default()
	cfg.ScanConcurrency = 100000
	cfg.ScanLimits = config.ScanLimits{MaxFileSizeMB: 100000, MaxScanSizeMB: 100000, MaxRecursion: 100000, MaxFiles: 100000, MaxPartitions: 100000, MaxScanTimeMS: 1000000}
	tn := ApplyScanLimits(DeriveTuning(64*gib, 32, ""), cfg, true)
	if tn.MaxThreads > config.MaxScanConcurrency || tn.MaxFileSizeMB > config.MaxFileSizeMB || tn.MaxScanSizeMB > config.MaxScanSizeMB || tn.MaxRecursion > FixedMaxRecursion || tn.MaxFiles > FixedMaxFiles || tn.MaxPartitions > FixedMaxPartitions || tn.MaxScanTimeMS > FixedMaxScanTime {
		t.Fatalf("bypassed hard ceilings: %+v", tn)
	}
}

func TestProfilingFallbackAppliesReductionsButCannotInflateFailsafe(t *testing.T) {
	for _, tc := range []struct {
		name  string
		probe func() (uint64, int, error)
	}{
		{"error", func() (uint64, int, error) { return 0, 0, errors.New("test profiling error") }},
		{"zero-memory", func() (uint64, int, error) { return 0, 4, nil }},
		{"zero-cores", func() (uint64, int, error) { return 64 * gib, 0, nil }},
		{"panic", func() (uint64, int, error) { panic("test profiling panic") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.Default()
			cfg.ScanConcurrency = 16
			cfg.ScanLimits = config.ScanLimits{MaxFileSizeMB: 200, MaxScanSizeMB: 400, MaxRecursion: 3, MaxFiles: 7, MaxScanTimeMS: 75}
			tn := planTuningWithProfiler(cfg, tc.probe)
			if tn.MaxThreads != 1 || tn.MaxFileSizeMB != 25 || tn.MaxScanSizeMB != 50 || tn.MaxRecursion != 3 || tn.MaxFiles != 7 || tn.MaxScanTimeMS != 75 {
				t.Fatalf("failsafe overridden: %+v", tn)
			}
			if !strings.Contains(tn.Reason, "failsafe") {
				t.Fatalf("fallback not observable: %+v", tn)
			}
		})
	}
}

func TestTinyHostLimitsInputWork(t *testing.T) {
	tn := DeriveTuning(1*gib, 8, "server")
	if tn.MaxThreads != 1 || tn.MaxFileSizeMB != 25 || tn.MaxScanSizeMB != 25 {
		t.Fatalf("small host must bound work even under an override: %+v", tn)
	}
}

func TestDeriveTuningConstrained(t *testing.T) {
	tn := DeriveTuning(2*gib, 2, "")
	if tn.Tier != TierConstrained {
		t.Fatalf("2GB/2core → constrained resident, got %+v", tn)
	}
	if tn.MaxThreads > 2 {
		t.Fatalf("constrained threads = %d, want ≤2", tn.MaxThreads)
	}
}

func TestDeriveTuningStandard(t *testing.T) {
	tn := DeriveTuning(8*gib, 4, "")
	if tn.Tier != TierStandard {
		t.Fatalf("8GB/4core → standard, got %s", tn.Tier)
	}
	if tn.MaxThreads != 4 {
		t.Fatalf("standard threads = %d, want 4", tn.MaxThreads)
	}
}

func TestDeriveTuningServerCapsThreadsAt16(t *testing.T) {
	tn := DeriveTuning(64*gib, 32, "")
	if tn.Tier != TierServer {
		t.Fatalf("64GB/32core → server, got %s", tn.Tier)
	}
	if tn.MaxThreads != 16 {
		t.Fatalf("server threads = %d, want 16 (capped)", tn.MaxThreads)
	}
	if tn.MaxFileSizeMB < 100 {
		t.Fatalf("server size cap too small: %d", tn.MaxFileSizeMB)
	}
}

func TestSecurityBoundsAlwaysStatic(t *testing.T) {
	for _, ram := range []uint64{1 * gib, 2 * gib, 8 * gib, 128 * gib} {
		tn := DeriveTuning(ram, 32, "")
		if tn.MaxRecursion != FixedMaxRecursion || tn.MaxFiles != FixedMaxFiles {
			t.Fatalf("security bounds scaled at %d bytes: %+v", ram, tn)
		}
	}
}

func TestSafeDefaultTuningIsLeanFunctionalAndSafe(t *testing.T) {
	tn := SafeDefaultTuning()
	if tn.MaxThreads != 1 {
		t.Fatalf("failsafe threads = %d, want 1 (leanest)", tn.MaxThreads)
	}
	// security bounds respected, size caps small
	if tn.MaxRecursion != FixedMaxRecursion || tn.MaxFiles != FixedMaxFiles {
		t.Fatalf("failsafe violates security bounds: %+v", tn)
	}
	if tn.MaxFileSizeMB > 50 {
		t.Fatalf("failsafe file cap too large for a low-spec host: %d", tn.MaxFileSizeMB)
	}

}

func TestOverrideForcesTier(t *testing.T) {
	// A big host forced down to constrained.
	tn := DeriveTuning(64*gib, 32, "constrained")
	if tn.Tier != TierConstrained || tn.MaxThreads > 2 {
		t.Fatalf("override to constrained failed: %+v", tn)
	}
}

// The native engine analyses in memory and never extracts to a temp directory,
// so free space on <install>\engine\tmp (present only on endpoints upgraded
// from the old daemon) must not shrink the per-file scan cap: the limits are
// the tier's, and a fresh and an upgraded install on the same host agree.
func TestFreeTempDoesNotShrinkInMemoryCaps(t *testing.T) {
	for _, tc := range []struct {
		ram              uint64
		cores            int
		file, scan, thrd int
	}{
		{8 * gib, 4, 100, 200, 4},
		{64 * gib, 16, 200, 400, 16},
		{2 * gib, 2, 25, 50, 2},
	} {
		tn := DeriveTuning(tc.ram, tc.cores, "")
		if tn.MaxFileSizeMB != tc.file || tn.MaxScanSizeMB != tc.scan || tn.MaxThreads != tc.thrd {
			t.Fatalf("%d bytes/%d cores: got %d/%d MB threads=%d, want %d/%d MB threads=%d",
				tc.ram, tc.cores, tn.MaxFileSizeMB, tn.MaxScanSizeMB, tn.MaxThreads, tc.file, tc.scan, tc.thrd)
		}
	}
}

// Engine.Tuning must keep the derivation reason visible in status.json, so a
// forced or downgraded tier is explained.
func TestEngineTuningKeepsDerivationReason(t *testing.T) {
	cfg := config.Default()
	cfg.TierOverride = string(TierConstrained)
	tn := New(cfg).Tuning()
	if (!strings.Contains(tn.Reason, "tier forced by override") && !strings.Contains(tn.Reason, "failsafe: host profiling unavailable")) || !strings.Contains(tn.Reason, "native Go engine") {
		t.Fatalf("reason lost: %q", tn.Reason)
	}
}
