package usecase

import (
	"context"
	"fmt"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/dto"
	"github.com/utmstack/utmstack/backend/pkg/secret"
)

type configUsecase struct {
	configs   connectors.ConfigRepository
	modules   connectors.ModuleRepository
	factory   connectors.ModuleFactory
	cipher    *secret.Cipher
	eventProc connectors.EventProcessorClient
}

func NewConfigUsecase(
	configs connectors.ConfigRepository,
	modules connectors.ModuleRepository,
	factory connectors.ModuleFactory,
	cipher *secret.Cipher,
	eventProc connectors.EventProcessorClient,
) connectors.ConfigUsecase {
	return &configUsecase{
		configs:   configs,
		modules:   modules,
		factory:   factory,
		cipher:    cipher,
		eventProc: eventProc,
	}
}

func (u *configUsecase) ListByGroupID(ctx context.Context, groupID int64) ([]dto.GroupConfigurationItem, error) {
	rows, err := u.configs.ListByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.GroupConfigurationItem, 0, len(rows))
	for _, c := range rows {
		out = append(out, dto.FromConfiguration(c, false))
	}
	return out, nil
}

func (u *configUsecase) GetByGroupAndKey(ctx context.Context, groupID int64, confKey string) (*dto.GroupConfigurationItem, error) {
	c, err := u.configs.GetByGroupAndKey(ctx, groupID, confKey)
	if err != nil {
		return nil, err
	}
	item := dto.FromConfiguration(*c, false)
	return &item, nil
}

// Update implements the legacy three-pass write the panel relies on:
//
//  1. Skip incoming rows whose sensitive value is still the mask sentinel — the
//     user didn't touch that field, so the stored ciphertext stays put.
//  2. Reject required rows submitted blank.
//  3. Encrypt any new sensitive value before persisting.
//
// After writing, flip the parent module's needs_restart flag based on the
// kind-name allow-list, then publish the resulting state to event-processor.
func (u *configUsecase) Update(ctx context.Context, req dto.UpdateGroupConfigurationRequest) error {
	if len(req.Keys) == 0 {
		return fmt.Errorf("no configuration keys provided")
	}

	module, err := u.modules.GetByID(ctx, req.ModuleID)
	if err != nil {
		return err
	}

	toSave := make([]domain.UtmModuleGroupConfiguration, 0, len(req.Keys))
	for _, input := range req.Keys {
		sensitive := domain.IsSensitive(input.ConfDataType)
		if sensitive && input.ConfValue == domain.MaskedValue {
			continue
		}
		if input.ConfRequired && input.ConfValue == "" {
			return fmt.Errorf("%w: %s (%s)", domain.ErrRequiredConfigEmpty, input.ConfName, input.ConfKey)
		}
		row := input.ToConfiguration()
		if sensitive && row.ConfValue != "" {
			enc, eerr := u.cipher.Encrypt(row.ConfValue)
			if eerr != nil {
				return fmt.Errorf("encrypt %s: %w", row.ConfKey, eerr)
			}
			row.ConfValue = enc
		}
		toSave = append(toSave, row)
	}

	if err := u.configs.SaveAll(ctx, toSave); err != nil {
		return err
	}

	markNeedsRestart(module)
	if err := u.modules.Save(ctx, module); err != nil {
		return err
	}

	if u.eventProc != nil {
		fresh, ferr := u.modules.GetByID(ctx, module.ID)
		if ferr != nil {
			return ferr
		}
		if perr := u.eventProc.UpdateModule(ctx, fresh.ModuleName, dto.ToEventProcessorPayload(*fresh)); perr != nil {
			return fmt.Errorf("event-processor publish failed: %w", perr)
		}
	}
	return nil
}
