package usecase

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/correlation/connectors"
	"github.com/utmstack/utmstack/backend/modules/correlation/domain"
	"github.com/utmstack/utmstack/backend/modules/correlation/dto"
	correrrors "github.com/utmstack/utmstack/backend/modules/correlation/errors"
)

type regexPatternUsecase struct {
	repo connectors.RegexPatternRepository
}

func NewRegexPatternUsecase(repo connectors.RegexPatternRepository) connectors.RegexPatternUsecase {
	return &regexPatternUsecase{repo: repo}
}

func (u *regexPatternUsecase) Create(ctx context.Context, req dto.CreateRegexPatternRequest) (*dto.RegexPatternResponse, error) {
	if req.ID != nil {
		return nil, correrrors.ErrIDMustBeAbsent
	}
	now := time.Now().UTC()
	p := &domain.UtmRegexPattern{
		PatternID:          req.PatternID,
		PatternDescription: req.PatternDescription,
		PatternDefinition:  req.PatternDefinition,
		SystemOwner:        false,
		LastUpdate:         &now,
	}
	saved, err := u.repo.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return dto.RegexPatternToResponse(saved), nil
}

func (u *regexPatternUsecase) Update(ctx context.Context, req dto.UpdateRegexPatternRequest) (*dto.RegexPatternResponse, error) {
	if req.ID == nil || *req.ID == 0 {
		return nil, correrrors.ErrIDRequired
	}
	existing, err := u.repo.GetByID(ctx, *req.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, correrrors.ErrRegexPatternNotFound
	}
	if existing.SystemOwner {
		return nil, correrrors.ErrRegexPatternSystemOwner
	}
	now := time.Now().UTC()
	existing.PatternID = req.PatternID
	existing.PatternDescription = req.PatternDescription
	existing.PatternDefinition = req.PatternDefinition
	existing.LastUpdate = &now

	saved, err := u.repo.Update(ctx, existing)
	if err != nil {
		return nil, err
	}
	return dto.RegexPatternToResponse(saved), nil
}

func (u *regexPatternUsecase) GetByID(ctx context.Context, id int64) (*dto.RegexPatternResponse, error) {
	p, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, correrrors.ErrRegexPatternNotFound
	}
	return dto.RegexPatternToResponse(p), nil
}

func (u *regexPatternUsecase) List(ctx context.Context, f dto.RegexPatternFilters) (*connectors.ListResult[dto.RegexPatternResponse], error) {
	page, size := normPage(f.Page, f.Size)
	items, total, err := u.repo.List(ctx, connectors.RegexPatternFilters{
		Search: f.Search,
		Page:   page,
		Size:   size,
	})
	if err != nil {
		return nil, err
	}
	responses := make([]dto.RegexPatternResponse, len(items))
	for i := range items {
		responses[i] = *dto.RegexPatternToResponse(&items[i])
	}
	return &connectors.ListResult[dto.RegexPatternResponse]{Items: responses, Total: total}, nil
}

func (u *regexPatternUsecase) Delete(ctx context.Context, id int64) error {
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return correrrors.ErrRegexPatternNotFound
	}
	if existing.SystemOwner {
		return correrrors.ErrRegexPatternSystemOwner
	}
	return u.repo.Delete(ctx, id)
}

func normPage(page, size int) (int, int) {
	if page < 0 {
		page = 0
	}
	if size < 1 || size > 200 {
		size = 20
	}
	return page, size
}
