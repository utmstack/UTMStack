package coordination

import (
	"context"
	"errors"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	// CursorBucketName is exported so plugins can name the bucket in their logs
	// without keeping a copy of the string.
	CursorBucketName = "ingest_cursors"

	cursorBucketHistory = 1
)

// Backed by a bucket with no TTL. The lease bucket expires by deleting its
// keys, and position must survive exactly that moment.
type NATSCursorStore struct {
	kv jetstream.KeyValue
}

func NewNATSCursorStore(ctx context.Context, js jetstream.JetStream) (*NATSCursorStore, error) {
	kv, err := js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:  CursorBucketName,
		History: cursorBucketHistory,
		// TTL omitted: keys in this bucket must never expire.
	})
	if err != nil {
		return nil, err
	}

	return &NATSCursorStore{kv: kv}, nil
}

func (s *NATSCursorStore) Load(ctx context.Context, key string) (Cursor, error) {
	entry, err := s.kv.Get(ctx, key)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return Cursor{}, ErrCursorNotFound
		}
		return Cursor{}, err
	}

	return Cursor{Data: entry.Value(), Revision: entry.Revision()}, nil
}

func (s *NATSCursorStore) Save(ctx context.Context, key string, c Cursor) (Cursor, error) {
	if c.Revision == 0 {
		revision, err := s.kv.Create(ctx, key, c.Data)
		if err != nil {
			if errors.Is(err, jetstream.ErrKeyExists) {
				return Cursor{}, ErrCursorConflict
			}
			return Cursor{}, translateSaveErr(err)
		}
		return Cursor{Data: c.Data, Revision: revision}, nil
	}

	revision, err := s.kv.Update(ctx, key, c.Data, c.Revision)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyRevisionMismatch) {
			return Cursor{}, ErrCursorConflict
		}
		return Cursor{}, translateSaveErr(err)
	}
	return Cursor{Data: c.Data, Revision: revision}, nil
}

func translateSaveErr(err error) error {
	if errors.Is(err, nats.ErrMaxPayload) {
		return ErrCursorTooLarge
	}
	return err
}
