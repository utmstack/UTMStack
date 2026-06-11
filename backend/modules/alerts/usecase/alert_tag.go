package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
)

type alertTagUsecase struct {
	repo connectors.AlertTagRepository
}

func NewAlertTagUsecase(repo connectors.AlertTagRepository) connectors.AlertTagUsecase {
	return &alertTagUsecase{repo: repo}
}

func (u *alertTagUsecase) Create(ctx context.Context, req dto.CreateAlertTagRequest) (*domain.UtmAlertTag, error) {
	tag := domain.UtmAlertTag{
		TagName:     req.TagName,
		TagColor:    req.TagColor,
		SystemOwner: false,
	}
	return u.repo.Create(ctx, tag)
}

func (u *alertTagUsecase) Update(ctx context.Context, req dto.UpdateAlertTagRequest) (*domain.UtmAlertTag, error) {
	existing, err := u.repo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrAlertTagNotFound
	}
	if existing.SystemOwner {
		return nil, domain.ErrAlertTagSystemOwned
	}

	existing.TagName = req.TagName
	existing.TagColor = req.TagColor

	return u.repo.Update(ctx, *existing)
}

func (u *alertTagUsecase) List(ctx context.Context, filters dto.AlertTagFilters) ([]domain.UtmAlertTag, int64, error) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Size < 1 || filters.Size > 200 {
		filters.Size = 20
	}

	rows, total, err := u.repo.List(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (u *alertTagUsecase) GetByID(ctx context.Context, id uint64) (*domain.UtmAlertTag, error) {
	row, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, domain.ErrAlertTagNotFound
	}
	return row, nil
}

func (u *alertTagUsecase) Delete(ctx context.Context, id uint64) error {
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrAlertTagNotFound
	}
	if existing.SystemOwner {
		return domain.ErrAlertTagSystemOwned
	}
	return u.repo.Delete(ctx, id)
}
