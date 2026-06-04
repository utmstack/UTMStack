package usecase

import (
	"fmt"
	"time"

	"github.com/utmstack/utmstack/backend/modules/incident_response/connectors"
	"github.com/utmstack/utmstack/backend/modules/incident_response/domain"
	"github.com/utmstack/utmstack/backend/modules/incident_response/dto"
)

type actionUsecase struct {
	repo connectors.ActionRepository
}

func NewActionUsecase(repo connectors.ActionRepository) connectors.ActionUsecase {
	return &actionUsecase{repo: repo}
}

func (u *actionUsecase) Create(req dto.CreateActionRequest, user string) (*domain.UtmIncidentAction, error) {
	a := &domain.UtmIncidentAction{
		ActionCommand:     req.ActionCommand,
		ActionDescription: req.ActionDescription,
		ActionParams:      req.ActionParams,
		ActionType:        req.ActionType,
		ActionEditable:    req.ActionEditable,
		CreatedDate:       time.Now().UTC(),
		CreatedUser:       user,
	}
	if err := u.repo.Save(a); err != nil {
		return nil, fmt.Errorf("actionUsecase.Create: %w", err)
	}
	return a, nil
}

func (u *actionUsecase) Update(req dto.UpdateActionRequest, user string) (*domain.UtmIncidentAction, error) {
	a, err := u.repo.FindByID(req.ID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	a.ActionCommand = req.ActionCommand
	a.ActionDescription = req.ActionDescription
	a.ActionParams = req.ActionParams
	a.ActionType = req.ActionType
	a.ActionEditable = req.ActionEditable
	a.ModifiedDate = &now
	a.ModifiedUser = &user
	if err := u.repo.Save(a); err != nil {
		return nil, fmt.Errorf("actionUsecase.Update: %w", err)
	}
	return a, nil
}

func (u *actionUsecase) FindByID(id int64) (*domain.UtmIncidentAction, error) {
	return u.repo.FindByID(id)
}

func (u *actionUsecase) FindAll(f dto.ActionFilter) ([]domain.UtmIncidentAction, int64, error) {
	return u.repo.FindAll(f)
}

func (u *actionUsecase) Delete(id int64) error {
	return u.repo.Delete(id)
}
