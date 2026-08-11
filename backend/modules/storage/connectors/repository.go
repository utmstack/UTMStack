package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/storage/domain"
)

type StoreRepository interface {
	Retention(ctx context.Context, d domain.Dataset) (domain.Retention, error)
	SetRetention(ctx context.Context, r domain.Retention) error

	// AdoptTiering moves a dataset onto the tiered policy and sets its
	// retention. The two cannot be one statement: a TTL may only name a volume
	// the table's policy already has.
	AdoptTiering(ctx context.Context, r domain.Retention) error

	Usage(ctx context.Context) ([]domain.Usage, error)
	Health(ctx context.Context) (domain.Health, error)

	// ColdStorageReady reports whether the server currently offers the cold
	// volume. It is about the server's own configuration, not about any table.
	ColdStorageReady(ctx context.Context) (bool, error)

	// ReloadConfig makes the server read its configuration again, which is what
	// turns a written file into a disk it can use.
	ReloadConfig(ctx context.Context) error
}

// ConfigRepository owns the file that declares cold storage to the event store.
// The secret it writes is never read back out: the only reader that needs it is
// the server itself.
type ConfigRepository interface {
	Read() (domain.Tiering, error)

	// Write returns how to put the file back as it was. The store validates a
	// bucket when it reads the file, so a rejected one has to be taken out
	// again — left behind, it is what the server would load on its next start.
	Write(o domain.ObjectStore) (undo func() error, err error)
}
