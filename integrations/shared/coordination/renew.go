package coordination

import (
	"context"
	"time"
)

// The only sanctioned lease-path cursor write: Renew first, Save only if it
// succeeded, so a holder whose lease was taken over cannot reach Save.
func RenewAndPersistCursor(
	ctx context.Context, store Store, cursors CursorStore,
	l Lease, ttl time.Duration,
	cursorKey string, cursorData []byte, cursorRev uint64,
) (Lease, Cursor, error) {
	renewed, err := store.Renew(ctx, l, ttl)
	if err != nil {
		// ErrFenced short-circuits here, before Save.
		return Lease{}, Cursor{}, err
	}

	saved, err := cursors.Save(ctx, cursorKey, Cursor{Data: cursorData, Revision: cursorRev})
	if err != nil {
		// The returned lease is still held; only this snapshot was lost.
		return renewed, Cursor{}, err
	}

	return renewed, saved, nil
}
