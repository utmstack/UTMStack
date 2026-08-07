package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

type actionUsecase struct {
	repo connectors.ActionRepository
}

func NewActionUsecase(repo connectors.ActionRepository) connectors.ActionUsecase {
	return &actionUsecase{repo: repo}
}

func (u *actionUsecase) Create(ctx context.Context, req dto.CreateActionRequest, user string) (*domain.UtmIncidentAction, error) {
	a := &domain.UtmIncidentAction{
		ActionCommand:     req.ActionCommand,
		ActionDescription: req.ActionDescription,
		ActionParams:      req.ActionParams,
		ActionType:        req.ActionType,
		ActionEditable:    req.ActionEditable,
		CreatedDate:       time.Now().UTC(),
		CreatedUser:       user,
	}
	if err := u.repo.Save(ctx, a); err != nil {
		return nil, fmt.Errorf("actionUsecase.Create: %w", err)
	}
	return a, nil
}

func (u *actionUsecase) Update(ctx context.Context, req dto.UpdateActionRequest, user string) (*domain.UtmIncidentAction, error) {
	a, err := u.repo.FindByID(ctx, req.ID)
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
	if err := u.repo.Save(ctx, a); err != nil {
		return nil, fmt.Errorf("actionUsecase.Update: %w", err)
	}
	return a, nil
}

func (u *actionUsecase) FindByID(ctx context.Context, id int64) (*domain.UtmIncidentAction, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *actionUsecase) FindAll(ctx context.Context, f dto.ActionFilter) ([]domain.UtmIncidentAction, int64, error) {
	return u.repo.FindAll(ctx, f)
}

func (u *actionUsecase) Delete(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}
