//go:build linux

package ransomware

import (
	"context"
	"fmt"
	"strings"
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
		{unix.FAN_CREATE, OpCreate, true},
		{unix.FAN_MOVED_TO, OpRename, true},
		{unix.FAN_DELETE, OpDelete, true},
		// write bits take priority over a compound mask.
		{unix.FAN_CLOSE_WRITE | unix.FAN_MOVED_TO, OpWrite, true},
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

// TestEffectiveMaskDegradation: a kernel that accepts only the write bits
// yields exactly those bits; one that accepts all yields the full mask.
func TestEffectiveMaskDegradation(t *testing.T) {
	// Simulate a kernel that declines CREATE and MOVED_TO (EINVAL at mount
	// level) but accepts the write bits — the observed target behavior.
	writeOnly := effectiveMask(func(bit uint64) error {
		if bit == unix.FAN_CREATE || bit == unix.FAN_MOVED_TO {
			return fmt.Errorf("EINVAL")
		}
		return nil
	})
	wantWrite := uint64(unix.FAN_CLOSE_WRITE | unix.FAN_MODIFY)
	if writeOnly != wantWrite {
		t.Errorf("write-only kernel mask = 0x%x, want 0x%x", writeOnly, wantWrite)
	}

	// A fully-capable kernel accepts every desired bit.
	full := effectiveMask(func(uint64) error { return nil })
	if full != fullFeedMask {
		t.Errorf("full kernel mask = 0x%x, want 0x%x", full, fullFeedMask)
	}

	// A kernel that accepts nothing yields 0 → the caller skips the volume.
	none := effectiveMask(func(uint64) error { return fmt.Errorf("EINVAL") })
	if none != 0 {
		t.Errorf("no-accept kernel mask = 0x%x, want 0", none)
	}
}

// TestConsiderKeepsCreate: a create op is a mutating op and must survive the
// pre-filter (it is a churn/ransom-note input), so a create on an in-scope
// path is kept while a create under an excluded path is dropped.
func TestConsiderKeepsCreate(t *testing.T) {
	if !Consider(FileEvent{PID: 1, Path: "/data/new", Op: OpCreate}, nil, nil, 999) {
		t.Error("create on in-scope path should be kept")
	}
	if Consider(FileEvent{PID: 1, Path: "/excluded/new", Op: OpCreate}, nil, func(p string) bool { return strings.Contains(p, "excluded") }, 999) {
		t.Error("create under an excluded path should be dropped")
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
