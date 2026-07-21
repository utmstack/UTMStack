package registry

import "strings"

func NormalizeKey(s string) string {
	lower := strings.ToLower(s)
	parts := strings.Fields(lower)
	return strings.Join(parts, "_")
}
