package engine

import "testing"

const bigTemp = 500 * gib // effectively unbounded temp for these tests

func TestDeriveTuningTinyHostNotViable(t *testing.T) {
	// 1 GB / 1 core: resident daemon is not viable → graceful downgrade.
	tn := DeriveTuning(1*gib, 1, bigTemp, "")
	if tn.ResidentViable {
		t.Fatal("1GB host must not run a resident daemon")
	}
	if tn.ConcurrentReload {
		t.Fatal("concurrent reload must be off on a tiny host")
	}
}

func TestDeriveTuningConstrained(t *testing.T) {
	tn := DeriveTuning(2*gib, 2, bigTemp, "")
	if tn.Tier != TierConstrained || !tn.ResidentViable {
		t.Fatalf("2GB/2core → constrained resident, got %+v", tn)
	}
	if tn.MaxThreads > 2 {
		t.Fatalf("constrained threads = %d, want ≤2", tn.MaxThreads)
	}
	if tn.ConcurrentReload {
		t.Fatal("constrained: reload off (RAM can't absorb 2x spike)")
	}
}

func TestDeriveTuningStandard(t *testing.T) {
	tn := DeriveTuning(8*gib, 4, bigTemp, "")
	if tn.Tier != TierStandard {
		t.Fatalf("8GB/4core → standard, got %s", tn.Tier)
	}
	if tn.MaxThreads != 4 {
		t.Fatalf("standard threads = %d, want 4", tn.MaxThreads)
	}
	if !tn.ConcurrentReload {
		t.Fatal("standard (8GB) → concurrent reload on")
	}
}

func TestDeriveTuningServerCapsThreadsAt16(t *testing.T) {
	tn := DeriveTuning(64*gib, 32, bigTemp, "")
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
		tn := DeriveTuning(ram, 32, bigTemp, "")
		if tn.MaxRecursion != FixedMaxRecursion || tn.MaxFiles != FixedMaxFiles {
			t.Fatalf("security bounds scaled at %d bytes: %+v", ram, tn)
		}
	}
}

func TestFDConstraintHolds(t *testing.T) {
	tn := DeriveTuning(64*gib, 32, bigTemp, "")
	if tn.MaxThreads*tn.MaxRecursion+tn.MaxQueue+6 >= fdLimit {
		t.Fatalf("fd constraint violated: %+v", tn)
	}
}

func TestSizeCapsClampToFreeTemp(t *testing.T) {
	// Server tier wants MaxFileSize 200M; a tiny temp must force it down.
	small := uint64(2 * gib) // free temp
	tn := DeriveTuning(64*gib, 16, small, "")
	worst := uint64(tn.MaxThreads) * uint64(tn.MaxFileSizeMB) * mib * uint64(tn.MaxRecursion)
	if worst > small/2 {
		t.Fatalf("worst-case unpack %d exceeds half of free temp %d: %+v", worst, small, tn)
	}
}

func TestSafeDefaultTuningIsLeanFunctionalAndSafe(t *testing.T) {
	tn := SafeDefaultTuning()
	if !tn.ResidentViable {
		t.Fatal("failsafe must keep the engine functional")
	}
	if tn.MaxThreads != 1 {
		t.Fatalf("failsafe threads = %d, want 1 (leanest)", tn.MaxThreads)
	}
	if tn.ConcurrentReload {
		t.Fatal("failsafe must not enable concurrent reload (avoids 2x RAM spike)")
	}
	// security bounds respected, size caps small
	if tn.MaxRecursion != FixedMaxRecursion || tn.MaxFiles != FixedMaxFiles {
		t.Fatalf("failsafe violates security bounds: %+v", tn)
	}
	if tn.MaxFileSizeMB > 50 {
		t.Fatalf("failsafe file cap too large for a low-spec host: %d", tn.MaxFileSizeMB)
	}
	// the fd constraint must hold for the failsafe too
	if tn.MaxThreads*tn.MaxRecursion+tn.MaxQueue+6 >= fdLimit {
		t.Fatalf("failsafe violates fd constraint: %+v", tn)
	}
}

func TestOverrideForcesTier(t *testing.T) {
	// A big host forced down to constrained.
	tn := DeriveTuning(64*gib, 32, bigTemp, "constrained")
	if tn.Tier != TierConstrained || tn.MaxThreads > 2 {
		t.Fatalf("override to constrained failed: %+v", tn)
	}
}
