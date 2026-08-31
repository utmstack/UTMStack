package main

import (
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/shared/coordination"
)

// leaseKeyPrefix namespaces this plugin's keys inside the shared lease bucket.
// The rest of the key is ModuleGroup.Key() passed through unmodified.
const leaseKeyPrefix = "crowdstrike."

// runOwnedStream blocks until this worker owns the unit's lease, then calls run
// synchronously. The caller must already be in the unit's own goroutine.
func runOwnedStream(coord coordination.LeasePath, holder string, stream *activeStream, run func(*activeStream)) {
	key := leaseKeyPrefix + stream.groupKey

	lease, err := coordination.AcquireWithRetry(
		stream.ctx, coord.Leases, key, holder,
		coordination.LeaseTTL, coordination.AcquireRetryInterval,
		func(err error) {
			_ = catcher.Error("failed to acquire the unit's lease", err, map[string]any{
				"process": processName,
				"key":     key,
			})
		},
	)
	if err != nil {
		// Stopped before it was ever owned; no socket to unwind.
		return
	}

	logLeaseAcquired(key, holder)

	cursorRev, err := activateCursor(stream.ctx, coord.Cursors, key, stream.cursor, time.Now())
	if err != nil {
		_ = catcher.Error("failed to establish the unit's starting position", err, map[string]any{
			"process": processName,
			"key":     key,
		})
	}

	// No floor means no usable position; opening the feed would emit everything
	// CrowdStrike still retains. Released, not abandoned, so another worker can
	// retry without waiting out the TTL.
	if stream.cursor.startsAfter() == 0 {
		_ = catcher.Error("refusing to open the event stream without a starting position", err, map[string]any{
			"process": processName,
			"key":     key,
		})
		_ = coord.Leases.Release(stream.ctx, lease)
		stream.cancel()
		return
	}

	logStartingPosition(key, stream.cursor.startsAfter())

	go groupHeartbeat(coord, key, stream.cursor, stream.cancel).Run(stream.ctx, lease, cursorRev)

	run(stream)
}
