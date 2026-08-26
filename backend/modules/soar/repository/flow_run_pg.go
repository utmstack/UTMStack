package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

type pgFlowRunRepository struct {
	db *gorm.DB
}

func NewFlowRunRepository(db *gorm.DB) connectors.FlowRunRepository {
	return &pgFlowRunRepository{db: db}
}

func (r *pgFlowRunRepository) Create(ctx context.Context, run *domain.SoarFlowRun) (*domain.SoarFlowRun, error) {
	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now().UTC()
	}
	if run.Status == "" {
		run.Status = domain.ExecutionStatusPending
	}
	if err := r.db.WithContext(ctx).Create(run).Error; err != nil {
		return nil, err
	}
	return run, nil
}

func (r *pgFlowRunRepository) Get(ctx context.Context, id uuid.UUID) (*domain.SoarFlowRun, error) {
	var run domain.SoarFlowRun
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrIncidentRecordNotFound
		}
		return nil, err
	}
	return &run, nil
}

// MaybeComplete inspects every execution in the run and transitions the run to
// EXECUTED (all children in a terminal-good state) or FAILED (any node ended
// in FAILED/DEAD with no compensating success). A run with any non-terminal
// execution is left alone.
func (r *pgFlowRunRepository) MaybeComplete(ctx context.Context, id uuid.UUID) (bool, error) {
	var counts struct {
		NonTerminal int64
		Failed      int64
		Executed    int64
	}
	if err := r.db.WithContext(ctx).
		Raw(`SELECT
		     SUM(CASE WHEN status IN (?, ?, ?) THEN 1 ELSE 0 END) AS non_terminal,
		     SUM(CASE WHEN status IN (?, ?) THEN 1 ELSE 0 END) AS failed,
		     SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS executed
		     FROM soar_executions WHERE flow_run_id = ?`,
			domain.ExecutionStatusWaiting, domain.ExecutionStatusPending, domain.ExecutionStatusExecuting,
			domain.ExecutionStatusFailed, domain.ExecutionStatusDead,
			domain.ExecutionStatusExecuted,
			id).
		Scan(&counts).Error; err != nil {
		return false, err
	}
	if counts.NonTerminal > 0 {
		return false, nil
	}
	next := domain.ExecutionStatusExecuted
	if counts.Executed == 0 && counts.Failed > 0 {
		next = domain.ExecutionStatusFailed
	}
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).
		Model(&domain.SoarFlowRun{}).
		Where("id = ? AND finished_at IS NULL", id).
		Updates(map[string]any{"status": next, "finished_at": now})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
