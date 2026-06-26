package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/opensearch/connectors"
	"github.com/utmstack/utmstack/backend/modules/opensearch/domain"
	"github.com/utmstack/utmstack/backend/modules/opensearch/dto"
	"github.com/utmstack/utmstack/backend/pkg/constants"
)

var errIDMustBeAbsent = errors.New("id must be absent on create")
var errInvalidPattern = errors.New("invalid pattern")

type indexPatternUsecase struct {
	repo    connectors.IndexPatternRepository
	ismRepo connectors.ISMRepository // optional; nil → fields returns empty list
}

func NewIndexPatternUsecase(repo connectors.IndexPatternRepository) connectors.IndexPatternUsecase {
	return &indexPatternUsecase{repo: repo}
}

func NewIndexPatternUsecaseWithISM(repo connectors.IndexPatternRepository, ismRepo connectors.ISMRepository) connectors.IndexPatternUsecase {
	return &indexPatternUsecase{repo: repo, ismRepo: ismRepo}
}

func (u *indexPatternUsecase) Create(ctx context.Context, req dto.CreateIndexPatternRequest) (*dto.IndexPatternResponse, error) {
	if req.ID != nil {
		return nil, errIDMustBeAbsent
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	systemFalse := false

	if req.Pattern =="" || req.Pattern ==" "{
		return nil,errInvalidPattern
	}

	if req.Pattern =="" && req.PatternModule!=nil {
        normalized:=strings.ReplaceAll(strings.Trim(strings.ToLower(*req.PatternModule),"")," ","-")
		req.Pattern=fmt.Sprintf("%s-%s-*",constants.CUSTOM_INDEX_PREFIX,normalized)
	}


	p := &domain.UtmIndexPattern{
		Pattern:       req.Pattern,
		PatternModule: req.PatternModule,
		PatternSystem: &systemFalse, // always false in non-dev mode (no isInDevelop)
		IsActive:      &isActive,
	}

	if err := u.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return toResponse(p), nil
}

func (u *indexPatternUsecase) Update(ctx context.Context, req dto.UpdateIndexPatternRequest) (*dto.IndexPatternResponse, error) {
	p := &domain.UtmIndexPattern{
		ID:            req.ID,
		Pattern:       req.Pattern,
		PatternModule: req.PatternModule,
		PatternSystem: req.PatternSystem,
		IsActive:      req.IsActive,
	}
	if err := u.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return toResponse(p), nil
}

func (u *indexPatternUsecase) GetByID(ctx context.Context, id int64) (*dto.IndexPatternResponse, error) {
	p, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil
	}
	return toResponse(p), nil
}

func (u *indexPatternUsecase) List(ctx context.Context, f dto.IndexPatternFilters) ([]dto.IndexPatternResponse, int64, error) {
	items, total, err := u.repo.List(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]dto.IndexPatternResponse, 0, len(items))
	for i := range items {
		resp = append(resp, *toResponse(&items[i]))
	}
	return resp, total, nil
}

func (u *indexPatternUsecase) Fields(ctx context.Context, f dto.IndexPatternFilters) ([]dto.IndexPatternFieldsResponse, int64, error) {
	items, total, err := u.repo.List(ctx, f)
	if err != nil {
		return nil, 0, err
	}

	out := make([]dto.IndexPatternFieldsResponse, 0, len(items))

	for i := range items {
		entry := dto.IndexPatternFieldsResponse{
			IndexPattern: *toResponse(&items[i]),
		}

		if u.ismRepo != nil {
			fields, ferr := u.ismRepo.GetIndexProperties(ctx, items[i].Pattern)
			if ferr != nil {
				catcher.Warn("opensearch: GetIndexProperties("+items[i].Pattern+"): "+ferr.Error(), nil)
				entry.Fields = nil
			} else {
				entry.Fields = fields
			}
		} else {
			entry.Fields = []dto.IndexField{}
		}

		out = append(out, entry)
	}

	return out, total, nil
}

func (u *indexPatternUsecase) Delete(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}

func toResponse(p *domain.UtmIndexPattern) *dto.IndexPatternResponse {
	return &dto.IndexPatternResponse{
		ID:            p.ID,
		Pattern:       p.Pattern,
		PatternModule: p.PatternModule,
		PatternSystem: p.PatternSystem,
		IsActive:      p.IsActive,
	}
}
