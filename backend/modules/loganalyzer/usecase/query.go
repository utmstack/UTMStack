package usecase

import (
	"context"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/loganalyzer/connectors"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/domain"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/dto"
)

type queryUsecase struct{ repo connectors.QueryRepository }

func NewQueryUsecase(repo connectors.QueryRepository) connectors.QueryUsecase {
	return &queryUsecase{repo: repo}
}

func (u *queryUsecase) Create(ctx context.Context, q *domain.SavedQuery, owner string) (*domain.SavedQuery, error) {
	if q.ID != 0 {
		return nil, domain.ErrIDForbidden
	}
	if strings.TrimSpace(q.Name) == "" {
		return nil, domain.ErrNameRequired
	}
	// CreatedAt / UpdatedAt are GORM's to fill.
	if q.Owner == "" {
		q.Owner = owner
	}
	if err := u.repo.Save(ctx, q); err != nil {
		return nil, err
	}
	return q, nil
}

func (u *queryUsecase) Update(ctx context.Context, q *domain.SavedQuery, owner string) (*domain.SavedQuery, error) {
	if q.ID == 0 {
		return nil, domain.ErrIDRequired
	}
	existing, err := u.repo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrNotFound
	}

	q.CreatedAt = existing.CreatedAt
	q.TenantID = existing.TenantID
	if q.Owner == "" {
		q.Owner = existing.Owner
	}
	if err := u.repo.Save(ctx, q); err != nil {
		return nil, err
	}
	return q, nil
}

func (u *queryUsecase) GetByID(ctx context.Context, id uint64) (*domain.SavedQuery, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *queryUsecase) List(ctx context.Context, f dto.QueryFilter) ([]domain.SavedQuery, int64, error) {
	return u.repo.List(ctx, f)
}

func (u *queryUsecase) Delete(ctx context.Context, id uint64) error {
	return u.repo.Delete(ctx, id)
}
