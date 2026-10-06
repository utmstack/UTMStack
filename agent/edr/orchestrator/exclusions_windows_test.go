//go:build windows

package orchestrator

import "testing"

// TestExcluder pins the Windows behaviour: case-insensitive matching and
// separator agnosticism.
func TestExcluder(t *testing.T) {
	e := NewExcluder([]string{`C:\Windows\Temp`, `C:\Users\me\build`, `*.log`, `C:\logs\*.tmp`})
	cases := map[string]bool{
		`C:\Windows\Temp\x.dat`:      true,  // subtree of an excluded dir
		`c:\windows\temp\y`:          true,  // case-insensitive
		`C:\Windows\Temp`:            true,  // the excluded dir itself
		`C:\Users\me\build\out\a`:    true,  // nested subtree
		`C:\Users\me\buildkite\a`:    false, // sibling dir sharing a prefix must NOT match
		`C:\Users\a\app.log`:         true,  // basename glob
		`C:\logs\session.tmp`:        true,  // full-path glob
		`C:\Users\a\a.exe`:           false,
	}
	for p, want := range cases {
		if got := e.Excluded(p); got != want {
			t.Fatalf("Excluded(%q) = %v, want %v", p, got, want)
		}
	}
}
