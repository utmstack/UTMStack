package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

type statusRepo struct {
	connectors.ExecutionRepository
	err     error
	written []domain.ExecutionStatus
}

func (r *statusRepo) UpdateStatus(_ context.Context, _ uuid.UUID, u connectors.ExecutionStatusUpdate) error {
	if r.err != nil {
		return r.err
	}
	if u.Status != nil {
		r.written = append(r.written, *u.Status)
	}
	return nil
}

// A command sent while its row still reads PENDING can never be settled, so the
// lease runs out and it is sent again. If the row cannot be marked, nothing goes
// out.
func TestACommandIsNotSentWhenItsRowCannotBeMarkedExecuting(t *testing.T) {
	repo := &statusRepo{err: errors.New("connection reset")}
	d := &Dispatcher{exec: repo}

	if d.markExecuting(context.Background(), uuid.New()) {
		t.Fatal("the dispatcher would go on to send the command")
	}
}

func TestTheRowIsMarkedExecutingBeforeTheCommandGoesOut(t *testing.T) {
	repo := &statusRepo{}
	d := &Dispatcher{exec: repo}

	if !d.markExecuting(context.Background(), uuid.New()) {
		t.Fatal("refused a healthy write")
	}
	if len(repo.written) != 1 || repo.written[0] != domain.ExecutionStatusExecuting {
		t.Errorf("written = %v, want EXECUTING", repo.written)
	}
}
