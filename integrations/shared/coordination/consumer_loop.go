package coordination

import (
	"context"
	"errors"
)

// No error ends the loop: one group's transient failure must not stop
// ingestion for the rest. ErrNoJobs and cancellation never reach onError.
func RunJobConsumer(
	ctx context.Context,
	consumer JobConsumer,
	handle func(ctx context.Context, delivery JobDelivery) error,
	onError func(err error),
) {
	for {
		// Checked here and again below before ErrNoJobs: Fetch reports a
		// cancelled context as ErrNoJobs, so testing that first would spin.
		if ctx.Err() != nil {
			return
		}

		delivery, err := consumer.Fetch(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if errors.Is(err, ErrNoJobs) {
				continue
			}
			if onError != nil {
				onError(err)
			}
			continue
		}

		if err := handle(ctx, delivery); err != nil && onError != nil {
			onError(err)
		}
	}
}
