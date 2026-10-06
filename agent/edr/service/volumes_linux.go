//go:build linux

package service

import (
	"os"
	"strings"
)

// defaultFixedVolumes enumerates the local disk mount points under
// /proc/mounts and returns them as absolute paths, mirroring the Windows
// "all fixed drives" default. Pseudo and network filesystems (proc, sysfs,
// cgroup*, tmpfs, devtmpfs, NFS, CIFS, 9P, fuse mounts, …) are excluded:
// they are not user data and several cannot be marked by the fanotify
// watcher anyway. Deduplicated by device so multiple mounts of the same disk
// are reported once (the first mount point wins).
func defaultFixedVolumes() []string {
	vols, err := localMounts("/proc/mounts")
	if err != nil {
		return []string{"/"}
	}
	if len(vols) == 0 {
		return []string{"/"}
	}
	return vols
}

// localMounts parses a /proc/mounts-formatted table (whitespace-separated:
// device mountpoint fstype options dump pass, one mount per line) and returns
// the mount points whose fstype is a local filesystem. It takes the table
// path so the parser is testable against a fixture file without a live mount
// table.
func localMounts(tablePath string) ([]string, error) {
	data, err := os.ReadFile(tablePath)
	if err != nil {
		return nil, err
	}
	var out []string
	seen := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		device, point, fstype := fields[0], fields[1], fields[2]
		if device == "none" || device == "overlay" {
			continue // initramfs stub / container root, not user data
		}
		if !isLocalType(fstype) {
			continue
		}
		if seen[device] {
			continue
		}
		seen[device] = true
		out = append(out, point)
	}
	return out, nil
}

// isLocalType reports whether fstype is a plain local disk filesystem worth
// watching. The allowlist keeps this forward-compatible: any new pseudo or
// network type defaults to excluded.
func isLocalType(fstype string) bool {
	switch fstype {
	case "ext2", "ext3", "ext4", "xfs", "btrfs", "zfs", "vfat", "ntfs",
		"ntfs3", "exfat", "f2fs", "jfs", "reiserfs":
		return true
	default:
		return false
	}
}
