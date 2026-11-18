//go:build linux

package dependency

import "testing"

func TestEDRLoadedFailedSystemdUnit(t *testing.T) {
	for _, output := range []string{"LoadState=loaded\nActiveState=failed\n", "ActiveState=failed\nLoadState=loaded\n"} {
		if !edrLoadedFailedSystemdUnit(output) {
			t.Fatalf("installed failed unit not recognized: %q", output)
		}
	}
	for _, output := range []string{"", "ActiveState=failed\n", "LoadState=not-found\nActiveState=failed\n", "LoadState=masked\nActiveState=failed\n", "LoadState=error\nActiveState=failed\n", "LoadState=loaded\nActiveState=active\n", "LoadState=loaded\nActiveState=inactive\n", "LoadState=loaded\nActiveState=activating\n", "LoadState=loaded\nLoadState=loaded\n", "LoadState=loaded\nActiveState=failed\nActiveState=active\n"} {
		if edrLoadedFailedSystemdUnit(output) {
			t.Fatalf("ambiguous or different unit state overrides service error: %q", output)
		}
	}
}
