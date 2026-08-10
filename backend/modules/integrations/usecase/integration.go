package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"github.com/utmstack/utmstack/backend/modules/integrations/dto"
)

type integrationUsecase struct {
	repo   connectors.IntegrationRepository
	config connectors.ConfigRepository
}

func NewIntegrationUsecase(repo connectors.IntegrationRepository, config connectors.ConfigRepository) connectors.IntegrationUsecase {
	return &integrationUsecase{repo: repo, config: config}
}

// configured answers whether the caller's tenant has set this integration up.
// Only pullers keep their configuration here; everything else is configured on
// the device or the agent, where this service cannot see it.
func (u *integrationUsecase) configured(ctx context.Context, i domain.Integration) bool {
	if i.IngestType != domain.IngestTypePlugin {
		return false
	}
	groups, err := u.config.Load(ctx, i.Name)
	if err != nil {
		return false
	}
	return len(groups) > 0
}

func (u *integrationUsecase) toResponse(ctx context.Context, i domain.Integration) dto.IntegrationResponse {
	resp := dto.FromIntegration(i)
	resp.Configured = u.configured(ctx, i)
	return resp
}

var _ connectors.IntegrationUsecase = (*integrationUsecase)(nil)

func (u *integrationUsecase) Create(ctx context.Context, req dto.CreateIntegrationRequest) (*dto.IntegrationResponse, error) {
	if !req.IngestType.Valid() {
		return nil, fmt.Errorf("%w: %q", domain.ErrInvalidIngestType, req.IngestType)
	}

	i := &domain.Integration{
		Name:        req.Name,
		DataType:    req.DataType,
		IngestType:  req.IngestType,
		Description: req.Description,
		Icon:        req.Icon,
	}
	if err := u.repo.Save(ctx, i); err != nil {
		return nil, err
	}

	resp := u.toResponse(ctx, *i)
	return &resp, nil
}

func (u *integrationUsecase) Update(ctx context.Context, id uuid.UUID, req dto.UpdateIntegrationRequest) (*dto.IntegrationResponse, error) {
	i, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if i.SystemOwner {
		return nil, domain.ErrSystemIntegration
	}

	i.Description = req.Description
	i.Icon = req.Icon
	if err := u.repo.Save(ctx, i); err != nil {
		return nil, err
	}

	resp := u.toResponse(ctx, *i)
	return &resp, nil
}

func (u *integrationUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	i, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if i.SystemOwner {
		return domain.ErrSystemIntegration
	}
	return u.repo.Delete(ctx, id)
}

func (u *integrationUsecase) List(ctx context.Context, filter connectors.IntegrationListFilter) ([]dto.IntegrationResponse, int64, error) {
	items, total, err := u.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.IntegrationResponse, 0, len(items))
	for _, i := range items {
		out = append(out, u.toResponse(ctx, i))
	}
	return out, total, nil
}

func (u *integrationUsecase) GetByID(ctx context.Context, id uuid.UUID) (*dto.IntegrationResponse, error) {
	i, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := u.toResponse(ctx, *i)
	return &resp, nil
}

func (u *integrationUsecase) GetByName(ctx context.Context, name string) (*dto.IntegrationResponse, error) {
	i, err := u.repo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	resp := u.toResponse(ctx, *i)
	return &resp, nil
}

func (u *integrationUsecase) DataTypes(ctx context.Context) ([]dto.DataTypeOption, error) {
	items, err := u.repo.DataTypes(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.DataTypeOption, 0, len(items))
	for _, i := range items {
		out = append(out, dto.DataTypeOption{
			DataType:    i.DataType,
			Name:        i.Name,
			SystemOwner: i.SystemOwner,
		})
	}
	return out, nil
}
