package usecase

import (
	"fmt"

	"github.com/utmstack/utmstack/backend/modules/incident_response/connectors"
	"github.com/utmstack/utmstack/backend/modules/incident_response/domain"
	"github.com/utmstack/utmstack/backend/modules/incident_response/dto"
)

type actionCommandUsecase struct {
	repo connectors.ActionCommandRepository
}

func NewActionCommandUsecase(repo connectors.ActionCommandRepository) connectors.ActionCommandUsecase {
	return &actionCommandUsecase{repo: repo}
}

func (u *actionCommandUsecase) Create(req dto.CreateActionCommandRequest) (*domain.UtmIncidentActionCommand, error) {
	c := &domain.UtmIncidentActionCommand{
		ActionID:   req.ActionID,
		OsPlatform: req.OsPlatform,
		Command:    req.Command,
	}
	if err := u.repo.Save(c); err != nil {
		return nil, fmt.Errorf("actionCommandUsecase.Create: %w", err)
	}
	return c, nil
}

func (u *actionCommandUsecase) Update(req dto.UpdateActionCommandRequest) (*domain.UtmIncidentActionCommand, error) {
	c, err := u.repo.FindByID(req.ID)
	if err != nil {
		return nil, err
	}
	c.ActionID = req.ActionID
	c.OsPlatform = req.OsPlatform
	c.Command = req.Command
	if err := u.repo.Save(c); err != nil {
		return nil, fmt.Errorf("actionCommandUsecase.Update: %w", err)
	}
	return c, nil
}

func (u *actionCommandUsecase) FindByID(id int64) (*domain.UtmIncidentActionCommand, error) {
	return u.repo.FindByID(id)
}

func (u *actionCommandUsecase) FindAll(f dto.ActionCommandFilter) ([]domain.UtmIncidentActionCommand, int64, error) {
	return u.repo.FindAll(f)
}

func (u *actionCommandUsecase) Delete(id int64) error {
	return u.repo.Delete(id)
}
