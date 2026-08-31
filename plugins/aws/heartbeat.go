package main

import (
	"context"
	"errors"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/shared/coordination"
)

// cancel must be this group's own CancelFunc: fencing uses it to stop the
// group, so a wider one would take down unrelated groups. Nil is allowed
// only when the caller drives Tick itself.
func groupHeartbeat(coord coordination.LeasePath, cursorKey string, cursors *cursorMap, cancel context.CancelFunc) coordination.Heartbeat {
	return coordination.Heartbeat{
		Leases:   coord.Leases,
		Cursors:  coord.Cursors,
		Key:      cursorKey,
		Snapshot: cursors.marshalSnapshot,
		TTL:      coordination.LeaseTTL,
		Interval: coordination.HeartbeatInterval,
		Cancel:   cancel,
		OnError:  reportHeartbeatFailure(cursorKey),
	}
}

func reportHeartbeatFailure(cursorKey string) func(error) {
	return func(err error) {
		switch {
		case errors.Is(err, coordination.ErrFenced):
			catcher.Info("lease fenced, stopping group", map[string]any{
				"process": processName,
				"key":     cursorKey,
			})

		case errors.Is(err, coordination.ErrCursorTooLarge):
			_ = catcher.Error("cursor snapshot exceeds the maximum persistable size, will NOT resolve by retrying — this group's stream count is too large for a single cursor entry", err, map[string]any{
				"process": processName,
				"key":     cursorKey,
			})

		default:
			_ = catcher.Error("failed to persist cursor snapshot, will retry next heartbeat", err, map[string]any{
				"process": processName,
				"key":     cursorKey,
			})
		}
	}
}
