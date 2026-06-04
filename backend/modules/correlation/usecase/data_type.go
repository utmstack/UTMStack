package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/correlation/connectors"
	"github.com/utmstack/utmstack/backend/modules/correlation/domain"
	"github.com/utmstack/utmstack/backend/modules/correlation/dto"
	correrrors "github.com/utmstack/utmstack/backend/modules/correlation/errors"
)

type dataTypeUsecase struct {
	repo        connectors.DataTypeRepository
	assetSyncer connectors.AssetSyncer
}

func NewDataTypeUsecase(repo connectors.DataTypeRepository, assetSyncer connectors.AssetSyncer) connectors.DataTypeUsecase {
	return &dataTypeUsecase{repo: repo, assetSyncer: assetSyncer}
}

func (u *dataTypeUsecase) Create(ctx context.Context, req dto.CreateDataTypeRequest) (*dto.DataTypeResponse, error) {
	if req.ID != nil {
		return nil, correrrors.ErrIDMustBeAbsent
	}
	dt := &domain.UtmDataTypes{
		DataType:            req.DataType,
		DataTypeName:        req.DataTypeName,
		DataTypeDescription: req.DataTypeDescription,
		Included:            req.Included,
		SystemOwner:         false,
	}
	saved, err := u.repo.Create(ctx, dt)
	if err != nil {
		return nil, err
	}
	return dto.DataTypeToResponse(saved), nil
}

func (u *dataTypeUsecase) Update(ctx context.Context, req dto.UpdateDataTypeRequest) (*dto.DataTypeResponse, error) {
	if req.ID == nil || *req.ID == 0 {
		return nil, correrrors.ErrIDRequired
	}
	existing, err := u.repo.GetByID(ctx, *req.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, correrrors.ErrDataTypeNotFound
	}
	if existing.SystemOwner {
		return nil, correrrors.ErrDataTypeSystemOwner
	}
	existing.DataType = req.DataType
	existing.DataTypeName = req.DataTypeName
	existing.DataTypeDescription = req.DataTypeDescription
	existing.Included = req.Included

	saved, err := u.repo.Update(ctx, existing)
	if err != nil {
		return nil, err
	}
	return dto.DataTypeToResponse(saved), nil
}

func (u *dataTypeUsecase) GetByID(ctx context.Context, id int64) (*dto.DataTypeResponse, error) {
	dt, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dt == nil {
		return nil, correrrors.ErrDataTypeNotFound
	}
	return dto.DataTypeToResponse(dt), nil
}

func (u *dataTypeUsecase) List(ctx context.Context, f dto.DataTypeFilters) (*connectors.ListResult[dto.DataTypeResponse], error) {
	page, size := normPage(f.Page, f.Size)
	items, total, err := u.repo.List(ctx, connectors.DataTypeFilters{
		Search: f.Search,
		Page:   page,
		Size:   size,
	})
	if err != nil {
		return nil, err
	}
	responses := make([]dto.DataTypeResponse, len(items))
	for i := range items {
		responses[i] = *dto.DataTypeToResponse(&items[i])
	}
	return &connectors.ListResult[dto.DataTypeResponse]{Items: responses, Total: total}, nil
}

func (u *dataTypeUsecase) Delete(ctx context.Context, id int64) error {
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return correrrors.ErrDataTypeNotFound
	}
	if existing.SystemOwner {
		return correrrors.ErrDataTypeSystemOwner
	}
	return u.repo.Delete(ctx, id)
}

// TODO(module-33): steps B+C require utm_network_scan — implement when module #33 is ported.
func (u *dataTypeUsecase) UpdateIncludeExcludeList(ctx context.Context, items []dto.UpdateIncludeExcludeItem) error {
	// Filter out items with null id.
	valid := make([]dto.UpdateIncludeExcludeItem, 0, len(items))
	for _, item := range items {
		if item.ID != nil {
			valid = append(valid, item)
		}
	}
	if len(valid) == 0 {
		return nil
	}

	// Step A: for each valid item, fetch from DB and update only the included flag.
	updated := make([]domain.UtmDataTypes, 0, len(valid))
	for _, item := range valid {
		existing, err := u.repo.GetByID(ctx, *item.ID)
		if err != nil {
			return err
		}
		if existing == nil {
			continue // silently skip missing records (matches Java behavior)
		}
		existing.Included = item.Included
		saved, err := u.repo.Update(ctx, existing)
		if err != nil {
			return err
		}
		updated = append(updated, *saved)
	}

	// Collect excluded data type names for asset cleanup.
	excludedTypes := make([]string, 0, len(updated))
	for _, dt := range updated {
		if !dt.Included {
			excludedTypes = append(excludedTypes, dt.DataType)
		}
	}

	// TODO(module-33): steps B+C require utm_network_scan — implement when module #33 is ported.
	if err := u.assetSyncer.DeleteAllAssetsByDataType(ctx, excludedTypes); err != nil {
		return err
	}

	// TODO(module-33): implement when utm_network_scan is ported.
	return u.assetSyncer.SynchronizeSourcesToAssets(ctx)
}
