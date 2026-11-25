package coordination

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

type NATSCloser interface {
	Close()
}

type LeasePath struct {
	Leases  Store
	Cursors CursorStore

	// Only non-nil on a successful setup.
	Close func()
}

type LeasePathBuild struct {
	Connect        func(ctx context.Context) (jetstream.JetStream, NATSCloser, error)
	NewLeaseStore  func(ctx context.Context, js jetstream.JetStream, clock Clock) (Store, error)
	NewCursorStore func(ctx context.Context, js jetstream.JetStream) (CursorStore, error)
}

func DefaultLeasePathBuild(connName string) LeasePathBuild {
	return LeasePathBuild{
		Connect: func(ctx context.Context) (jetstream.JetStream, NATSCloser, error) {
			js, nc, err := ConnectJetStream(ctx, connName)
			if err != nil {
				// Untyped nil, not the nil *nats.Conn: a nil pointer boxed in
				// a non-nil interface passes `closer != nil`, then panics.
				return nil, nil, err
			}
			return js, nc, nil
		},
		NewLeaseStore: func(ctx context.Context, js jetstream.JetStream, clock Clock) (Store, error) {
			return NewNATSStore(ctx, js, clock)
		},
		NewCursorStore: func(ctx context.Context, js jetstream.JetStream) (CursorStore, error) {
			return NewNATSCursorStore(ctx, js)
		},
	}
}

// Atomic. Callers retry forever, so a connection left open on a late failure
// would leak one per attempt, and a partial LeasePath would ingest unowned.
func SetupLeasePath(ctx context.Context, b LeasePathBuild) (LeasePath, error) {
	js, nc, err := b.Connect(ctx)
	if err != nil {
		return LeasePath{}, fmt.Errorf("connecting to NATS JetStream: %w", err)
	}

	leases, err := b.NewLeaseStore(ctx, js, RealClock{})
	if err != nil {
		nc.Close()
		return LeasePath{}, fmt.Errorf("creating the lease store: %w", err)
	}

	cursors, err := b.NewCursorStore(ctx, js)
	if err != nil {
		nc.Close()
		return LeasePath{}, fmt.Errorf("creating the cursor store: %w", err)
	}

	return LeasePath{Leases: leases, Cursors: cursors, Close: nc.Close}, nil
}

func DialLeasePath(connName string) func(ctx context.Context) (LeasePath, error) {
	build := DefaultLeasePathBuild(connName)
	return func(ctx context.Context) (LeasePath, error) {
		return SetupLeasePath(ctx, build)
	}
}
