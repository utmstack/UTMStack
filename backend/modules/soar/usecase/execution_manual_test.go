package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

type createSpyRepo struct {
	connectors.ExecutionRepository
	created *domain.SoarExecution
}

func (r *createSpyRepo) Create(_ context.Context, e *domain.SoarExecution) (*domain.SoarExecution, error) {
	e.ID = uuid.New()
	r.created = e
	return e, nil
}

// The console runs a manual command itself and closes its record when it ends.
// The dispatcher claims every PENDING row and re-runs it through the shell
// executor, which looks the agent up by hostname: an agent id never matches, so
// a PENDING record was marked FAILED a few seconds in, whatever the console did.
func TestManualExecutionIsNotLeftForTheDispatcher(t *testing.T) {
	repo := &createSpyRepo{}
	uc := &executionUsecase{repo: repo}

	if _, err := uc.StartManual(context.Background(), "2", "hostname", "admin"); err != nil {
		t.Fatalf("StartManual: %v", err)
	}

	if repo.created.Status == domain.ExecutionStatusPending {
		t.Fatal("manual execution opened as PENDING: the dispatcher will claim it")
	}
	if repo.created.Status != domain.ExecutionStatusExecuting {
		t.Errorf("status = %q, want %q", repo.created.Status, domain.ExecutionStatusExecuting)
	}
}
