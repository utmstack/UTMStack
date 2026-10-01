//go:build linux

package ransomware

import (
	"context"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestOpForMask(t *testing.T) {
	cases := []struct {
		mask uint64
		op   FileOp
		ok   bool
	}{
		{unix.FAN_CLOSE_WRITE, OpWrite, true},
		{unix.FAN_MODIFY, OpWrite, true},
		{unix.FAN_CLOSE_WRITE | unix.FAN_MODIFY, OpWrite, true},
		{0, 0, false},
		{unix.FAN_Q_OVERFLOW, 0, false},
	}
	for _, c := range cases {
		op, ok := opForMask(c.mask)
		if ok != c.ok || (ok && op != c.op) {
			t.Errorf("opForMask(0x%x) = %v,%v ; want %v,%v", c.mask, op, ok, c.op, c.ok)
		}
	}
}

func TestNewFeedNoVolumesBlocksUntilCancel(t *testing.T) {
	feed := NewFeed()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := feed.Run(ctx, func(FileEvent) {})
	if err != context.DeadlineExceeded {
		t.Fatalf("Run with no volumes = %v ; want context.DeadlineExceeded", err)
	}
}
