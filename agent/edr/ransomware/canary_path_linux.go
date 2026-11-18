//go:build linux

package ransomware

import "strings"

func normCanaryPath(p string) string {
	if len(p) > 2 && p[1] == ':' || strings.HasPrefix(p, `\\`) {
		return strings.ToLower(strings.ReplaceAll(strings.TrimRight(p, `/\`), `\`, "/"))
	}
	return strings.TrimRight(p, "/")
}
