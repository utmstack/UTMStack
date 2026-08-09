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
)

type TenantUsecase struct {
	repo        connectors.TenantRepository
	schema      connectors.SchemaProvider
	verifier    connectors.CredentialVerifier
	cipher      connectors.Cipher
	modules     connectors.ModuleRepository
	datasources ds_connectors.DatasourceUsecase
}

func NewTenantUsecase(
	repo connectors.TenantRepository,
	schema connectors.SchemaProvider,
	verifier connectors.CredentialVerifier,
	cipher connectors.Cipher,
	modules connectors.ModuleRepository,
	datasources ds_connectors.DatasourceUsecase,
) *TenantUsecase {
	return &TenantUsecase{
		repo:        repo,
		schema:      schema,
		verifier:    verifier,
		cipher:      cipher,
		modules:     modules,
		datasources: datasources,
	}
}

// Save validates, verifies and persists a tenant. Incoming Config holds plaintext
// values from the form (or MaskedValue for a sensitive field the user left
// unchanged). Flow: resolve masked secrets to the stored plaintext → verify the
// credentials against the provider (fail fast, nothing written) → encrypt the
// sensitive fields → upsert the tenant → register its datasource.
func (u *TenantUsecase) Save(ctx context.Context, module string, req dto.TenantRequest) error {
	tenantId := authz.TenantIDFromContext(ctx)

	schema, err := u.schema.Schema(module)
	if err != nil {
		return err
	}

	plain, err := u.resolveMasked(module, req.Name, tenantId, schema, req.Config)
	if err != nil {
		return err
	}

	if err := u.verifier.Verify(module, plain); err != nil {
		return fmt.Errorf("credential verification failed: %w", err)
	}

	enc, err := u.encryptSensitive(schema, plain)
	if err != nil {
		return err
	}
	if err := u.repo.Upsert(module, domain.Tenant{Name: req.Name, TenantId: tenantId, Config: enc}); err != nil {
		return err
	}

	if err := u.registerDatasource(ctx, module, req.Name); err != nil {
		_ = catcher.Error("integrations: failed to register datasource for tenant", err,
			map[string]any{"module": module, "tenant": req.Name})
	}
	return nil
}

func (u *TenantUsecase) Delete(ctx context.Context, module, name string) error {
	return u.repo.Delete(module, name, authz.TenantIDFromContext(ctx))
}

func (u *TenantUsecase) SyncDatasources(ctx context.Context) error {
	mods, err := u.modules.DataTypes(ctx)
	if err != nil {
		return err
	}
	for _, m := range mods {
		tenants, err := u.repo.Load(m.ModuleName)
		if err != nil {
			_ = catcher.Error("integrations: failed to load tenants for datasource sync", err,
				map[string]any{"module": m.ModuleName})
			continue
		}
		for _, t := range tenants {
			if err := u.datasources.Register(ctx, ds_dto.RegisterRequest{
				Name:       t.Name,
				DataType:   m.DataType,
				SourceKind: ds_domain.SourceKindPuller,
			}); err != nil {
				_ = catcher.Error("integrations: failed to register datasource", err,
					map[string]any{"module": m.ModuleName, "tenant": t.Name})
			}
		}
	}
	return nil
}

func (u *TenantUsecase) registerDatasource(ctx context.Context, module, name string) error {
	m, err := u.modules.GetByName(ctx, module)
	if err != nil {
		return err
	}
	return u.datasources.Register(ctx, ds_dto.RegisterRequest{
		Name:       name,
		DataType:   m.DataType,
		SourceKind: ds_domain.SourceKindPuller,
	})
}

func (u *TenantUsecase) List(ctx context.Context, module string) ([]dto.TenantResponse, error) {
	return u.transform(module, authz.TenantIDFromContext(ctx))
}

func (u *TenantUsecase) transform(module, tenantId string) ([]dto.TenantResponse, error) {
	tenants, err := u.repo.Load(module)
	if err != nil {
		return nil, err
	}
	schema, err := u.schema.Schema(module)
	if errors.Is(err, domain.ErrNotConfigurable) {
		// Not a configurable integration (custom/syslog) → no tenants.
		return []dto.TenantResponse{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := make([]dto.TenantResponse, 0, len(tenants))
	for i := range tenants {
		// A SaaS-scoped actor only ever sees their own tenant's connector
		// instances. On-prem/global actors (tenantId == "") see everything,
		// matching legacy behavior.
		if tenantId != "" && tenants[i].TenantId != tenantId {
			continue
		}
		cfg := make(map[string]string, len(tenants[i].Config))
		for k, v := range tenants[i].Config {
			cfg[k] = v
			if !isSensitive(schema[k]) || v == "" {
				continue
			}
			cfg[k] = maskedValue
		}
		out = append(out, dto.TenantResponse{Name: tenants[i].Name, Config: cfg})
	}
	return out, nil
}

func (u *TenantUsecase) encryptSensitive(schema, config map[string]string) (map[string]string, error) {
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

// resolveMasked replaces any sensitive field submitted as MaskedValue with the
// currently-stored (decrypted) value — the user left that secret unchanged.
func (u *TenantUsecase) resolveMasked(module, name, tenantId string, schema, config map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(config))
	var existing map[string]string
	for k, v := range config {
		out[k] = v
		if !isSensitive(schema[k]) || v != maskedValue {
			continue
		}
		if existing == nil {
			ex, err := u.existingConfig(module, name, tenantId)
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

func (u *TenantUsecase) existingConfig(module, name, tenantId string) (map[string]string, error) {
	tenants, err := u.repo.Load(module)
	if err != nil {
		return nil, err
	}
	for _, t := range tenants {
		if t.Name == name && t.TenantId == tenantId {
			return t.Config, nil
		}
	}
	return nil, domain.ErrTenantNotFound
}

const (
	confTypePassword = "password"
	confTypeFile     = "file"
	maskedValue      = "*****"
)

func isSensitive(confType string) bool {
	return confType == confTypePassword || confType == confTypeFile
}
