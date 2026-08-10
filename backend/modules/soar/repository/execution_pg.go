package repository

import (
	"context"
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
		Where("id = ? AND origin = ? AND status = ? AND (claimed_at IS NULL OR claimed_at < ?)",
			id, domain.ExecutionOriginFlow, domain.ExecutionStatusPending, staleBefore).
		Update("claimed_at", time.Now().UTC())
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
