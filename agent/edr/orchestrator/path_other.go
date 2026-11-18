//go:build !linux

package orchestrator

import "strings"

func normPath(s string) string {
	return strings.TrimRight(strings.ReplaceAll(strings.ToLower(s), `\`, "/"), "/")
}
