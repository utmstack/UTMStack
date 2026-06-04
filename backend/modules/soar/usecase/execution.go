package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

type executionUsecase struct {
	repo connectors.ExecutionRepository
}

func NewExecutionUsecase(repo connectors.ExecutionRepository) connectors.ExecutionUsecase {
	return &executionUsecase{repo: repo}
}

func (u *executionUsecase) List(ctx context.Context, f dto.ExecutionFilters) (*connectors.ListResult[dto.ExecutionResponse], error) {
	page, size := normPage(f.Page, f.Size)

	executions, total, err := u.repo.List(ctx, connectors.ExecutionFilters{
		ID:                       f.ID,
		RuleID:                   f.RuleID,
		RuleIDGreaterThanOrEqual: f.RuleIDGreaterThanOrEqual,
		RuleIDLessThanOrEqual:    f.RuleIDLessThanOrEqual,
		AlertID:                  f.AlertID,
		Agent:                    f.Agent,
		ExecutionStatus:          f.ExecutionStatus,
		NonExecutionCause:        f.NonExecutionCause,
		ExecutionDateGTE:         f.ExecutionDateGTE,
		ExecutionDateLTE:         f.ExecutionDateLTE,
		Page:                     page,
		Size:                     size,
	})
	if err != nil {
		return nil, err
	}

	items := make([]dto.ExecutionResponse, len(executions))
	for i, e := range executions {
		items[i] = dto.ExecutionResponse{
			ID:                e.ID,
			RuleID:            e.RuleID,
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

	return &connectors.ListResult[dto.ExecutionResponse]{Items: items, Total: total}, nil
}
