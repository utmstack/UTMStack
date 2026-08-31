// Package coordination distributes ingestion work across worker processes.
package coordination

import (
	"context"
	"errors"
	"time"
)

type Clock interface {
	Now() time.Time
}

// No position field: this bucket's TTL deletes the key, and failover happens by
// that expiry, so a position here would die exactly when a successor needed it.
type Lease struct {
	// Opaque and precomputed; Store implementations never parse it.
	Key string
	// The worker process that owns this lease.
	Holder string
	// Fencing token: increases on every Acquire or Renew, so a stale Lease
	// loses any later CAS.
	Revision uint64
	// When this lease becomes eligible for another Acquire.
	ExpiresAt time.Time
}

var ErrHeld = errors.New("lease already held by a live owner")

var ErrFenced = errors.New("lease fenced: revision stale, another holder took over")

// Ownership only. The lease path writes cursors via RenewAndPersistCursor.
type Store interface {
	Acquire(ctx context.Context, key, holder string, ttl time.Duration) (Lease, error)

	Renew(ctx context.Context, l Lease, ttl time.Duration) (Lease, error)

	Release(ctx context.Context, l Lease) error
}
