package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/adaudit/connectors"
	"github.com/utmstack/utmstack/backend/modules/adaudit/domain"
	"github.com/utmstack/utmstack/backend/modules/adaudit/dto"
	"github.com/utmstack/utmstack/backend/modules/adaudit/repository"
)

var (
	ErrIdentityNameInvalid = errors.New("name must be 2 to 64 characters")
)

type identityUsecase struct {
	repo connectors.IdentityRepository
}

func NewIdentityUsecase(repo connectors.IdentityRepository) connectors.IdentityUsecase {
	return &identityUsecase{repo: repo}
}

func (u *identityUsecase) Create(ctx context.Context, name, actor string) (*dto.IdentityResponse, error) {
	name = strings.TrimSpace(name)
	if !domain.ValidIdentityName(name) {
		return nil, ErrIdentityNameInvalid
	}
	alias, err := u.repo.Create(ctx, name, actor)
	if err != nil {
		if errors.Is(err, repository.ErrIdentityNameTaken) {
			return nil, err
		}
		return nil, err
	}
	return &dto.IdentityResponse{ID: alias.ID.String(), Name: alias.Name, CreatedAt: alias.CreatedAt}, nil
}

func (u *identityUsecase) List(ctx context.Context, search string) ([]dto.IdentityListItem, error) {
	return u.repo.List(ctx, search)
}

func (u *identityUsecase) Detail(ctx context.Context, id uuid.UUID) (*dto.IdentityDetail, error) {
	alias, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	users, err := u.repo.Users(ctx, id)
	if err != nil {
		return nil, err
	}
	return &dto.IdentityDetail{
		ID:        alias.ID.String(),
		Name:      alias.Name,
		CreatedAt: alias.CreatedAt,
		Users:     users,
	}, nil
}

func (u *identityUsecase) Delete(ctx context.Context, id uuid.UUID) (int, error) {
	return u.repo.Delete(ctx, id)
}

func (u *identityUsecase) Assign(ctx context.Context, userID, aliasID uuid.UUID, actor string) error {
	return u.repo.Assign(ctx, userID, aliasID, actor)
}

func (u *identityUsecase) Unassign(ctx context.Context, userID uuid.UUID, actor string) error {
	return u.repo.Unassign(ctx, userID, actor)
}
