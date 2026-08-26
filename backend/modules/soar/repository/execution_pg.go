package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

// RecordEdge is the atomic hinge of the DAG engine. Everything happens inside
// one transaction so concurrent parents can't race the pending_parents counter.
func (r *pgExecutionRepository) RecordEdge(ctx context.Context, req connectors.RecordEdgeRequest) (*domain.SoarExecution, error) {
	var child *domain.SoarExecution
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1) find-or-create child, keyed by (flow_run_id, node_id, depth).
		var existing domain.SoarExecution
		findErr := tx.Where("flow_run_id = ? AND node_id = ? AND depth = ?",
			req.FlowRunID, req.ChildNodeID, req.ChildDepth).
			First(&existing).Error

		var created domain.SoarExecution
		switch {
		case errors.Is(findErr, gorm.ErrRecordNotFound):
			created = domain.SoarExecution{
				TenantID:       req.TenantID,
				Origin:         domain.ExecutionOriginFlow,
				FlowRunID:      &req.FlowRunID,
				RulePath:       req.RulePath,
				AlertID:        req.AlertID,
				NodeID:         req.ChildNodeID,
				Depth:          req.ChildDepth,
				Kind:           req.ChildKind,
				Executor:       req.ChildExecutor,
				PendingParents: req.IncomingCount,
				Status:         domain.ExecutionStatusWaiting,
				StartedAt:      time.Now().UTC(),
			}
			// ON CONFLICT protects against two parents racing the initial insert.
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "flow_run_id"}, {Name: "node_id"}, {Name: "depth"}},
				DoNothing: true,
			}).Create(&created).Error; err != nil {
				return err
			}
			// Re-read to pick up the winning row after ON CONFLICT.
			if err := tx.Where("flow_run_id = ? AND node_id = ? AND depth = ?",
				req.FlowRunID, req.ChildNodeID, req.ChildDepth).
				First(&existing).Error; err != nil {
				return err
			}
		case findErr != nil:
			return findErr
		}

		// 2) record the edge (may already exist if the parent retries; ignore
		// duplicates by primary key).
		edge := domain.SoarExecutionEdge{
			ChildExecID:  existing.ID,
			ParentExecID: req.Parent.ID,
			FlowRunID:    req.FlowRunID,
			Branch:       req.Branch,
			Fired:        req.Fired,
			CreatedAt:    time.Now().UTC(),
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&edge).Error; err != nil {
			return err
		}

		// 3) lock the child row and decrement counters. Only decrement when the
		// edge was fresh (RowsAffected on the edge insert), otherwise this call
		// is a retry.
		var locked domain.SoarExecution
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", existing.ID).First(&locked).Error; err != nil {
			return err
		}
		if !locked.Status.Terminal() && locked.Status == domain.ExecutionStatusWaiting {
			updates := map[string]any{
				"pending_parents": gorm.Expr("GREATEST(pending_parents - 1, 0)"),
			}
			if !req.Fired {
				updates["dead_parents"] = gorm.Expr("dead_parents + 1")
			}
			if err := tx.Model(&domain.SoarExecution{}).
				Where("id = ?", locked.ID).Updates(updates).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("id = ?", existing.ID).First(&locked).Error; err != nil {
			return err
		}
		child = &locked
		return nil
	})
	return child, err
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
