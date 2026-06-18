package repository

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"gorm.io/gorm"
)

type pgExecutionRepository struct {
	db *gorm.DB
}

func NewExecutionRepository(db *gorm.DB) connectors.ExecutionRepository {
	return &pgExecutionRepository{db: db}
}

func (r *pgExecutionRepository) Create(ctx context.Context, e *domain.AlertResponseRuleExecution) (*domain.AlertResponseRuleExecution, error) {
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return nil, err
	}
	return e, nil
}

func (r *pgExecutionRepository) List(ctx context.Context, f connectors.ExecutionFilters) ([]domain.AlertResponseRuleExecution, int64, error) {

	q := r.db.WithContext(ctx).Model(&domain.AlertResponseRuleExecution{})

	// id.equals
	if f.ID != 0 {
		q = q.Where("id = ?", f.ID)
	}
	// rulePath.equals
	if f.RulePath != "" {
		q = q.Where("rule_path = ?", f.RulePath)
	}
	// alertId.contains
	if f.AlertID != "" {
		q = q.Where("alert_id ILIKE ?", "%"+f.AlertID+"%")
	}
	// agent.contains
	if f.Agent != "" {
		q = q.Where("agent ILIKE ?", "%"+f.Agent+"%")
	}
	// executionStatus.equals
	if f.ExecutionStatus != "" {
		q = q.Where("execution_status = ?", f.ExecutionStatus)
	}
	// nonExecutionCause.equals
	if f.NonExecutionCause != "" {
		q = q.Where("non_execution_cause = ?", f.NonExecutionCause)
	}
	// executionDate.greaterThanOrEqual
	if f.ExecutionDateGTE != "" {
		q = q.Where("execution_date >= ?", f.ExecutionDateGTE)
	}
	// executionDate.lessThanOrEqual
	if f.ExecutionDateLTE != "" {
		q = q.Where("execution_date <= ?", f.ExecutionDateLTE)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var executions []domain.AlertResponseRuleExecution
	if err := q.Order("id DESC").
		Offset(f.Offset()).
		Limit(f.Limit()).
		Find(&executions).Error; err != nil {
		return nil, 0, err
	}
	return executions, total, nil
}

// Retry bumps are server-side increments so concurrent dispatcher ticks don't lose counts.
func (r *pgExecutionRepository) UpdateStatus(ctx context.Context, id int64, u connectors.ExecutionStatusUpdate) error {
	updates := map[string]any{}
	if u.ExecutionStatus != nil {
		updates["execution_status"] = *u.ExecutionStatus
	}
	if u.CommandResult != nil {
		updates["command_result"] = *u.CommandResult
	}
	if u.NonExecutionCause != nil {
		updates["non_execution_cause"] = *u.NonExecutionCause
	}
	if u.IncrementRetries {
		updates["execution_retries"] = gorm.Expr("COALESCE(execution_retries, 0) + 1")
	}
	if len(updates) == 0 {
		return nil
	}

	res := r.db.WithContext(ctx).
		Model(&domain.AlertResponseRuleExecution{}).
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
