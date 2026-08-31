package coordination

import (
	"context"
	"errors"
	"time"
)

// No error ends the loop: giving up would leave the unit ingested by nobody.
// ErrHeld never reaches onError; a contended key is the expected steady state.
func AcquireWithRetry(
	ctx context.Context, store Store,
	key, holder string,
	ttl, retryInterval time.Duration,
	onError func(error),
) (Lease, error) {
	for {
		if err := ctx.Err(); err != nil {
			return Lease{}, err
		}

		lease, err := store.Acquire(ctx, key, holder, ttl)
		if err == nil {
			return lease, nil
		}

		// A Store propagating ctx fails with the context's error, not ErrHeld.
		if onError != nil && !errors.Is(err, ErrHeld) && ctx.Err() == nil {
			onError(err)
		}

		select {
		case <-ctx.Done():
			// Zero Lease: nothing the caller could mistake for ownership.
			return Lease{}, ctx.Err()
		case <-time.After(retryInterval):
		}
	}
}
