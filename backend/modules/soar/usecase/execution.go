package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type executionUsecase struct {
	repo connectors.ExecutionRepository
}

func NewExecutionUsecase(repo connectors.ExecutionRepository) connectors.ExecutionUsecase {
	return &executionUsecase{repo: repo}
}

func (u *executionUsecase) Create(ctx context.Context, req dto.CreateExecutionRequest) (*dto.ExecutionResponse, error) {
	e := &domain.AlertResponseRuleExecution{
		RulePath:        req.RulePath,
		AlertID:         req.AlertID,
		Command:         req.Command,
		Agent:           req.Agent,
		ExecutionStatus: domain.ExecutionStatusPending,
	}
	saved, err := u.repo.Create(ctx, e)
	if err != nil {
		return nil, err
	}
	return &dto.ExecutionResponse{
		ID:               saved.ID,
		RulePath:         saved.RulePath,
		AlertID:          saved.AlertID,
		Command:          saved.Command,
		Agent:            saved.Agent,
		ExecutionDate:    saved.ExecutionDate,
		ExecutionStatus:  saved.ExecutionStatus,
		ExecutionRetries: saved.ExecutionRetries,
	}, nil
}

func (u *executionUsecase) UpdateStatus(ctx context.Context, id int64, req dto.UpdateExecutionRequest) error {
	return u.repo.UpdateStatus(ctx, id, connectors.ExecutionStatusUpdate{
		ExecutionStatus:   req.ExecutionStatus,
		CommandResult:     req.CommandResult,
		NonExecutionCause: req.NonExecutionCause,
		IncrementRetries:  req.IncrementRetries,
	})
}

func (u *executionUsecase) List(ctx context.Context, f dto.ExecutionFilters) (*database.List[dto.ExecutionResponse], error) {
	executions, total, err := u.repo.List(ctx, connectors.ExecutionFilters{
		ID:                f.ID,
		RulePath:          f.RulePath,
		AlertID:           f.AlertID,
		Agent:             f.Agent,
		ExecutionStatus:   f.ExecutionStatus,
		NonExecutionCause: f.NonExecutionCause,
		ExecutionDateGTE:  f.ExecutionDateGTE,
		ExecutionDateLTE:  f.ExecutionDateLTE,
		Params:            f.Params,
	})
	if err != nil {
		return nil, err
	}

	items := make([]dto.ExecutionResponse, len(executions))
	for i, e := range executions {
		items[i] = dto.ExecutionResponse{
			ID:                e.ID,
			RulePath:          e.RulePath,
			AlertID:           e.AlertID,
			Command:           e.Command,
			CommandResult:     e.CommandResult,
			Agent:             e.Agent,
			ExecutionDate:     e.ExecutionDate,
			ExecutionStatus:   e.ExecutionStatus,
			NonExecutionCause: e.NonExecutionCause,
			ExecutionRetries:  e.ExecutionRetries,
		}
	}

	return &database.List[dto.ExecutionResponse]{Items: items, Total: total}, nil
}
