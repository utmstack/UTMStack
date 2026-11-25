package eventstore

import (
	"context"
	"sync"
	"time"

	"github.com/utmstack/go-sdk/v2/store"
	"golang.org/x/sync/singleflight"
)

const (
	fieldsTTL            = 10 * time.Minute
	fieldsComputeTimeout = 3 * time.Minute
)

type fieldsEntry struct {
	fields []store.Field
	at     time.Time
}

type fieldsCache struct {
	mu    sync.Mutex
	items map[string]fieldsEntry
	group singleflight.Group
}

// DescribeFields is the store's own, with a cache in front. Listing the fields of a dataset reads the JSON
// columns of the table to find the paths inside them, which on a big table takes longer than the callers
// wait, and the answer hardly changes. So it is computed once for everyone asking at the same moment, kept
// for ten minutes, and an old copy is served at once while a fresh one is worked out. A caller that stops
// waiting does not stop the work: the next one finds it done.
func (s *Store) DescribeFields(ctx context.Context, sc store.Scope) ([]store.Field, error) {
	key := string(sc.Dataset) + "|" + sc.Tenant + "|" + sc.DataType

	s.fields.mu.Lock()
	entry, have := s.fields.items[key]
	s.fields.mu.Unlock()

	fresh := have && time.Since(entry.at) < fieldsTTL
	if fresh {
		return entry.fields, nil
	}

	ch := s.fields.group.DoChan(key, func() (any, error) {
		work, cancel := context.WithTimeout(context.WithoutCancel(ctx), fieldsComputeTimeout)
		defer cancel()

		fields, err := s.Store.DescribeFields(work, sc)
		if err != nil {
			return nil, err
		}

		s.fields.mu.Lock()
		if s.fields.items == nil {
			s.fields.items = map[string]fieldsEntry{}
		}

		s.fields.items[key] = fieldsEntry{fields: fields, at: time.Now()}
		s.fields.mu.Unlock()

		return fields, nil
	})

	if have {
		return entry.fields, nil // stale, but right enough: the refresh above is on its way
	}

	select {
	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}

		return res.Val.([]store.Field), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
