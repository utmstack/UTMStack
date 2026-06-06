package usecase

import (
	"context"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
)

type filterUsecase struct {
	store *FilterStore
}

func NewFilterUsecase(store *FilterStore) connectors.FilterUsecase {
	return &filterUsecase{store: store}
}

func (u *filterUsecase) Create(_ context.Context, req dto.CreateFilterRequest) (*dto.FilterResponse, error) {
	entry, err := u.store.Create(req.RelPath, []byte(req.Content))
	if err != nil {
		return nil, mapStoreFilterErr(err)
	}
	return toFilterResponse(entry), nil
}

func (u *filterUsecase) Update(_ context.Context, req dto.UpdateFilterRequest) (*dto.FilterResponse, error) {
	entry, err := u.store.Update(req.RelPath, []byte(req.Content))
	if err != nil {
		return nil, mapStoreFilterErr(err)
	}
	return toFilterResponse(entry), nil
}

func (u *filterUsecase) GetByRelPath(_ context.Context, relPath string) (*dto.FilterResponse, error) {
	entry := u.store.GetByRelPath(relPath)
	if entry == nil {
		return nil, domain.ErrFilterNotFound
	}
	return toFilterResponse(entry), nil
}

func (u *filterUsecase) List(_ context.Context, f dto.FilterFilters) ([]dto.FilterResponse, int64, error) {
	all := u.store.List()

	// Apply in-memory filters.
	out := make([]dto.FilterResponse, 0, len(all))
	for i := range all {
		e := &all[i]
		if f.IsActiveEq != nil && e.Active != *f.IsActiveEq {
			continue
		}
		if f.SystemEq != nil && e.System != *f.SystemEq {
			continue
		}
		if f.RelPathContains != nil && !strings.Contains(e.RelPath, *f.RelPathContains) {
			continue
		}
		out = append(out, *toFilterResponse(e))
	}

	total := int64(len(out))

	// Pagination.
	page, size := f.Page, f.Size
	if size <= 0 {
		size = 50
	}
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * size
	if start >= len(out) {
		return []dto.FilterResponse{}, total, nil
	}
	end := start + size
	if end > len(out) {
		end = len(out)
	}
	return out[start:end], total, nil
}

func (u *filterUsecase) Delete(_ context.Context, relPath string) error {
	return mapStoreFilterErr(u.store.Delete(relPath))
}

func (u *filterUsecase) SetActive(_ context.Context, relPath string, active bool) error {
	return mapStoreFilterErr(u.store.SetEnabled(relPath, active))
}

func toFilterResponse(e *FilterEntry) *dto.FilterResponse {
	return &dto.FilterResponse{
		RelPath: e.RelPath,
		Content: string(e.Content),
		System:  e.System,
		Active:  e.Active,
	}
}

func mapStoreFilterErr(err error) error {
	if err == nil {
		return nil
	}
	switch err.(type) {
	case errFilterNotFound:
		return domain.ErrFilterNotFound
	case errFilterSystemOwner:
		return domain.ErrFilterSystemOwner
	}
	return err
}
