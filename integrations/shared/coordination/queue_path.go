package coordination

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

type QueuePath struct {
	Scheduler Store
	Publisher JobPublisher
	Consumer  JobConsumer
	Cursors   CursorStore

	// Only non-nil on a successful setup.
	Close func()
}

type QueuePathBuild struct {
	Connect           func(ctx context.Context) (jetstream.JetStream, NATSCloser, error)
	NewSchedulerStore func(ctx context.Context, js jetstream.JetStream, clock Clock) (Store, error)
	NewJobPublisher   func(ctx context.Context, js jetstream.JetStream) (JobPublisher, error)
	NewJobConsumer    func(ctx context.Context, js jetstream.JetStream) (JobConsumer, error)
	NewCursorStore    func(ctx context.Context, js jetstream.JetStream) (CursorStore, error)
}

func DefaultQueuePathBuild(connName, plugin string) QueuePathBuild {
	return QueuePathBuild{
		Connect: func(ctx context.Context) (jetstream.JetStream, NATSCloser, error) {
			js, nc, err := ConnectJetStream(ctx, connName)
			if err != nil {
				// Untyped nil, not the nil *nats.Conn: a nil pointer boxed
				// in a non-nil interface passes `closer != nil`, then panics.
				return nil, nil, err
			}
			return js, nc, nil
		},
		NewSchedulerStore: func(ctx context.Context, js jetstream.JetStream, clock Clock) (Store, error) {
			return NewNATSSchedulerStore(ctx, js, clock)
		},
		NewJobPublisher: func(ctx context.Context, js jetstream.JetStream) (JobPublisher, error) {
			return NewNATSJobPublisher(ctx, js, plugin)
		},
		NewJobConsumer: func(ctx context.Context, js jetstream.JetStream) (JobConsumer, error) {
			return NewNATSJobConsumer(ctx, js, plugin)
		},
		NewCursorStore: func(ctx context.Context, js jetstream.JetStream) (CursorStore, error) {
			return NewNATSCursorStore(ctx, js)
		},
	}
}

// Atomic. Callers retry forever, so a connection left open on a late failure
// would leak one per attempt, and a partial QueuePath would skip the election.
func SetupQueuePath(ctx context.Context, b QueuePathBuild) (QueuePath, error) {
	js, nc, err := b.Connect(ctx)
	if err != nil {
		return QueuePath{}, fmt.Errorf("connecting to NATS JetStream: %w", err)
	}

	scheduler, err := b.NewSchedulerStore(ctx, js, RealClock{})
	if err != nil {
		nc.Close()
		return QueuePath{}, fmt.Errorf("creating the scheduler store: %w", err)
	}

	publisher, err := b.NewJobPublisher(ctx, js)
	if err != nil {
		nc.Close()
		return QueuePath{}, fmt.Errorf("creating the job publisher: %w", err)
	}

	consumer, err := b.NewJobConsumer(ctx, js)
	if err != nil {
		nc.Close()
		return QueuePath{}, fmt.Errorf("creating the job consumer: %w", err)
	}

	cursors, err := b.NewCursorStore(ctx, js)
	if err != nil {
		nc.Close()
		return QueuePath{}, fmt.Errorf("creating the cursor store: %w", err)
	}

	return QueuePath{
		Scheduler: scheduler,
		Publisher: publisher,
		Consumer:  consumer,
		Cursors:   cursors,
		Close:     nc.Close,
	}, nil
}

func DialQueuePath(connName, plugin string) func(ctx context.Context) (QueuePath, error) {
	build := DefaultQueuePathBuild(connName, plugin)
	return func(ctx context.Context) (QueuePath, error) {
		return SetupQueuePath(ctx, build)
	}
}
