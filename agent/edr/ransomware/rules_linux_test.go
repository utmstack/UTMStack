//go:build linux

package ransomware

import "testing"

func TestMatchT1490Linux_Positives(t *testing.T) {
	cases := []struct {
		cmd, want string
	}{
		{`btrfs subvolume delete /data/snap1`, "btrfs_subvolume_delete"},
		{`systemctl stop restic-backup`, "systemctl_stop_backup"},
		{`rm -rf /var/backups`, "rm_backup_tree"},
		{`tune2fs -O ^has_journal /dev/sda1`, "tune2fs_disable_journal"},
		{`shred /backups/db.dump`, "shred_backup"},
		{`systemctl mask restic-backup.service`, "systemctl_mask_backup"},
		{`lvremove -f /dev/vg0/snap1`, "lvm_lvremove"},
		{`e2fsck -f -j /dev/null /dev/sda1`, "e2fsck_destroy_journal"},
	}
	for _, c := range cases {
		name, ok := MatchT1490("/usr/bin/"+firstWord(c.cmd), c.cmd, nil)
		if !ok || name != c.want {
			t.Errorf("MatchT1490(%q) = %q,%v ; want %q,true", c.cmd, name, ok, c.want)
		}
	}
}

func TestMatchT1490Linux_Negatives(t *testing.T) {
	cases := []string{
		`rm -rf /tmp/build`,
		`systemctl stop sshd`,
		`btrfs subvolume list /`,
	}
	for _, cmd := range cases {
		if _, ok := MatchT1490("/usr/bin/"+firstWord(cmd), cmd, nil); ok {
			t.Errorf("MatchT1490(%q) matched ; want no match", cmd)
		}
	}
}

func TestMatchT1490Linux_Allowlist(t *testing.T) {
	// A distinctive token from sanctioned tooling suppresses the rule.
	if _, ok := MatchT1490("/usr/bin/systemctl", `systemctl stop restic-backup`, []string{"restic"}); ok {
		t.Error("command containing an allowlisted token should be exempt")
	}
	// A non-matching allowlist entry must NOT suppress a real T1490 command.
	if name, ok := MatchT1490("/usr/bin/systemctl", `systemctl stop restic-backup`, []string{"acronis"}); !ok || name != "systemctl_stop_backup" {
		t.Errorf("unrelated allowlist must not exempt; got %q,%v", name, ok)
	}
}

func firstWord(s string) string {
	for i, r := range s {
		if r == ' ' {
			return s[:i]
		}
	}
	return s
}
