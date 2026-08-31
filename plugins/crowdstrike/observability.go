package main

import (
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/shared/coordination"
)

// Every INFO line this plugin emits must be tied to a state transition bounded
// per unit: nothing per event, nothing on a fixed timer, or a real failure
// becomes indistinguishable from normal operation.
const (
	coordinationReadyMsg = "coordination ready, owning units by lease"
	leaseAcquiredMsg     = "unit lease acquired, this worker owns the unit"
	startingPositionMsg  = "starting position established"
	streamOpenedMsg      = "event stream opened"
	firstEventMsg        = "first event ingested"
)

// logCoordinationReady announces that the lease path came up. holder is the only
// value mapping a key held on the server back to a running worker.
func logCoordinationReady(holder string) {
	catcher.Info(coordinationReadyMsg, map[string]any{
		"process":      processName,
		"holder":       holder,
		"leasePrefix":  leaseKeyPrefix,
		"leaseBucket":  coordination.LeaseBucketName,
		"cursorBucket": coordination.CursorBucketName,
	})
}

// logLeaseAcquired announces that this worker now owns a unit. Call once per
// acquisition, never per renewal: the heartbeat renews every few seconds.
func logLeaseAcquired(key, holder string) {
	catcher.Info(leaseAcquiredMsg, map[string]any{
		"process": processName,
		"key":     key,
		"holder":  holder,
	})
}

// logStartingPosition takes startsAfter in Unix milliseconds. Call it after the
// zero-floor guard: the line must not claim a position for a unit about to
// refuse to open.
func logStartingPosition(key string, startsAfter uint64) {
	catcher.Info(startingPositionMsg, map[string]any{
		"process":     processName,
		"key":         key,
		"startsAfter": time.UnixMilli(int64(startsAfter)).UTC().Format(time.RFC3339Nano),
	})
}

// streamID is the feed's DataFeedURL, safe to log: CrowdStrike carries the
// session token in a separate field of the stream resource, so the URL holds no
// credential.
func logStreamOpened(groupKey, streamID string, offset uint64) {
	catcher.Info(streamOpenedMsg, map[string]any{
		"process":  processName,
		"key":      groupKey,
		"streamID": streamID,
		"offset":   offset,
	})
}

// firstEventGate lets one event per stream session announce itself. No mutex: it
// must stay on the stack of the single goroutine running processStreamEvents.
// Keep it per session, not process-wide, or reconnects go unreported.
type firstEventGate struct {
	announced bool
}

func (g *firstEventGate) take() bool {
	if g.announced {
		return false
	}
	g.announced = true
	return true
}

// Call only for an event that passed the floor filter and was enqueued, so the
// line reports ingestion rather than mere delivery.
func logFirstEventIngested(groupKey, streamID string, offset uint64) {
	catcher.Info(firstEventMsg, map[string]any{
		"process":  processName,
		"key":      groupKey,
		"streamID": streamID,
		"offset":   offset,
	})
}
