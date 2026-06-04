package usecase

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/correlation/connectors"
	"github.com/utmstack/utmstack/backend/modules/correlation/domain"
	"github.com/utmstack/utmstack/backend/modules/correlation/dto"
	correrrors "github.com/utmstack/utmstack/backend/modules/correlation/errors"
)

type tenantConfigUsecase struct {
	repo connectors.TenantConfigRepository
}

func NewTenantConfigUsecase(repo connectors.TenantConfigRepository) connectors.TenantConfigUsecase {
	return &tenantConfigUsecase{repo: repo}
}

func (u *tenantConfigUsecase) Create(ctx context.Context, req dto.CreateTenantConfigRequest) (*dto.TenantConfigResponse, error) {
	if req.ID != nil {
		return nil, correrrors.ErrIDMustBeAbsent
	}
	now := time.Now().UTC()
	t := &domain.UtmTenantConfig{
		AssetName:            req.AssetName,
		AssetHostnameListDef: stringOrEmpty(req.AssetHostnameListDef),
		AssetIpListDef:       stringOrEmpty(req.AssetIpListDef),
		AssetConfidentiality: req.AssetConfidentiality,
		AssetIntegrity:       req.AssetIntegrity,
		AssetAvailability:    req.AssetAvailability,
		LastUpdate:           &now,
	}
	saved, err := u.repo.Create(ctx, t)
	if err != nil {
		return nil, err
	}
	return dto.TenantConfigToResponse(saved), nil
}

func (u *tenantConfigUsecase) Update(ctx context.Context, req dto.UpdateTenantConfigRequest) (*dto.TenantConfigResponse, error) {
	if req.ID == nil || *req.ID == 0 {
		return nil, correrrors.ErrIDRequired
	}
	existing, err := u.repo.GetByID(ctx, *req.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, correrrors.ErrTenantConfigNotFound
	}
	now := time.Now().UTC()
	existing.AssetName = req.AssetName
	existing.AssetHostnameListDef = stringOrEmpty(req.AssetHostnameListDef)
	existing.AssetIpListDef = stringOrEmpty(req.AssetIpListDef)
	existing.AssetConfidentiality = req.AssetConfidentiality
	existing.AssetIntegrity = req.AssetIntegrity
	existing.AssetAvailability = req.AssetAvailability
	existing.LastUpdate = &now

	saved, err := u.repo.Update(ctx, existing)
	if err != nil {
		return nil, err
	}
	return dto.TenantConfigToResponse(saved), nil
}

func (u *tenantConfigUsecase) GetByID(ctx context.Context, id int64) (*dto.TenantConfigResponse, error) {
	t, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, correrrors.ErrTenantConfigNotFound
	}
	return dto.TenantConfigToResponse(t), nil
}

func (u *tenantConfigUsecase) List(ctx context.Context, f dto.TenantConfigFilters) (*connectors.ListResult[dto.TenantConfigResponse], error) {
	page, size := normPage(f.Page, f.Size)
	items, total, err := u.repo.List(ctx, connectors.TenantConfigFilters{
		Search: f.Search,
		Page:   page,
		Size:   size,
	})
	if err != nil {
		return nil, err
	}
	responses := make([]dto.TenantConfigResponse, len(items))
	for i := range items {
		responses[i] = *dto.TenantConfigToResponse(&items[i])
	}
	return &connectors.ListResult[dto.TenantConfigResponse]{Items: responses, Total: total}, nil
}

func (u *tenantConfigUsecase) Delete(ctx context.Context, id int64) error {
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return correrrors.ErrTenantConfigNotFound
	}
	return u.repo.Delete(ctx, id)
}

func stringOrEmpty(r []byte) string {
	if len(r) == 0 {
		return ""
	}
	return string(r)
}
