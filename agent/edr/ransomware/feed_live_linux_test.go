//go:build linux

package ransomware

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestFanotifyFeedLive is a real kernel integration test. It requires root
// (CAP_SYS_ADMIN for fanotify_init), so it skips cleanly when unavailable.
// Self-contained: starts the feed over a temp dir, writes a file from the
// test process, and asserts the sink sees the event with the test's own PID.
// Run with sudo to exercise it: sudo go test -run TestFanotifyFeedLive ./edr/ransomware
func TestFanotifyFeedLive(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("fanotify requires root (CAP_SYS_ADMIN); run with sudo")
	}
	dir := t.TempDir()

	events := make(chan FileEvent, 16)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	feed := NewFeed(dir)
	done := make(chan error, 1)
	go func() { done <- feed.Run(ctx, func(ev FileEvent) { events <- ev }) }()

	// Give the feed a moment to init and mark before writing.
	time.Sleep(200 * time.Millisecond)

	target := filepath.Join(dir, "live_probe.txt")
	if err := os.WriteFile(target, []byte("x"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Read events until we find one for our file or the deadline fires.
	deadline := time.After(2 * time.Second)
found:
	for {
		select {
		case ev := <-events:
			if !strings.HasPrefix(ev.Path, dir) {
				continue
			}
			if ev.PID != os.Getpid() {
				t.Fatalf("event PID = %d ; want %d", ev.PID, os.Getpid())
			}
			if ev.Op != OpWrite {
				t.Fatalf("event Op = %v ; want OpWrite", ev.Op)
			}
			t.Logf("live event: pid=%d path=%s op=%v", ev.PID, ev.Path, ev.Op)
			break found
		case <-deadline:
			t.Fatal("no file event for the probe write within 2s")
		}
	}

	cancel()
	select {
	case err := <-done:
		if err != context.Canceled {
			t.Fatalf("Run after cancel = %v ; want context.Canceled", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not return within 3s of ctx cancel")
	}
}
