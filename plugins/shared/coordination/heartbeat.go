package coordination

import (
	"context"
	"errors"
	"time"
)

// Must stay defined as bucketTTL: that is the TTL the server enforces.
const LeaseTTL = bucketTTL

const HeartbeatInterval = LeaseTTL / 3

const AcquireRetryInterval = HeartbeatInterval

// Must read the caller's live position; a captured payload never advances.
type CursorSnapshot func() ([]byte, error)

type Heartbeat struct {
	Leases  Store
	Cursors CursorStore

	// One key for both lease and cursor, so a successor finds the position.
	Key string

	Snapshot CursorSnapshot

	TTL time.Duration

	Interval time.Duration

	// Must be the same CancelFunc that stops the unit for any other reason:
	// refusing new work leaves a feed socket held open for hours.
	Cancel context.CancelFunc

	OnError func(error)
}

// The caller carries the returned lease and revision into the next tick. On
// ErrFenced both are zero and the unit must stop; otherwise the lease is held.
func (h Heartbeat) Tick(ctx context.Context, lease Lease, cursorRev uint64) (Lease, uint64, error) {
	data, err := h.Snapshot()
	if err != nil {
		// Lease returned unchanged: a snapshot failure is a payload bug.
		return lease, cursorRev, err
	}

	renewed, saved, err := RenewAndPersistCursor(ctx, h.Leases, h.Cursors, lease, h.TTL, h.Key, data, cursorRev)
	if err != nil {
		if errors.Is(err, ErrFenced) {
			// Verbatim: wrapping disables the only signal that stops the unit.
			return Lease{}, 0, err
		}
		return renewed, cursorRev, err
	}

	return renewed, saved.Revision, nil
}

func (h Heartbeat) Run(ctx context.Context, lease Lease, cursorRev uint64) {
	ticker := time.NewTicker(h.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		nextLease, nextRev, err := h.Tick(ctx, lease, cursorRev)
		if err != nil {
			if h.OnError != nil {
				h.OnError(err)
			}
			if errors.Is(err, ErrFenced) {
				if h.Cancel != nil {
					h.Cancel()
				}
				return
			}
		}

		lease = nextLease
		cursorRev = nextRev
	}
}
