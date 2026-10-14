package coordination

import (
	"context"
	"errors"
)

type ConsumeWork struct {
	Resume func(ctx context.Context, job Job, cur Cursor) (next []byte, err error)

	// Called instead of Resume on first activation. The position it returns
	// must be seeded from "now", or the source's history is replayed whole.
	Activate func(ctx context.Context, job Job) (next []byte, err error)
}

var ErrMissingActivateWork = errors.New("coordination: ConsumeAndAdvanceCursor: first activation occurred but ConsumeWork.Activate is nil")

var ErrMissingResumeWork = errors.New("coordination: ConsumeAndAdvanceCursor: a cursor already exists for this key but ConsumeWork.Resume is nil")

// Failures without a cursor are neither Saved nor Acked, so AckWait drives
// redelivery; a failure with a returned cursor (partial progress) is Saved and
// Acked so one slow source cannot hammer the same window. On ErrCursorConflict
// the write is dropped but the job is Acked anyway: the work already published.
func ConsumeAndAdvanceCursor(
	ctx context.Context,
	cursors CursorStore,
	delivery JobDelivery,
	cursorKey string,
	work ConsumeWork,
) error {
	cur, loadErr := cursors.Load(ctx, cursorKey)

	var next []byte
	var workErr error

	switch {
	case loadErr == nil:
		if work.Resume == nil {
			return ErrMissingResumeWork
		}
		next, workErr = work.Resume(ctx, delivery.Job, cur)

	case errors.Is(loadErr, ErrCursorNotFound):
		if work.Activate == nil {
			return ErrMissingActivateWork
		}
		next, workErr = work.Activate(ctx, delivery.Job)

	default:
		// Load failed for a reason other than "no cursor yet".
		return loadErr
	}

	if workErr != nil {
		// A returned cursor carries partial progress: persist it and Ack.
		// Nil next keeps the legacy redelivery.
		if next != nil {
			if _, saveErr := cursors.Save(ctx, cursorKey, Cursor{Data: next, Revision: cur.Revision}); saveErr == nil || errors.Is(saveErr, ErrCursorConflict) {
				return delivery.Ack()
			}
		}
		return workErr
	}

	if _, saveErr := cursors.Save(ctx, cursorKey, Cursor{Data: next, Revision: cur.Revision}); saveErr != nil {
		if errors.Is(saveErr, ErrCursorConflict) {
			// CAS loser: discard this write, but still Ack.
			return delivery.Ack()
		}
		// Genuine Save failure: do not Ack, let AckWait retry.
		return saveErr
	}

	return delivery.Ack()
}
