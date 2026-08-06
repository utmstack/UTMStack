package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

func registerIntegrations(m *Module) {
	registerIntegrationsModules(m)
	registerIntegrationsTenants(m)
}

// ---- integrations.* --------------------------------------------------------

type integrationsListInput struct {
	Category     string `json:"category,omitempty"`
	Active       *bool  `json:"active,omitempty"`
	NameContains string `json:"name_contains,omitempty"`
	Page         int    `json:"page,omitempty"`
	Size         int    `json:"size,omitempty"`
}

type integrationsIDInput struct {
	ID int64 `json:"id"`
}

type integrationsNameInput struct {
	ModuleName string `json:"module_name"`
}

type integrationsActivateInput struct {
	ModuleName string `json:"module_name"`
	Active     bool   `json:"active"`
}

type integrationsCreateInput struct {
	ModuleName        string `json:"module_name"`
	DataType          string `json:"data_type"`
	PrettyName        string `json:"pretty_name,omitempty"`
	ModuleDescription string `json:"module_description,omitempty"`
	ModuleIcon        string `json:"module_icon,omitempty"`
	ModuleCategory    string `json:"module_category,omitempty"`
}

type integrationsUpdateInput struct {
	ID                int64  `json:"id"`
	PrettyName        string `json:"pretty_name,omitempty"`
	ModuleDescription string `json:"module_description,omitempty"`
	ModuleIcon        string `json:"module_icon,omitempty"`
	ModuleCategory    string `json:"module_category,omitempty"`
}

func registerIntegrationsModules(m *Module) {
	uc := m.deps.Integrations.Modules()

	Add(m, &mcp.Tool{
		Name: "integrations.list", Title: "List integrations",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "integrations.read"},
		func(ctx context.Context, _ *authz.Actor, in integrationsListInput) (any, error) {
			f := connectors.ModuleListFilter{
				Params: database.Params{Page: in.Page, Size: clampPageSize(in.Size)},
			}
			if in.Category != "" {
				c := in.Category
				f.ModuleCategory = &c
			}
			if in.Active != nil {
				f.ModuleActive = in.Active
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
			return uc.GetByName(ctx, in.ModuleName)
		})

	Add(m, &mcp.Tool{
		Name: "integrations.create", Title: "Create integration",
	}, Gate{Permission: "integrations.write"},
		func(ctx context.Context, _ *authz.Actor, in integrationsCreateInput) (any, error) {
			return uc.Create(ctx, dto.CreateModuleRequest{
				ModuleName: in.ModuleName, DataType: in.DataType,
				PrettyName: in.PrettyName, ModuleDescription: in.ModuleDescription,
				ModuleIcon: in.ModuleIcon, ModuleCategory: in.ModuleCategory,
			})
		})

	Add(m, &mcp.Tool{
		Name: "integrations.update", Title: "Update integration",
	}, Gate{Permission: "integrations.write"},
		func(ctx context.Context, _ *authz.Actor, in integrationsUpdateInput) (any, error) {
			return uc.Update(ctx, in.ID, dto.UpdateModuleRequest{
				PrettyName: in.PrettyName, ModuleDescription: in.ModuleDescription,
				ModuleIcon: in.ModuleIcon, ModuleCategory: in.ModuleCategory,
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
		Name: "integrations.activate", Title: "Activate or deactivate integration",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "integrations.write"},
		func(ctx context.Context, _ *authz.Actor, in integrationsActivateInput) (any, error) {
			active := in.Active
			return uc.ActivateDeactivate(ctx, dto.ModuleActivationRequest{
				ModuleName: in.ModuleName, ActivationStatus: &active,
			})
		})

	Add(m, &mcp.Tool{
		Name: "integrations.is_active", Title: "Check integration activation",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "integrations.read"},
		func(ctx context.Context, _ *authz.Actor, in integrationsNameInput) (any, error) {
			ok, err := uc.IsActive(ctx, in.ModuleName)
			if err != nil {
				return nil, err
			}
			return map[string]any{"active": ok}, nil
		})

	Add(m, &mcp.Tool{
		Name: "integrations.categories", Title: "List categories",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "integrations.read"},
		func(ctx context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			return uc.Categories(ctx)
		})

	Add(m, &mcp.Tool{
		Name: "integrations.data_types", Title: "List data type options",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "integrations.read"},
		func(ctx context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			return uc.DataTypes(ctx)
		})
}

// ---- integrations.tenants.* ------------------------------------------------

type integrationsTenantListInput struct {
	Module string `json:"module"`
}

type integrationsTenantSaveInput struct {
	Module string            `json:"module"`
	Name   string            `json:"name"`
	Config map[string]string `json:"config,omitempty"`
}

type integrationsTenantDeleteInput struct {
	Module string `json:"module"`
	Name   string `json:"name"`
}

func registerIntegrationsTenants(m *Module) {
	uc := m.deps.Integrations.Tenants()

	Add(m, &mcp.Tool{
		Name: "integrations.tenants.list", Title: "List integration tenant configs",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "integrations.read"},
		func(ctx context.Context, _ *authz.Actor, in integrationsTenantListInput) (any, error) {
			return uc.List(ctx, in.Module)
		})

	Add(m, &mcp.Tool{
		Name: "integrations.tenants.save", Title: "Create or update integration tenant",
	}, Gate{Permission: "integrations.write"},
		func(ctx context.Context, _ *authz.Actor, in integrationsTenantSaveInput) (any, error) {
			if err := uc.Save(ctx, in.Module, dto.TenantRequest{Name: in.Name, Config: in.Config}); err != nil {
				return nil, err
			}
			return map[string]any{"module": in.Module, "name": in.Name, "saved": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "integrations.tenants.delete", Title: "Delete integration tenant",
	}, Gate{Permission: "integrations.write"},
		func(ctx context.Context, _ *authz.Actor, in integrationsTenantDeleteInput) (any, error) {
			if err := uc.Delete(ctx, in.Module, in.Name); err != nil {
				return nil, err
			}
			return map[string]any{"module": in.Module, "name": in.Name, "deleted": true}, nil
		})
}
