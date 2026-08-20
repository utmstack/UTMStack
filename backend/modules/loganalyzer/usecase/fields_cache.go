package usecase

import (
	"context"
	"slices"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/utmstack/utmstack/backend/modules/loganalyzer/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

// fieldsTTL is how long a discovered field list is trusted.
//
// The list cannot be read from the table definition, because most of it is not
// there: the JSON columns carry whatever the parsers produced, so the only way
// to know which paths exist is to read them out of every row the tenant holds.
// Measured against 2.9M rows that was 25 seconds and 9.66 GiB, and the explorer
// asks for it on every page load.
//
// What the list answers is which fields an analyst can filter on, and that
// changes only when a parser learns a new one. Minutes of staleness cost them
// nothing; paying 25 seconds to re-derive the same answer costs them the page.
const fieldsTTL = 5 * time.Minute

type fieldsEntry struct {
	fields []dto.Field
	loaded time.Time
}

// fieldsCache keys on tenant as well as dataset. The field list is derived from
// one tenant's rows and describes only those, so a key without it would show an
// analyst the shape of someone else's logs.
type fieldsCache struct {
	mu    sync.RWMutex
	items map[string]*fieldsEntry
	group singleflight.Group
}

func newFieldsCache() *fieldsCache {
	return &fieldsCache{items: make(map[string]*fieldsEntry)}
}

func (c *fieldsCache) get(ctx context.Context, dataset string, load func(context.Context) ([]dto.Field, error)) ([]dto.Field, error) {
	key := authz.TenantIDFromContext(ctx) + "\x00" + dataset

	c.mu.RLock()
	entry := c.items[key]
	c.mu.RUnlock()

	if entry != nil {
		if time.Since(entry.loaded) < fieldsTTL {
			return slices.Clone(entry.fields), nil
		}
		// Stale beats absent here: an expired list is still a true description
		// of the fields that existed minutes ago, so it goes back immediately
		// and the refresh happens behind the answer. Detached from the request
		// context because that one is cancelled as soon as this returns, which
		// would abort the very read we are trying to stop repeating.
		refresh := context.WithoutCancel(ctx)
		go func() {
			_, _, _ = c.group.Do(key, func() (any, error) {
				return c.reload(refresh, key, load)
			})
		}()
		return slices.Clone(entry.fields), nil
	}

	// Nothing cached, so this one has to wait. singleflight is what keeps the
	// explorer's several openings at once from each starting their own read.
	v, err, _ := c.group.Do(key, func() (any, error) {
		return c.reload(ctx, key, load)
	})
	if err != nil {
		return nil, err
	}
	return slices.Clone(v.([]dto.Field)), nil
}

func (c *fieldsCache) reload(ctx context.Context, key string, load func(context.Context) ([]dto.Field, error)) ([]dto.Field, error) {
	fields, err := load(ctx)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.items[key] = &fieldsEntry{fields: fields, loaded: time.Now()}
	c.mu.Unlock()
	return fields, nil
}
