//go:build windows

package orchestrator

import "strings"

// platformNormPath lower-cases (Windows filesystems are case-insensitive) and
// canonicalises backslashes to forward slashes.
func platformNormPath(s string) string {
	return strings.TrimRight(strings.ReplaceAll(strings.ToLower(s), `\`, "/"), "/")
}
