package main

import (
	"context"
	"errors"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/shared/coordination"
)

// groupHeartbeat builds the heartbeat for one unit. cancel must cancel the
// context processStreamEvents is parked on: fencing relies on it to cut the open
// Falcon socket, or this worker keeps reading a feed another worker now owns.
func groupHeartbeat(coord coordination.LeasePath, cursorKey string, cursor *cursorState, cancel context.CancelFunc) coordination.Heartbeat {
	return coordination.Heartbeat{
		Leases:   coord.Leases,
		Cursors:  coord.Cursors,
		Key:      cursorKey,
		Snapshot: cursor.marshalSnapshot,
		TTL:      coordination.LeaseTTL,
		Interval: coordination.HeartbeatInterval,
		Cancel:   cancel,
		OnError:  reportHeartbeatFailure(cursorKey),
	}
}

func reportHeartbeatFailure(cursorKey string) func(error) {
	return func(err error) {
		if errors.Is(err, coordination.ErrFenced) {
			catcher.Info("lease fenced, cutting this unit's event stream", map[string]any{
				"process": processName,
				"key":     cursorKey,
			})
			return
		}

		_ = catcher.Error("failed to persist the offsets snapshot, will retry next heartbeat", err, map[string]any{
			"process": processName,
			"key":     cursorKey,
		})
	}
}
