//go:build !linux

package proctable

import "strings"

func imageKey(s string) string {
	return strings.TrimRight(strings.ReplaceAll(strings.ToLower(s), `\`, "/"), "/")
}
