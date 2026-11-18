//go:build linux

package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestLinuxResolverReadsBoundedRegularFilesAndRejectsFIFO(t *testing.T) {
	dir := t.TempDir()
	regular := filepath.Join(dir, "regular")
	if err := os.WriteFile(regular, []byte("nameserver 192.0.2.53\n"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(regular, link); err != nil {
		t.Fatal(err)
	}
	if data, err := readResolverFile(link); err != nil || string(data) != "nameserver 192.0.2.53\n" {
		t.Fatal("regular resolver symlink changed", string(data), err)
	}
	if _, err := readResolverFile(dir); err == nil {
		t.Fatal("resolver directory accepted")
	}
	if err := os.WriteFile(regular, make([]byte, (64<<10)+1), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readResolverFile(regular); err == nil {
		t.Fatal("oversized resolver file accepted")
	}
	fifo := filepath.Join(dir, "fifo")
	if err := unix.Mkfifo(fifo, 0600); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := readResolverFile(fifo); done <- err }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("resolver FIFO accepted")
		}
	case <-time.After(time.Second):
		writer, err := os.OpenFile(fifo, os.O_WRONLY|unix.O_NONBLOCK, 0)
		if err == nil {
			writer.Close()
		}
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("resolver reader did not release after FIFO control")
		}
		t.Fatal("resolver FIFO parked a service refresh")
	}
}
