package main

import (
	"strings"
	"testing"
)

func TestParseAllowlistValue(t *testing.T) {
	got := parseAllowlistValue(" 198.51.100.10, 203.0.113.0/24\nExample.com, 198.51.100.10 , ")
	want := []string{"198.51.100.10", "203.0.113.0/24", "Example.com"}
	if len(got) != len(want) {
		t.Fatalf("len = %d (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestBuildFeedsConfigYAML(t *testing.T) {
	yamlBytes, err := buildFeedsConfigYAML(true, []string{"198.51.100.10", "203.0.113.0/24", "*.example.com"})
	if err != nil {
		t.Fatalf("buildFeedsConfigYAML failed: %v", err)
	}
	out := string(yamlBytes)
	for _, sub := range []string{"enabled: true", "allowlist:", "198.51.100.10", "203.0.113.0/24", "*.example.com"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected YAML to contain %q, got:\n%s", sub, out)
		}
	}
}
