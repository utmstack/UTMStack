package config

import "testing"

func TestDefaultBlocklist(t *testing.T) {
	d := Default()
	b := d.Blocklist
	if !b.Enabled || !b.Enforce {
		t.Fatalf("blocklist should default enabled+enforcing, got %+v", b)
	}
	if !b.AllowPrivateRanges {
		t.Fatal("private ranges must be allowlisted by default")
	}
	if b.Direction != "both" {
		t.Fatalf("direction default = both, got %q", b.Direction)
	}
	if len(b.Levels) != 1 || b.Levels[0] != 1 {
		t.Fatalf("default levels = [1], got %v", b.Levels)
	}
	if b.RefreshHours != 6 {
		t.Fatalf("refresh default 6, got %d", b.RefreshHours)
	}
	if len(b.IndicatorTypes) == 0 || b.IndicatorTypes[0] != "ip" {
		t.Fatalf("indicator types must include ip, got %v", b.IndicatorTypes)
	}
}
