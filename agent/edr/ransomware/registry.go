package ransomware

import (
	"context"
	"strings"
)

// sensitiveKeyPatterns are the registry subtrees whose mutation disables
// Windows recovery (System Restore) or shadow copies. A mutating op on a key
// whose name (case-insensitively) contains one of these is recovery tampering
// (T1490 via the registry) and is scored as KindRegistryTamper. The list is
// deliberately short and specific — broad matches would false-positive on
// normal admin tooling.
var sensitiveKeyPatterns = []string{
	"systemrestore",
	"shadow",
	"recovery",
	"restoresettings",
}

// isSensitiveKey reports whether a registry key name is one whose mutation
// disables a recovery capability. Matching is case-insensitive substring.
func isSensitiveKey(key string) bool {
	k := strings.ToLower(key)
	for _, p := range sensitiveKeyPatterns {
		if strings.Contains(k, p) {
			return true
		}
	}
	return false
}

// RegistryEvent is a per-process registry mutation from the Kernel-Registry
// feed (Windows only). Key is the full key name the provider names in the
// mutate event.
type RegistryEvent struct {
	PID int
	Key string
}

// RegistryFeed delivers per-process registry mutations to sink until ctx is
// cancelled. Mirrors FileActivityFeed so the guard and service treat both
// streams uniformly.
type RegistryFeed interface {
	Run(ctx context.Context, sink func(RegistryEvent)) error
}
