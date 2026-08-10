package mcp

import (
	"context"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"github.com/utmstack/utmstack/backend/modules/integrations/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

func registerIntegrations(m *Module) {
	registerIntegrationsCatalog(m)
	registerIntegrationsConfig(m)
}

// ---- integrations.* --------------------------------------------------------

type integrationsListInput struct {
	IngestType   string `json:"ingest_type,omitempty"`
	NameContains string `json:"name_contains,omitempty"`
	Page         int    `json:"page,omitempty"`
	Size         int    `json:"size,omitempty"`
}

type integrationsIDInput struct {
	ID uuid.UUID `json:"id"`
}

type integrationsNameInput struct {
	Name string `json:"name"`
}

type integrationsCreateInput struct {
	Name        string `json:"name"`
	DataType    string `json:"data_type"`
	IngestType  string `json:"ingest_type"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

type integrationsUpdateInput struct {
	ID          uuid.UUID `json:"id"`
	Description string    `json:"description,omitempty"`
	Icon        string    `json:"icon,omitempty"`
}

func registerIntegrationsCatalog(m *Module) {
	uc := m.deps.Integrations.Integrations()

	Add(m, &mcp.Tool{
		Name: "integrations.list", Title: "List integrations",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "integrations.read"},
		func(ctx context.Context, _ *authz.Actor, in integrationsListInput) (any, error) {
			f := connectors.IntegrationListFilter{
				Params: database.Params{Page: in.Page, Size: clampPageSize(in.Size)},
			}
			if in.IngestType != "" {
				it := domain.IngestType(in.IngestType)
				if !it.Valid() {
					return nil, domain.ErrInvalidIngestType
				}
				f.IngestType = &it
			}
			if in.NameContains != "" {
				n := in.NameContains
				f.NameContains = &n
			}
			items, total, err := uc.List(ctx, f)
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "integrations.get", Title: "Get integration",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "integrations.read"},
		func(ctx context.Context, _ *authz.Actor, in integrationsIDInput) (any, error) {
			return uc.GetByID(ctx, in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "integrations.get_by_name", Title: "Get integration by name",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "integrations.read"},
		func(ctx context.Context, _ *authz.Actor, in integrationsNameInput) (any, error) {
			return uc.GetByName(ctx, in.Name)
		})

	Add(m, &mcp.Tool{
		Name: "integrations.create", Title: "Create integration",
	}, Gate{Permission: "integrations.write"},
		func(ctx context.Context, _ *authz.Actor, in integrationsCreateInput) (any, error) {
			return uc.Create(ctx, dto.CreateIntegrationRequest{
				Name:        in.Name,
				DataType:    in.DataType,
				IngestType:  domain.IngestType(in.IngestType),
				Description: in.Description,
				Icon:        in.Icon,
			})
		})

	Add(m, &mcp.Tool{
		Name: "integrations.update", Title: "Update integration",
	}, Gate{Permission: "integrations.write"},
		func(ctx context.Context, _ *authz.Actor, in integrationsUpdateInput) (any, error) {
			return uc.Update(ctx, in.ID, dto.UpdateIntegrationRequest{
				Description: in.Description,
				Icon:        in.Icon,
			})
		})

	Add(m, &mcp.Tool{
		Name: "integrations.delete", Title: "Delete integration",
	}, Gate{Permission: "integrations.write"},
		func(ctx context.Context, _ *authz.Actor, in integrationsIDInput) (any, error) {
			if err := uc.Delete(ctx, in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "integrations.data_types", Title: "List data type options",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "integrations.read"},
		func(ctx context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			return uc.DataTypes(ctx)
		})
}

// ---- integrations.config.* -------------------------------------------------

type integrationsConfigListInput struct {
	Integration string `json:"integration"`
}

type integrationsConfigSaveInput struct {
	Integration string            `json:"integration"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Config      map[string]string `json:"config,omitempty"`
}

type integrationsConfigDeleteInput struct {
	Integration string `json:"integration"`
	Name        string `json:"name"`
}

func registerIntegrationsConfig(m *Module) {
	uc := m.deps.Integrations.Groups()

	Add(m, &mcp.Tool{
		Name: "integrations.config.list", Title: "List integration configuration groups",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "integrations.read"},
		func(ctx context.Context, _ *authz.Actor, in integrationsConfigListInput) (any, error) {
			return uc.List(ctx, in.Integration)
		})

	Add(m, &mcp.Tool{
		Name: "integrations.config.save", Title: "Create or update a configuration group",
	}, Gate{Permission: "integrations.write"},
		func(ctx context.Context, _ *authz.Actor, in integrationsConfigSaveInput) (any, error) {
			req := dto.ConfigGroupRequest{Name: in.Name, Description: in.Description, Config: in.Config}
			if err := uc.Save(ctx, in.Integration, req); err != nil {
				return nil, err
			}
			return map[string]any{"integration": in.Integration, "name": in.Name, "saved": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "integrations.config.delete", Title: "Delete a configuration group",
	}, Gate{Permission: "integrations.write"},
		func(ctx context.Context, _ *authz.Actor, in integrationsConfigDeleteInput) (any, error) {
			if err := uc.Delete(ctx, in.Integration, in.Name); err != nil {
				return nil, err
			}
			return map[string]any{"integration": in.Integration, "name": in.Name, "deleted": true}, nil
		})
}
