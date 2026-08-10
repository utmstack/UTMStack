package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/threatwinds/go-sdk/catcher"

	ds_connectors "github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	ds_domain "github.com/utmstack/utmstack/backend/modules/datasources/domain"
	ds_dto "github.com/utmstack/utmstack/backend/modules/datasources/dto"
	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"github.com/utmstack/utmstack/backend/modules/integrations/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

type ConfigGroupUsecase struct {
	repo         connectors.ConfigRepository
	schema       connectors.SchemaProvider
	verifier     connectors.CredentialVerifier
	cipher       connectors.Cipher
	integrations connectors.IntegrationRepository
	datasources  ds_connectors.DatasourceUsecase
}

func NewConfigGroupUsecase(
	repo connectors.ConfigRepository,
	schema connectors.SchemaProvider,
	verifier connectors.CredentialVerifier,
	cipher connectors.Cipher,
	integrations connectors.IntegrationRepository,
	datasources ds_connectors.DatasourceUsecase,
) *ConfigGroupUsecase {
	return &ConfigGroupUsecase{
		repo:         repo,
		schema:       schema,
		verifier:     verifier,
		cipher:       cipher,
		integrations: integrations,
		datasources:  datasources,
	}
}

func (u *ConfigGroupUsecase) Save(ctx context.Context, integration string, req dto.ConfigGroupRequest) error {
	schema, err := u.schema.Schema(integration)
	if err != nil {
		return err
	}

	plain, err := u.resolveMasked(ctx, integration, req.Name, schema, req.Config)
	if err != nil {
		return err
	}

	if err := u.verifier.Verify(integration, plain); err != nil {
		return fmt.Errorf("credential verification failed: %w", err)
	}

	enc, err := u.encryptSensitive(schema, plain)
	if err != nil {
		return err
	}

	group := domain.ConfigGroup{Name: req.Name, Description: req.Description, Config: enc}
	if err := u.repo.Upsert(ctx, integration, group); err != nil {
		return err
	}

	if err := u.registerDatasource(ctx, integration, req.Name); err != nil {
		_ = catcher.Error("integrations: failed to register datasource for config group", err,
			map[string]any{"integration": integration, "group": req.Name})
	}
	return nil
}

func (u *ConfigGroupUsecase) Delete(ctx context.Context, integration, groupName string) error {
	return u.repo.Delete(ctx, integration, groupName)
}

func (u *ConfigGroupUsecase) List(ctx context.Context, integration string) ([]dto.ConfigGroupResponse, error) {
	groups, err := u.repo.Load(ctx, integration)
	if err != nil {
		return nil, err
	}

	schema, err := u.schema.Schema(integration)
	if errors.Is(err, domain.ErrNotConfigurable) {
		return []dto.ConfigGroupResponse{}, nil
	}
	if err != nil {
		return nil, err
	}

	out := make([]dto.ConfigGroupResponse, 0, len(groups))
	for _, g := range groups {
		cfg := make(map[string]string, len(g.Config))
		for k, v := range g.Config {
			cfg[k] = v
			if isSensitive(schema[k]) && v != "" {
				cfg[k] = maskedValue
			}
		}
		out = append(out, dto.ConfigGroupResponse{Name: g.Name, Description: g.Description, Config: cfg})
	}
	return out, nil
}

func (u *ConfigGroupUsecase) SyncDatasources(ctx context.Context) error {
	integrations, err := u.integrations.DataTypes(tenancy.WithAllTenantsRead(ctx))
	if err != nil {
		return err
	}

	for _, integration := range integrations {
		tenants, err := u.repo.LoadAllTenants(ctx, integration.Name)
		if err != nil {
			_ = catcher.Error("integrations: failed to load config groups for datasource sync", err,
				map[string]any{"integration": integration.Name})
			continue
		}

		for _, tc := range tenants {
			tenantCtx := authz.WithTenantID(ctx, tc.ID.String())
			for _, g := range tc.Groups {
				if err := u.datasources.Register(tenantCtx, ds_dto.RegisterRequest{
					Name:       g.Name,
					DataType:   integration.DataType,
					SourceKind: ds_domain.SourceKindPuller,
				}); err != nil {
					_ = catcher.Error("integrations: failed to register datasource", err,
						map[string]any{"integration": integration.Name, "group": g.Name, "tenant": tc.ID})
				}
			}
		}
	}
	return nil
}

func (u *ConfigGroupUsecase) registerDatasource(ctx context.Context, integration, groupName string) error {
	i, err := u.integrations.GetByName(ctx, integration)
	if err != nil {
		return err
	}
	return u.datasources.Register(ctx, ds_dto.RegisterRequest{
		Name:       groupName,
		DataType:   i.DataType,
		SourceKind: ds_domain.SourceKindPuller,
	})
}

func (u *ConfigGroupUsecase) encryptSensitive(schema, config map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(config))
	for k, v := range config {
		out[k] = v
		if isSensitive(schema[k]) && v != "" {
			enc, err := u.cipher.Encrypt(v)
			if err != nil {
				return nil, fmt.Errorf("encrypt %s: %w", k, err)
			}
			out[k] = enc
		}
	}
	return out, nil
}

func (u *ConfigGroupUsecase) resolveMasked(ctx context.Context, integration, groupName string, schema, config map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(config))
	var existing map[string]string

	for k, v := range config {
		out[k] = v
		if !isSensitive(schema[k]) || v != maskedValue {
			continue
		}
		if existing == nil {
			ex, err := u.existingConfig(ctx, integration, groupName)
			if err != nil {
				return nil, err
			}
			existing = ex
		}
		stored := existing[k]
		if stored == "" {
			return nil, fmt.Errorf("%w: %s", domain.ErrRequiredConfigEmpty, k)
		}
		dec, err := u.cipher.Decrypt(stored)
		if err != nil {
			return nil, fmt.Errorf("decrypt %s: %w", k, err)
		}
		out[k] = dec
	}
	return out, nil
}

func (u *ConfigGroupUsecase) existingConfig(ctx context.Context, integration, groupName string) (map[string]string, error) {
	groups, err := u.repo.Load(ctx, integration)
	if err != nil {
		return nil, err
	}
	for _, g := range groups {
		if g.Name == groupName {
			return g.Config, nil
		}
	}
	return nil, domain.ErrConfigGroupNotFound
}

const (
	confTypePassword = "password"
	confTypeFile     = "file"
	maskedValue      = "*****"
)

func isSensitive(confType string) bool {
	return confType == confTypePassword || confType == confTypeFile
}
