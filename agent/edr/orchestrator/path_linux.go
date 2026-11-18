//go:build linux

package orchestrator

import "strings"

// Windows-shaped paths remain portable for policy import and tests. Absolute
// Linux paths preserve both case and literal backslashes.
func normPath(s string) string {
	if len(s) >= 3 && ((s[0] >= 'A' && s[0] <= 'Z') || (s[0] >= 'a' && s[0] <= 'z')) && s[1] == ':' && (s[2] == '/' || s[2] == '\\') || strings.HasPrefix(s, `\\`) {
		return strings.TrimRight(strings.ReplaceAll(strings.ToLower(s), `\`, "/"), "/")
	}
	return strings.TrimRight(s, "/")
}
