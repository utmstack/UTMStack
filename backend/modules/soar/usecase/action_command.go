package usecase

import (
	"context"
	"fmt"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

type actionCommandUsecase struct {
	repo connectors.ActionCommandRepository
}

func NewActionCommandUsecase(repo connectors.ActionCommandRepository) connectors.ActionCommandUsecase {
	return &actionCommandUsecase{repo: repo}
}

func (u *actionCommandUsecase) Create(ctx context.Context, req dto.CreateActionCommandRequest) (*domain.UtmIncidentActionCommand, error) {
	c := &domain.UtmIncidentActionCommand{
		ActionID:   req.ActionID,
		OsPlatform: req.OsPlatform,
		Command:    req.Command,
	}
	if err := u.repo.Save(ctx, c); err != nil {
		return nil, fmt.Errorf("actionCommandUsecase.Create: %w", err)
	}
	return c, nil
}

func (u *actionCommandUsecase) Update(ctx context.Context, req dto.UpdateActionCommandRequest) (*domain.UtmIncidentActionCommand, error) {
	c, err := u.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	c.ActionID = req.ActionID
	c.OsPlatform = req.OsPlatform
	c.Command = req.Command
	if err := u.repo.Save(ctx, c); err != nil {
		return nil, fmt.Errorf("actionCommandUsecase.Update: %w", err)
	}
	return c, nil
}

func (u *actionCommandUsecase) FindByID(ctx context.Context, id int64) (*domain.UtmIncidentActionCommand, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *actionCommandUsecase) FindAll(ctx context.Context, f dto.ActionCommandFilter) ([]domain.UtmIncidentActionCommand, int64, error) {
	return u.repo.FindAll(ctx, f)
}

func (u *actionCommandUsecase) Delete(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}
