package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/loganalyzer/connectors"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/domain"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/dto"
)

type queryUsecase struct{ repo connectors.QueryRepository }

func NewQueryUsecase(repo connectors.QueryRepository) connectors.QueryUsecase {
	return &queryUsecase{repo: repo}
}

func (u *queryUsecase) Create(ctx context.Context, q *domain.UtmLogAnalyzerQuery, owner string) (*domain.UtmLogAnalyzerQuery, error) {
	if q.ID != 0 {
		return nil, domain.ErrIDForbidden
	}
	if strings.TrimSpace(q.Name) == "" {
		return nil, domain.ErrNameRequired
	}
	now := time.Now().UTC()
	q.CreationDate = now
	q.ModificationDate = now
	if q.Owner == "" {
		q.Owner = owner
	}
	if err := u.repo.Save(ctx, q); err != nil {
		return nil, err
	}
	return q, nil
}

func (u *queryUsecase) Update(ctx context.Context, q *domain.UtmLogAnalyzerQuery, owner string) (*domain.UtmLogAnalyzerQuery, error) {
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
	q.CreationDate = existing.CreationDate
	if q.Owner == "" {
		q.Owner = existing.Owner
	}
	q.ModificationDate = time.Now().UTC()
	if err := u.repo.Save(ctx, q); err != nil {
		return nil, err
	}
	return q, nil
}

func (u *queryUsecase) GetByID(ctx context.Context, id uint64) (*domain.UtmLogAnalyzerQuery, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *queryUsecase) List(ctx context.Context, f dto.QueryFilter) ([]domain.UtmLogAnalyzerQuery, int64, error) {
	return u.repo.List(ctx, f)
}

func (u *queryUsecase) Delete(ctx context.Context, id uint64) error {
	return u.repo.Delete(ctx, id)
}
