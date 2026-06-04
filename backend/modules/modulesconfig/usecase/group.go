package usecase

import (
	"context"
	"fmt"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/dto"
	"github.com/utmstack/utmstack/backend/pkg/secret"
)

type groupUsecase struct {
	groups    connectors.GroupRepository
	configs   connectors.ConfigRepository
	modules   connectors.ModuleRepository
	factory   connectors.ModuleFactory
	cipher    *secret.Cipher
	eventProc connectors.EventProcessorClient
}

func NewGroupUsecase(
	groups connectors.GroupRepository,
	configs connectors.ConfigRepository,
	modules connectors.ModuleRepository,
	factory connectors.ModuleFactory,
	cipher *secret.Cipher,
	eventProc connectors.EventProcessorClient,
) connectors.GroupUsecase {
	return &groupUsecase{
		groups:    groups,
		configs:   configs,
		modules:   modules,
		factory:   factory,
		cipher:    cipher,
		eventProc: eventProc,
	}
}

// Create inserts the group row, then asks the registered ModuleKind for its
// default configuration keys and persists them. If the kind is not yet ported
// the caller gets ErrModuleKindNotPorted so the panel can surface an
// actionable message (rather than silently creating an empty group).
func (u *groupUsecase) Create(ctx context.Context, req dto.CreateModuleGroupRequest) (*dto.ModuleGroupResponse, error) {
	module, err := u.modules.GetByID(ctx, req.ModuleID)
	if err != nil {
		return nil, err
	}

	kind, ok := u.factory.Get(module.ModuleName)
	if !ok {
		return nil, fmt.Errorf("%w: %s", domain.ErrModuleKindNotPorted, module.ModuleName)
	}

	group := &domain.UtmModuleGroup{
		ModuleID:         req.ModuleID,
		GroupName:        req.GroupName,
		GroupDescription: req.GroupDescription,
		Collector:        req.Collector,
	}
	if err := u.groups.Save(ctx, group); err != nil {
		return nil, err
	}

	defaults := kind.ConfigurationKeys(group.ID)
	if len(defaults) > 0 {
		rows := make([]domain.UtmModuleGroupConfiguration, 0, len(defaults))
		for _, k := range defaults {
			row := k.ToConfiguration(group.ID)
			if domain.IsSensitive(row.ConfDataType) && row.ConfValue != "" {
				enc, eerr := u.cipher.Encrypt(row.ConfValue)
				if eerr != nil {
					return nil, fmt.Errorf("encrypt default value for %s: %w", row.ConfKey, eerr)
				}
				row.ConfValue = enc
			}
			rows = append(rows, row)
		}
		if err := u.configs.SaveAll(ctx, rows); err != nil {
			return nil, err
		}
	}

	fresh, err := u.groups.GetByID(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	resp := dto.FromGroup(*fresh, false)
	return &resp, nil
}

// Update changes group metadata only. Config rows are owned by ConfigUsecase.
func (u *groupUsecase) Update(ctx context.Context, req dto.UpdateModuleGroupRequest) (*dto.ModuleGroupResponse, error) {
	existing, err := u.groups.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if req.GroupName != "" {
		existing.GroupName = req.GroupName
	}
	existing.GroupDescription = req.GroupDescription
	existing.Collector = req.Collector
	if err := u.groups.Save(ctx, existing); err != nil {
		return nil, err
	}
	resp := dto.FromGroup(*existing, false)
	return &resp, nil
}

func (u *groupUsecase) ListByModuleID(ctx context.Context, moduleID int64) ([]dto.ModuleGroupResponse, error) {
	rows, err := u.groups.ListByModuleID(ctx, moduleID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ModuleGroupResponse, 0, len(rows))
	for _, g := range rows {
		out = append(out, dto.FromGroup(g, false))
	}
	return out, nil
}

func (u *groupUsecase) GetByID(ctx context.Context, id int64) (*dto.ModuleGroupResponse, error) {
	g, err := u.groups.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := dto.FromGroup(*g, false)
	return &resp, nil
}

// Delete removes a group (cascading into its configurations via the FK) and
// re-publishes the parent module's now-shrunken state so event-processor stays
// in sync.
func (u *groupUsecase) Delete(ctx context.Context, id int64) error {
	g, err := u.groups.GetByID(ctx, id)
	if err != nil {
		return err
	}
	moduleID := g.ModuleID

	if err := u.groups.Delete(ctx, id); err != nil {
		return err
	}

	if u.eventProc == nil {
		return nil
	}
	module, err := u.modules.GetByID(ctx, moduleID)
	if err != nil {
		return err
	}
	if perr := u.eventProc.UpdateModule(ctx, module.ModuleName, dto.ToEventProcessorPayload(*module)); perr != nil {
		return fmt.Errorf("event-processor publish failed: %w", perr)
	}
	return nil
}
