//go:build !linux

package ransomware

import (
	"path/filepath"
	"strings"
)

func normCanaryPath(p string) string {
	return strings.ToLower(filepath.ToSlash(strings.TrimRight(p, `/\`)))
}
