//go:build !windows

package orchestrator

import "testing"

// TestExcluder pins the POSIX behaviour: matching is case-SENSITIVE
// (the filesystem is) and separator-agnostic (a Windows-style pattern from a
// shared config still matches a forward-slash path).
func TestExcluder(t *testing.T) {
	e := NewExcluder([]string{`C:\Windows\Temp`, `/home/me/build`, `*.log`, `/var/log/*.tmp`})
	cases := map[string]bool{
		`C:\Windows\Temp\x.dat`:      true,  // Windows pattern, backslashes normalized
		`C:\WINDOWS\TEMP\y`:          false, // case-sensitive: different dir
		`/home/me/build`:             true,  // the excluded dir itself
		`/home/me/build/out/a`:       true,  // nested subtree
		`/home/me/buildkite/a`:       false, // sibling dir sharing a prefix must NOT match
		`/home/me/BUILD/out/a`:       false, // case-sensitive: not a subtree
		`/home/me/app.log`:           true,  // basename glob
		`/home/me/app.LOG`:           false, // case-sensitive basename glob
		`/var/log/session.tmp`:       true,  // full-path glob
		`/var/log/session.TMP`:       false, // case-sensitive full-path glob
		`/home/me/app.exe`:           false,
	}
	for p, want := range cases {
		if got := e.Excluded(p); got != want {
			t.Fatalf("Excluded(%q) = %v, want %v", p, got, want)
		}
	}
}

// TestExcluderSetHotSwap verifies the runtime Set path on POSIX.
func TestExcluderSetHotSwap(t *testing.T) {
	e := NewExcluder([]string{`/a`})
	if !e.Excluded(`/a/x`) || e.Excluded(`/b/y`) {
		t.Fatal("initial patterns wrong")
	}
	e.Set([]string{`/b`})
	if e.Excluded(`/a/x`) {
		t.Fatal("old pattern still active after Set")
	}
	if !e.Excluded(`/b/y`) {
		t.Fatal("new pattern not active after Set")
	}
}
