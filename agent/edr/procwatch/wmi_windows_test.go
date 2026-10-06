//go:build windows

package procwatch

import (
	"strconv"
	"testing"
)

// The WMI instance-creation generation interval must be the
// narrowed 0.5 s window (5 in 0.1 s units), not the old 1 s one that let
// sub-second processes slip past the watcher. WMI rejects 0, and going below
// 5 would add CPU cost for a window the platform does not need.
func TestWMIWithinIntervalNarrowed(t *testing.T) {
	if wmiWithin100ms >= 10 {
		t.Fatalf("wmiWithin100ms = %d (>= 1 s), want < 10 (0.5 s window)", wmiWithin100ms)
	}
	if wmiWithin100ms < 1 {
		t.Fatalf("wmiWithin100ms = %d, want >= 1 (WMI rejects 0)", wmiWithin100ms)
	}
	if got := strconv.Itoa(wmiWithin100ms); got != "5" {
		t.Fatalf("wmiWithin100ms = %s, want 5 (0.5 s)", got)
	}
}
