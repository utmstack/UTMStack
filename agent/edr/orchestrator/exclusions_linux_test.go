//go:build linux

package orchestrator

import "testing"

func TestLinuxExclusionsPreserveCaseAndBackslash(t *testing.T) {
	e := NewExcluder([]string{"/Trusted", "*.LOG", `/literal\name`})
	for p, want := range map[string]bool{"/Trusted/a": true, "/trusted/a": false, "/TrustedElse/a": false, "/x/A.LOG": true, "/x/a.log": false, `/literal\name/x`: true, "/literal/name/x": false} {
		if got := e.Excluded(p); got != want {
			t.Errorf("Excluded(%q)=%v want %v", p, got, want)
		}
	}
}
