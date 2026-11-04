package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

type pgExecutionRepository struct {
	db *gorm.DB
}

func NewExecutionRepository(db *gorm.DB) connectors.ExecutionRepository {
	return &pgExecutionRepository{db: db}
}

func (r *pgExecutionRepository) Create(ctx context.Context, e *domain.SoarExecution) (*domain.SoarExecution, error) {
	if e.StartedAt.IsZero() {
		e.StartedAt = time.Now().UTC()
	}
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return nil, err
	}
	return e, nil
}

func (r *pgExecutionRepository) Get(ctx context.Context, id uuid.UUID) (*domain.SoarExecution, error) {
	var e domain.SoarExecution
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrIncidentRecordNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *pgExecutionRepository) List(ctx context.Context, f connectors.ExecutionFilters) ([]domain.SoarExecution, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.SoarExecution{})

	if f.Origin != "" {
		q = q.Where("origin = ?", f.Origin)
	}
	if f.RulePath != "" {
		q = q.Where("rule_path = ?", f.RulePath)
	}
	if f.AlertID != "" {
		q = q.Where("alert_id ILIKE ?", "%"+f.AlertID+"%")
	}
	if f.Agent != "" {
		q = q.Where("agent ILIKE ?", "%"+f.Agent+"%")
	}
	if f.TriggeredBy != "" {
		q = q.Where("triggered_by ILIKE ?", "%"+f.TriggeredBy+"%")
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.NonExecutionCause != "" {
		q = q.Where("non_execution_cause = ?", f.NonExecutionCause)
	}
	if f.StartedAtGTE != "" {
		q = q.Where("started_at >= ?", f.StartedAtGTE)
	}
	if f.StartedAtLTE != "" {
		q = q.Where("started_at <= ?", f.StartedAtLTE)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var executions []domain.SoarExecution
	if err := q.Order("started_at DESC, id DESC").
		Offset(f.Offset()).
		Limit(f.Limit()).
		Find(&executions).Error; err != nil {
		return nil, 0, err
	}
	return executions, total, nil
}

func (r *pgExecutionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, u connectors.ExecutionStatusUpdate) error {
	updates := map[string]any{}
	if u.Status != nil {
		updates["status"] = *u.Status
	}
	if u.Result != nil {
		updates["result"] = *u.Result
	}
	if u.NonExecutionCause != nil {
		updates["non_execution_cause"] = *u.NonExecutionCause
	}
	if u.FinishedAt != nil {
		updates["finished_at"] = *u.FinishedAt
	}
	if u.DispatchedAt != nil {
		updates["dispatched_at"] = *u.DispatchedAt
	}
	if u.IncrementRetries {
		updates["retries"] = gorm.Expr("COALESCE(retries, 0) + 1")
	}
	if len(updates) == 0 {
		return nil
	}

	res := r.db.WithContext(ctx).
		Model(&domain.SoarExecution{}).
		Where("id = ?", id).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrIncidentRecordNotFound
	}
	return nil
}

func (r *pgExecutionRepository) ClaimPending(ctx context.Context, id uuid.UUID, leaseDuration time.Duration) (bool, error) {
	staleBefore := time.Now().UTC().Add(-leaseDuration)
	res := r.db.WithContext(ctx).
		Model(&domain.SoarExecution{}).
		Where("id = ? AND status = ? AND (claimed_at IS NULL OR claimed_at < ?)",
			id, domain.ExecutionStatusPending, staleBefore).
		Update("claimed_at", time.Now().UTC())
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *pgExecutionRepository) SaveOutput(ctx context.Context, id uuid.UUID, output []byte) error {
	return r.db.WithContext(ctx).
		Model(&domain.SoarExecution{}).
		Where("id = ?", id).
		Update("output", output).Error
}

func (r *pgExecutionRepository) SettleExecuting(ctx context.Context, id uuid.UUID, u connectors.ExecutionStatusUpdate) (bool, error) {
	updates := map[string]any{}
	if u.Status != nil {
		updates["status"] = *u.Status
	}
	if u.Result != nil {
		updates["result"] = *u.Result
	}
	if u.NonExecutionCause != nil {
		updates["non_execution_cause"] = *u.NonExecutionCause
	}
	if u.FinishedAt != nil {
		updates["finished_at"] = *u.FinishedAt
	}
	if len(updates) == 0 {
		return false, nil
	}
	res := r.db.WithContext(ctx).
		Model(&domain.SoarExecution{}).
		Where("id = ? AND status = ?", id, domain.ExecutionStatusExecuting).
		Updates(updates)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *pgExecutionRepository) RequeueStuck(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.SoarExecution{}).
		Where("id = ? AND status = ?", id, domain.ExecutionStatusExecuting).
		Updates(map[string]any{"status": domain.ExecutionStatusPending, "claimed_at": nil}).Error
}

func (r *pgExecutionRepository) CloseStuckLost(ctx context.Context, id uuid.UUID) (bool, error) {
	cause := domain.NonExecutionCauseDispatchLost
	failed := domain.ExecutionStatusFailed
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).
		Model(&domain.SoarExecution{}).
		Where("id = ? AND status = ?", id, domain.ExecutionStatusExecuting).
		Updates(map[string]any{"status": failed, "non_execution_cause": cause, "finished_at": now})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *pgExecutionRepository) ListStuckExecuting(ctx context.Context, cutoff time.Time) ([]domain.SoarExecution, error) {
	var out []domain.SoarExecution
	err := r.db.WithContext(ctx).
		Where("status = ? AND claimed_at IS NOT NULL AND claimed_at < ?",
			domain.ExecutionStatusExecuting, cutoff).
		Order("claimed_at ASC").
		Find(&out).Error
	return out, err
}

func (r *pgExecutionRepository) RecordEdge(ctx context.Context, req connectors.RecordEdgeRequest) (*domain.SoarExecution, error) {
	var child *domain.SoarExecution
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		seq, err := nextSequence(tx, req.FlowRunID, req.ChildNodeID, req.ChildDepth)
		if err != nil {
			return err
		}
		created := domain.SoarExecution{
			TenantID:  req.TenantID,
			Origin:    domain.ExecutionOriginFlow,
			FlowRunID: &req.FlowRunID,
			RulePath:  req.RulePath,
			AlertID:   req.AlertID,
			NodeID:    req.ChildNodeID,
			Depth:     req.ChildDepth,
			Sequence:  seq,
			Kind:      req.ChildKind,
			Executor:  req.ChildExecutor,
			Status:    domain.ExecutionStatusWaiting,
			StartedAt: time.Now().UTC(),
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		edge := domain.SoarExecutionEdge{
			ChildExecID:  created.ID,
			ParentExecID: req.Parent.ID,
			FlowRunID:    req.FlowRunID,
			Branch:       req.Branch,
			Fired:        req.Fired,
			CreatedAt:    time.Now().UTC(),
		}
		if err := tx.Create(&edge).Error; err != nil {
			return err
		}
		child = &created
		return nil
	})
	return child, err
}

func (r *pgExecutionRepository) RecordDeadBranch(ctx context.Context, req connectors.RecordEdgeRequest) error {
	seq, err := nextSequence(r.db.WithContext(ctx), req.FlowRunID, req.ChildNodeID, req.ChildDepth)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	created := domain.SoarExecution{
		TenantID:   req.TenantID,
		Origin:     domain.ExecutionOriginFlow,
		FlowRunID:  &req.FlowRunID,
		RulePath:   req.RulePath,
		AlertID:    req.AlertID,
		NodeID:     req.ChildNodeID,
		Depth:      req.ChildDepth,
		Sequence:   seq,
		Kind:       req.ChildKind,
		Executor:   req.ChildExecutor,
		Status:     domain.ExecutionStatusDead,
		StartedAt:  now,
		FinishedAt: &now,
	}
	err = r.db.WithContext(ctx).Create(&created).Error
	if err != nil {
		if isDuplicateKey(err) {
			return r.RecordDeadBranch(ctx, req)
		}
		return err
	}
	edge := domain.SoarExecutionEdge{
		ChildExecID:  created.ID,
		ParentExecID: req.Parent.ID,
		FlowRunID:    req.FlowRunID,
		Branch:       req.Branch,
		Fired:        req.Fired,
		CreatedAt:    now,
	}
	return r.db.WithContext(ctx).Create(&edge).Error
}

func isDuplicateKey(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key")
}

func nextSequence(db *gorm.DB, runID uuid.UUID, nodeID string, depth int) (int, error) {
	var next int
	if err := db.Model(&domain.SoarExecution{}).
		Where("flow_run_id = ? AND node_id = ? AND depth = ?",
			runID, nodeID, depth).
		Order("sequence DESC").
		Limit(1).
		Pluck("sequence", &next).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		} else {
			return 0, err
		}
	}
	if err := db.Exec(`UPDATE soar_executions SET sequence = ?
		WHERE flow_run_id = ? AND node_id = ? AND depth = ?
			AND status = ? AND sequence < ?`,
		next, runID, nodeID, depth, domain.ExecutionStatusWaiting, next).Error; err != nil {
		return 0, err
	}
	return next, nil
}

func (r *pgExecutionRepository) ListFiredParents(ctx context.Context, childID uuid.UUID) ([]domain.SoarExecution, error) {
	var parents []domain.SoarExecution
	err := r.db.WithContext(ctx).
		Raw(`SELECT p.* FROM soar_executions p
		     JOIN soar_execution_edges e ON e.parent_exec_id = p.id
		     WHERE e.child_exec_id = ? AND e.fired = TRUE`, childID).
		Scan(&parents).Error
	return parents, err
}

func (r *pgExecutionRepository) TransitionReady(ctx context.Context, id uuid.UUID, ready connectors.ReadyUpdate) error {
	updates := map[string]any{"status": ready.Status}
	if len(ready.Context) > 0 {
		updates["context"] = ready.Context
	}
	if len(ready.Params) > 0 {
		updates["params"] = ready.Params
	}
	if ready.Command != "" {
		updates["command"] = ready.Command
	}
	if ready.Shell != "" {
		updates["shell"] = ready.Shell
	}
	if ready.Agent != "" {
		updates["agent"] = ready.Agent
	}
	res := r.db.WithContext(ctx).
		Model(&domain.SoarExecution{}).
		Where("id = ? AND status = ?", id, domain.ExecutionStatusWaiting).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	return nil
}
