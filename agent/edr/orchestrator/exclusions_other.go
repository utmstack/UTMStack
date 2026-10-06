//go:build !windows

package orchestrator

import "strings"

// platformNormPath keeps the original case (POSIX filesystems are
// case-sensitive) and canonicalises backslashes to forward slashes.
func platformNormPath(s string) string {
	return strings.TrimRight(strings.ReplaceAll(s, `\`, "/"), "/")
}
