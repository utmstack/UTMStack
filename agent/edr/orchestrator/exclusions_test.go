package orchestrator

import "testing"

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

func TestExcluderSetHotSwap(t *testing.T) {
	e := NewExcluder([]string{`C:\a`})
	if !e.Excluded(`C:\a\x`) || e.Excluded(`C:\b\y`) {
		t.Fatal("initial patterns wrong")
	}
	e.Set([]string{`C:\b`})
	if e.Excluded(`C:\a\x`) {
		t.Fatal("old pattern still active after Set")
	}
	if !e.Excluded(`C:\b\y`) {
		t.Fatal("new pattern not active after Set")
	}
}
