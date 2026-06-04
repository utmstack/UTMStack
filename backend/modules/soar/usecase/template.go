package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

type templateUsecase struct {
	templates connectors.TemplateRepository
}

func NewTemplateUsecase(templates connectors.TemplateRepository) connectors.TemplateUsecase {
	return &templateUsecase{templates: templates}
}

func (u *templateUsecase) List(ctx context.Context, f dto.TemplateFilters) (*connectors.ListResult[dto.TemplateResponse], error) {
	page, size := normPage(f.Page, f.Size)
	items, total, err := u.templates.List(ctx, connectors.TemplateFilters{
		ID:          f.ID,
		Label:       f.Label,
		Description: f.Description,
		Command:     f.Command,
		SystemOwner: f.SystemOwner,
		Page:        page,
		Size:        size,
	})
	if err != nil {
		return nil, err
	}

	responses := make([]dto.TemplateResponse, len(items))
	for i, t := range items {
		responses[i] = dto.TemplateResponse{
			ID:          t.ID,
			Title:       t.Title,
			Description: t.Description,
			Command:     t.Command,
			SystemOwner: t.SystemOwner,
		}
	}

	return &connectors.ListResult[dto.TemplateResponse]{Items: responses, Total: total}, nil
}
