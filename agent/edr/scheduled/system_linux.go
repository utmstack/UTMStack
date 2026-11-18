//go:build linux

package scheduled

import "strings"

func skipSystemPath(path string) bool {
	for _, root := range []string{"/proc", "/sys", "/dev"} {
		if path == root || strings.HasPrefix(path, root+"/") {
			return true
		}
	}
	return false
}
