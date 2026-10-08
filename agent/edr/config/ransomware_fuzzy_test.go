package config

import "testing"

// TestX1FuzzyWeightsBelowSuspend pins the X1 invariant: every fuzzy weight in
// Default() must stay strictly below the default suspend threshold so no
// single behavioural signal can escalate on its own. Canary/T1490 are
// max-weight and are not part of this check.
func TestX1FuzzyWeightsBelowSuspend(t *testing.T) {
	d := Default()
	suspend := d.Ransomware.SuspendThreshold
	for _, w := range []int{
		d.Ransomware.Fuzzy.Entropy,
		d.Ransomware.Fuzzy.ExtChurn,
		d.Ransomware.Fuzzy.MassRate,
		d.Ransomware.Fuzzy.RansomNote,
		d.Ransomware.Fuzzy.Registry,
	} {
		if w >= suspend {
			t.Fatalf("fuzzy weight %d must be < suspend threshold %d", w, suspend)
		}
	}
}

// TestX1FuzzyLoadBackfill verifies a partial fuzzy/behavior block in edr.json
// keeps the default for absent fields and honours the set ones.
func TestX1FuzzyLoadBackfill(t *testing.T) {
	c := Default()
	// Partial fuzzy: operator lowers only entropy.
	c.Ransomware.Fuzzy.Entropy = 20
	c.Ransomware.Fuzzy = FuzzySensors{Entropy: 20}
	if got := c.Ransomware.FuzzyWeight("entropy"); got != 20 {
		t.Fatalf("FuzzyWeight(entropy) = %d, want 20", got)
	}
	if got := c.Ransomware.FuzzyWeight("mass_rate"); got != 35 {
		t.Fatalf("FuzzyWeight(mass_rate) = %d, want default 35", got)
	}
	// FuzzyWeight falls back to defaults for a zero field.
	if got := c.Ransomware.FuzzyWeight("ransom_note"); got != 30 {
		t.Fatalf("FuzzyWeight(ransom_note) fallback = %d, want 30", got)
	}
	if got := c.Ransomware.FuzzyWeight("not_a_kind"); got != 0 {
		t.Fatalf("FuzzyWeight(unknown) = %d, want 0", got)
	}
}
