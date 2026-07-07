package orchestrator

import (
	"path" // forward-slash globbing, OS-independent (paths are normalized below)
	"strings"
	"sync"
)

// Excluder matches paths/images against a set of patterns. Its pattern set is
// swappable at runtime via Set (under an RWMutex) so a config reload can update
// the allowlists live — every holder of the *Excluder sees the new patterns
// without being re-wired.
type Excluder struct {
	mu       sync.RWMutex
	patterns []string
}

// normPath lower-cases and converts a Windows or POSIX path to a canonical
// forward-slash form so matching is case- and separator-insensitive on any host
// (the EDR handles Windows paths but its tests run on the dev host).
func normPath(s string) string {
	return strings.TrimRight(strings.ReplaceAll(strings.ToLower(s), `\`, "/"), "/")
}

func normPatterns(patterns []string) []string {
	low := make([]string, 0, len(patterns))
	for _, p := range patterns {
		if np := normPath(p); np != "" {
			low = append(low, np)
		}
	}
	return low
}

func NewExcluder(patterns []string) *Excluder {
	return &Excluder{patterns: normPatterns(patterns)}
}

// Set replaces the pattern set (used on config reload to hot-apply an updated
// allowlist). Safe for concurrent use with Excluded.
func (e *Excluder) Set(patterns []string) {
	np := normPatterns(patterns)
	e.mu.Lock()
	e.patterns = np
	e.mu.Unlock()
}

// Excluded reports whether path is covered by any configured exclusion. A
// pattern is matched as:
//   - a directory / path prefix on a path boundary — "C:\build" excludes
//     "C:\build" and everything under "C:\build\...", but NOT a sibling like
//     "C:\buildkite";
//   - a glob against the full path (e.g. "C:\logs\*.tmp"); or
//   - a glob against the file's base name (e.g. "*.log").
// Matching is case-insensitive and separator-agnostic (\\ or /).
func (e *Excluder) Excluded(path_ string) bool {
	p := normPath(path_)
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, pat := range e.patterns { // already normalized in Set/NewExcluder
		// Directory / path prefix match, respecting path boundaries so a
		// configured directory only excludes itself and its subtree.
		if p == pat || strings.HasPrefix(p, pat+"/") {
			return true
		}
		if ok, _ := path.Match(pat, p); ok {
			return true
		}
		if ok, _ := path.Match(pat, path.Base(p)); ok {
			return true
		}
	}
	return false
}
