//go:build linux

package ransomware

// t1490Rules — Linux recovery/backup-destruction commands (MITRE T1490).
//
// Why command rules are the Linux sensor for delete-based tampering: the
// fanotify feed (feed_linux.go) can only mark FAN_MODIFY | FAN_CLOSE_WRITE at
// mount level — the kernel rejects FAN_DELETE, FAN_DELETE_SELF, FAN_RENAME,
// FAN_CREATE and FAN_MOVED_TO with EINVAL. Deleting or renaming a backup tree
// is therefore invisible to the file feed, so these command-line rules (fired
// by the process watcher) are what catch recovery-tampering on Linux.
//
// Each rule maps to a real command an operator or attacker actually runs.
// XFS has no user-space "delete snapshot" CLI (snapshots are plain
// subvolumes), so the btrfs + lvm rules cover the practical snapshot-removal
// surface.
var t1490Rules = []t1490Rule{
	{"rm_backup_tree", []string{"rm", "-rf", "backup"}},
	{"shred_backup", []string{"shred", "backup"}},
	{"systemctl_stop_backup", []string{"systemctl", "stop", "backup"}},
	{"systemctl_mask_backup", []string{"systemctl", "mask", "backup"}},
	{"btrfs_subvolume_delete", []string{"btrfs", "subvolume", "delete"}},
	{"btrfs_snapshot_delete", []string{"btrfs", "snapshot", "delete"}},
	{"lvm_lvremove", []string{"lvremove", "-f"}},
	{"tune2fs_disable_journal", []string{"tune2fs", "^has_journal"}},
	{"e2fsck_destroy_journal", []string{"e2fsck", "-j", "/dev/null"}},
}
