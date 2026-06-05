package usecase

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
	"github.com/utmstack/utmstack/backend/modules/network_scan/dto"
)

type assetGroupUsecase struct {
	repo connectors.AssetGroupRepository
}

func NewAssetGroupUsecase(repo connectors.AssetGroupRepository) connectors.AssetGroupUsecase {
	return &assetGroupUsecase{repo: repo}
}

func (u *assetGroupUsecase) Create(ctx context.Context, g *domain.UtmAssetGroup) (*domain.UtmAssetGroup, error) {
	if g.CreatedDate.IsZero() {
		g.CreatedDate = time.Now().UTC()
	}
	if err := u.repo.Save(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (u *assetGroupUsecase) Update(ctx context.Context, g *domain.UtmAssetGroup) (*domain.UtmAssetGroup, error) {
	if err := u.repo.Save(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (u *assetGroupUsecase) GetByID(ctx context.Context, id uint64) (*domain.UtmAssetGroup, error) {
	g, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, domain.ErrNotFound
	}
	return g, nil
}

func (u *assetGroupUsecase) SearchByFilter(
	ctx context.Context,
	f domain.AssetGroupFilter,
	p domain.Pagination,
) (*dto.AssetGroupListResponse, error) {
	rows, total, err := u.repo.SearchByFilter(ctx, f, p)
	if err != nil {
		return nil, err
	}
	page, size := p.Page, p.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(size) - 1) / int64(size))
	}
	data := make([]dto.AssetGroupDTO, 0, len(rows))
	for _, r := range rows {
		g := r.Group
		data = append(data, dto.ToAssetGroupDTO(&g, r.AssetsCount))
	}
	return &dto.AssetGroupListResponse{
		Data: data,
		PageInfo: dto.PageInfo{
			Page:       page,
			PageSize:   size,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}, nil
}

func (u *assetGroupUsecase) Delete(ctx context.Context, id uint64) error {
	return u.repo.Delete(ctx, id)
}
