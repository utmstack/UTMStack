package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
	"github.com/utmstack/utmstack/backend/modules/network_scan/dto"
)

type networkScanUsecase struct {
	repo       connectors.NetworkScanRepository
	portsRepo  connectors.PortsRepository
	agent      connectors.AgentGateway
	dataInputs connectors.DataInputStatusGateway
}

func NewNetworkScanUsecase(
	repo connectors.NetworkScanRepository,
	portsRepo connectors.PortsRepository,
	agentGW connectors.AgentGateway,
	dataInputsGW connectors.DataInputStatusGateway,
) connectors.NetworkScanUsecase {
	return &networkScanUsecase{
		repo:       repo,
		portsRepo:  portsRepo,
		agent:      agentGW,
		dataInputs: dataInputsGW,
	}
}

func (u *networkScanUsecase) SaveOrUpdateCustomAsset(ctx context.Context, input dto.NetworkScanDTO) error {
	if strings.TrimSpace(input.AssetName) == "" {
		return domain.ErrInvalidInput
	}
	now := time.Now().UTC()
	e := dto.FromNetworkScanDTO(input)
	e.RegisteredMode = string(domain.AssetRegisteredModeCustom)
	e.ModifiedAt = &now
	if e.ID == 0 {
		e.DiscoveredAt = &now
		if e.AssetStatus == "" {
			e.AssetStatus = string(domain.AssetStatusNew)
		}
	}
	return u.repo.Save(ctx, e)
}

func (u *networkScanUsecase) Update(ctx context.Context, scan *domain.UtmNetworkScan) (*domain.UtmNetworkScan, error) {
	now := time.Now().UTC()
	scan.ModifiedAt = &now
	if err := u.repo.Save(ctx, scan); err != nil {
		return nil, err
	}
	return scan, nil
}

func (u *networkScanUsecase) UpdateType(ctx context.Context, req dto.UpdateTypeRequest) error {
	return u.repo.UpdateType(ctx, req.AssetIDs, req.AssetTypeID)
}

func (u *networkScanUsecase) UpdateGroup(ctx context.Context, req dto.UpdateGroupRequest) error {
	return u.repo.UpdateGroup(ctx, req.AssetIDs, req.AssetGroupID)
}

func (u *networkScanUsecase) ListByCriteria(
	ctx context.Context,
	c domain.NetworkScanCriteria,
	p domain.Pagination,
) ([]domain.UtmNetworkScan, int64, error) {
	return u.repo.ListByCriteria(ctx, c, p)
}

func (u *networkScanUsecase) CountByCriteria(ctx context.Context, c domain.NetworkScanCriteria) (int64, error) {
	return u.repo.CountByCriteria(ctx, c)
}

func (u *networkScanUsecase) GetByID(ctx context.Context, id uint64) (*dto.NetworkScanDTO, error) {
	e, err := u.repo.FindByIDWithDetails(ctx, id)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, domain.ErrNotFound
	}
	out := dto.ToNetworkScanDTO(e, true, nil)
	return &out, nil
}

func (u *networkScanUsecase) SearchByFilters(
	ctx context.Context,
	f domain.NetworkScanFilter,
	p domain.Pagination,
) (*dto.NetworkScanListResponse, error) {
	items, total, err := u.repo.SearchByFilters(ctx, f, p)
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
	data := make([]dto.NetworkScanDTO, 0, len(items))
	for i := range items {
		data = append(data, dto.ToNetworkScanDTO(&items[i], false, nil))
	}
	return &dto.NetworkScanListResponse{
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

func (u *networkScanUsecase) SearchPropertyValues(
	ctx context.Context,
	propKey, value string,
	forGroups bool,
	p domain.Pagination,
) ([]string, error) {
	prop, ok := domain.LookupProperty(propKey)
	if !ok {
		return nil, domain.ErrUnknownProperty
	}
	return u.repo.SearchPropertyValues(ctx, prop, value, forGroups, p)
}

func (u *networkScanUsecase) CountNewAssets(ctx context.Context) (int64, error) {
	return u.repo.CountByStatus(ctx, domain.AssetStatusNew)
}

func (u *networkScanUsecase) DeleteCustomAsset(ctx context.Context, id uint64) error {
	e, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if e == nil {
		return domain.ErrNotFound
	}
	if e.IsAgent != nil && *e.IsAgent && u.agent != nil {
		if err := u.agent.DeleteAgentByName(ctx, e.AssetName); err != nil {
			// Java swallows AgentNotFoundException — propagate other errors.
			if !errors.Is(err, errAgentNotFound) {
				return err
			}
		}
	}
	if u.dataInputs != nil && strings.TrimSpace(e.AssetName) != "" {
		if err := u.dataInputs.DeleteByDataSource(ctx, e.AssetName); err != nil {
			return err
		}
	}
	if err := u.portsRepo.DeleteByScanID(ctx, e.ID); err != nil {
		return err
	}
	return u.repo.Delete(ctx, e.ID)
}

func (u *networkScanUsecase) CanRunCommand(ctx context.Context, assetName string) (bool, error) {
	e, err := u.repo.FindByAssetName(ctx, assetName)
	if err != nil {
		return false, err
	}
	if e == nil {
		return false, nil
	}
	return e.IsAgent != nil && *e.IsAgent, nil
}

func (u *networkScanUsecase) GetAgentsOSPlatform(ctx context.Context) ([]string, error) {
	return u.repo.FindAgentsOSPlatforms(ctx)
}

func (u *networkScanUsecase) AssetGroupsForAlerts(ctx context.Context) ([]domain.AssetGroupMapping, error) {
	return u.repo.FindAllAssetGroupMappings(ctx)
}

// errAgentNotFound is returned by agentGateway implementations when DeleteAgentByName
// cannot find the agent. The Java side swallows the equivalent exception so we mirror it.
var errAgentNotFound = errors.New("agent not found")
