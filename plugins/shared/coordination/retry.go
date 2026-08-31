package coordination

import (
	"context"
	"time"
)

// Must never fall back to proceeding without a successful dial.
// nats.MaxReconnects(-1) only protects an already-established connection, and
// `docker stack deploy` ignores depends_on, so NATS may be down at startup.
func SetupWithRetry[T any](ctx context.Context, dial func(ctx context.Context) (T, error), retryDelay time.Duration, onFailure func(err error)) (T, error) {
	for {
		v, err := dial(ctx)
		if err == nil {
			return v, nil
		}

		if onFailure != nil {
			onFailure(err)
		}

		select {
		case <-ctx.Done():
			var zero T
			return zero, ctx.Err()
		case <-time.After(retryDelay):
		}
	}
}
