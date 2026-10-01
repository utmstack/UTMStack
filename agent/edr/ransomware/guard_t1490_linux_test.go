//go:build linux

package ransomware

import "strings"

func t1490Scenario() (string, string, string, func(string) bool) {
	trust := func(image string) bool {
		return strings.HasPrefix(strings.ToLower(image), "/opt/restic/")
	}
	return `/opt/restic/restic`, `/usr/sbin/btrfs`,
		"btrfs subvolume delete /data/snap1", trust
}
