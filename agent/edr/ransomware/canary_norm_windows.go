//go:build windows

package ransomware

import "strings"

// canaryNorm lower-cases: Windows filesystems are case-insensitive, so a
// feed path that differs only in case is the same planted file.
func canaryNorm(s string) string { return strings.ToLower(s) }
