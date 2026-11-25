package coordination

import (
	"context"
	"errors"
)

type Cursor struct {
	Data []byte
	// CAS token in CursorStore, unrelated to Lease.Revision.
	Revision uint64
}

// Callers must treat this as "seed from now", never "seed from epoch".
var ErrCursorNotFound = errors.New("cursor: key has no persisted value yet")

var ErrCursorConflict = errors.New("cursor: revision stale, another writer saved since Load")

// Save exceeded the transport's max message payload: 1,048,576 bytes on the
// deployed nats:2.10-alpine v2.10.29 server. It will fail identically on retry.
var ErrCursorTooLarge = errors.New("cursor: payload exceeds the transport's maximum message size")

type CursorStore interface {
	Load(ctx context.Context, key string) (Cursor, error)

	// CAS write. c.Revision == 0 means "create" and fails if key exists; any
	// other value means "update only if still current".
	Save(ctx context.Context, key string, c Cursor) (Cursor, error)
}
