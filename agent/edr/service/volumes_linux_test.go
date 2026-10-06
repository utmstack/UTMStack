//go:build linux

package service

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLocalMountsParser(t *testing.T) {
	fixture := `udev /dev devtmpfs rw,nosuid 0 0
none /sys sysfs rw,nosuid 0 0
none /proc proc rw,nosuid 0 0
tmpfs /run tmpfs rw,nosuid,mode=755 0 0
/dev/sda2 / ext4 rw,relatime,errors=remount-ro 0 1
/dev/sda1 /home ext4 rw,relatime 0 2
/dev/sda3 /var xfs rw,relatime 0 2
none /var/snap squashfs ro,nosuid,nodev,relatime,linktime=... 0 0
overlay /var/lib/docker/overlay2/abc/merged overlay rw,nosuid 0 0
/dev/sr0 /media/cdrom iso9660 ro,noexec 0 0
host /mnt/nfs nfs4 rw,vers=4 0 0
//? /mnt/cifs cifs rw 0 0
/dev/loop0 /mnt/fusefuse fuse.myfs rw 0 0
/dev/sdb1 /mnt/usb vfat rw,relatime 0 0
`
	dir := t.TempDir()
	table := filepath.Join(dir, "mounts")
	if err := os.WriteFile(table, []byte(fixture), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := localMounts(table)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"/", "/home", "/var", "/mnt/usb"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("localMounts = %v, want %v", got, want)
	}
}

// TestLocalMountsDedupByDevice: two mount points of the same disk are
// reported once (first mount point wins), so the watcher does not mark the
// same device twice.
func TestLocalMountsDedupByDevice(t *testing.T) {
	dir := t.TempDir()
	table := filepath.Join(dir, "mounts")
	if err := os.WriteFile(table, []byte("/dev/sda2 / ext4 rw 0 1\n/dev/sda2 /home ext4 rw 0 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := localMounts(table)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"/"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("localMounts = %v, want %v (same-device mounts dedup)", got, want)
	}
}

func TestLocalMountsMissing(t *testing.T) {
	if _, err := localMounts("/nonexistent/mounts"); err == nil {
		t.Fatal("expected an error for a missing mount table")
	}
}

func TestIsLocalType(t *testing.T) {
	yes := []string{"ext2", "ext3", "ext4", "xfs", "btrfs", "zfs", "vfat", "ntfs", "exfat", "f2fs"}
	no := []string{"proc", "sysfs", "tmpfs", "devtmpfs", "squashfs", "overlay", "iso9660", "nfs", "nfs4", "cifs", "fuse.sshfs"}
	for _, fs := range yes {
		if !isLocalType(fs) {
			t.Errorf("isLocalType(%q) = false, want true", fs)
		}
	}
	for _, fs := range no {
		if isLocalType(fs) {
			t.Errorf("isLocalType(%q) = true, want false", fs)
		}
	}
}
