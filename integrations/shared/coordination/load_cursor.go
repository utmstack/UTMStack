package coordination

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// An absent cursor returns found=false, no error, dst untouched: callers must
// treat that as "seed from now". A decode failure still returns the stored
// revision, or the next Save means "create" on a live key and fails forever.
func LoadCursorInto[T any](ctx context.Context, cursors CursorStore, key string, dst *T) (uint64, bool, error) {
	cur, err := cursors.Load(ctx, key)
	if err != nil {
		if errors.Is(err, ErrCursorNotFound) {
			return 0, false, nil
		}
		return 0, false, err
	}

	if err := json.Unmarshal(cur.Data, dst); err != nil {
		return cur.Revision, false, fmt.Errorf("decoding the persisted cursor for %q: %w", key, err)
	}

	return cur.Revision, true, nil
}
