package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	appconfigdto "github.com/utmstack/utmstack/backend/modules/appconfig/dto"
	compdomain "github.com/utmstack/utmstack/backend/modules/compliance/domain"
	epdto "github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
	tenantdom "github.com/utmstack/utmstack/backend/modules/tenant/domain"
	tenantdto "github.com/utmstack/utmstack/backend/modules/tenant/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

// Every tool in this file is SUPER-ADMIN-ONLY (platform admin — an admin of
// the default tenant, per authz.IsPlatform). Any regular tenant user calling
// these is rejected by the Platform gate. The tools apply one change across
// N selected tenants; system-owned resources refuse per tenant and land in
// the `failed` array while the rest of the loop continues.

// bulkSelector picks which tenants the call targets. When AllTenants is true
// TenantIDs is ignored and every ACTIVE tenant is enumerated (the default
// platform tenant is filtered out for singleton configs like SMTP/branding).
type bulkSelector struct {
	TenantIDs  []string `json:"tenant_ids,omitempty" jsonschema:"Explicit list of tenant UUIDs. Ignored when all_tenants=true."`
	AllTenants bool     `json:"all_tenants,omitempty" jsonschema:"When true, enumerate every ACTIVE tenant."`
}

func (s bulkSelector) toCommon() common_models.BulkTenantSelector {
	return common_models.BulkTenantSelector{TenantIDs: s.TenantIDs, AllTenants: s.AllTenants}
}

// resolveBulkTenants returns the tenant IDs a bulk call should touch.
// stripPlatform=true drops DefaultTenantID from AllTenants enumeration
// (used for SMTP + branding where the platform plane must not be silently
// overwritten).
func resolveBulkTenants(ctx context.Context, m *Module, sel bulkSelector, stripPlatform bool) ([]string, error) {
	if !sel.AllTenants {
		return sel.TenantIDs, nil
	}
	if m.deps == nil || m.deps.Tenant == nil {
		return nil, fmt.Errorf("tenant module not wired: cannot enumerate all tenants")
	}
	tenants, _, err := m.deps.Tenant.GetTenantUsecase().List(ctx, tenantdto.Filter{Size: 10000, Status: tenantdom.StatusActive})
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(tenants))
	for _, t := range tenants {
		id := t.ID.String()
		if stripPlatform && id == authz.DefaultTenantID {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// bulkLoop runs fn once per tenant, injecting the tenant into the ctx, and
// aggregates outcomes. Errors do NOT stop the loop — they are recorded per
// tenant so a partial-fail batch still returns a useful result.
func bulkLoop(ctx context.Context, tenantIDs []string, fn func(ctx context.Context) error) common_models.BulkResult {
	var result common_models.BulkResult
	for _, tid := range tenantIDs {
		tenantCtx := authz.WithTenantID(ctx, tid)
		result.Append(tid, fn(tenantCtx))
	}
	return result
}

// registerPlatformBulk registers every cross-tenant bulk tool. Called once
// from buildServer(); every tool here is gated with Platform: true.
func registerPlatformBulk(m *Module) {
	registerBulkPipelines(m)
	registerBulkCorrelationRules(m)
	registerBulkComplianceFrameworks(m)
	registerBulkComplianceControls(m)
	registerBulkSMTP(m)
	registerBulkBranding(m)
}

// ---- platform.pipeline.bulk.* ----------------------------------------------

type bulkPipelineUpsertInput struct {
	Selector bulkSelector `json:"selector"`
	RelPath  string       `json:"rel_path"`
	Content  string       `json:"content"`
}

type bulkPipelineRelPathInput struct {
	Selector bulkSelector `json:"selector"`
	RelPath  string       `json:"rel_path"`
}

type bulkPipelineActivateInput struct {
	Selector bulkSelector `json:"selector"`
	RelPath  string       `json:"rel_path"`
	Active   bool         `json:"active"`
}

func registerBulkPipelines(m *Module) {
	uc := m.deps.EventProcessing.GetPipelineUsecase()

	Add(m, &mcp.Tool{
		Name:        "platform.pipeline.bulk_create",
		Title:       "Bulk create pipeline across tenants (super admin only)",
		Description: "SUPER ADMIN ONLY. Installs the same pipeline (relPath + YAML content) in every selected tenant. Regular tenant users cannot call this. Partial success is normal: system-owned pipelines refuse per tenant and are reported in `failed`.",
	}, Gate{Permission: "eventprocessing.write", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in bulkPipelineUpsertInput) (any, error) {
			tids, err := resolveBulkTenants(ctx, m, in.Selector, false)
			if err != nil {
				return nil, err
			}
			return bulkLoop(ctx, tids, func(ctx context.Context) error {
				_, err := uc.Create(ctx, epdto.CreatePipelineRequest{RelPath: in.RelPath, Content: in.Content})
				return err
			}), nil
		})

	Add(m, &mcp.Tool{
		Name:        "platform.pipeline.bulk_update",
		Title:       "Bulk update pipeline across tenants (super admin only)",
		Description: "SUPER ADMIN ONLY. Rewrites the pipeline with matching relPath across selected tenants. System-owned pipelines refuse per tenant.",
	}, Gate{Permission: "eventprocessing.write", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in bulkPipelineUpsertInput) (any, error) {
			tids, err := resolveBulkTenants(ctx, m, in.Selector, false)
			if err != nil {
				return nil, err
			}
			return bulkLoop(ctx, tids, func(ctx context.Context) error {
				_, err := uc.Update(ctx, epdto.UpdatePipelineRequest{RelPath: in.RelPath, Content: in.Content})
				return err
			}), nil
		})

	Add(m, &mcp.Tool{
		Name:        "platform.pipeline.bulk_delete",
		Title:       "Bulk delete pipeline across tenants (super admin only)",
		Description: "SUPER ADMIN ONLY. Deletes the pipeline with matching relPath across selected tenants. System-owned pipelines refuse per tenant.",
	}, Gate{Permission: "eventprocessing.write", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in bulkPipelineRelPathInput) (any, error) {
			tids, err := resolveBulkTenants(ctx, m, in.Selector, false)
			if err != nil {
				return nil, err
			}
			return bulkLoop(ctx, tids, func(ctx context.Context) error {
				return uc.Delete(ctx, in.RelPath)
			}), nil
		})

	Add(m, &mcp.Tool{
		Name:        "platform.pipeline.bulk_set_active",
		Title:       "Bulk activate/deactivate pipeline across tenants (super admin only)",
		Description: "SUPER ADMIN ONLY. Enables or disables the pipeline per tenant.",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "eventprocessing.write", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in bulkPipelineActivateInput) (any, error) {
			tids, err := resolveBulkTenants(ctx, m, in.Selector, false)
			if err != nil {
				return nil, err
			}
			return bulkLoop(ctx, tids, func(ctx context.Context) error {
				return uc.SetActive(ctx, in.RelPath, in.Active)
			}), nil
		})
}

// ---- platform.correlation_rule.bulk.* --------------------------------------

type bulkCorrelationRuleCreateInput struct {
	Selector bulkSelector          `json:"selector"`
	Rule     epRuleCreateInput     `json:"rule"`
}

type bulkCorrelationRuleUpdateInput struct {
	Selector bulkSelector      `json:"selector"`
	Rule     epRuleUpdateInput `json:"rule"`
}

type bulkCorrelationRuleRelPathInput struct {
	Selector bulkSelector `json:"selector"`
	RelPath  string       `json:"rel_path"`
}

type bulkCorrelationRuleActivateInput struct {
	Selector bulkSelector `json:"selector"`
	RelPath  string       `json:"rel_path"`
	Active   bool         `json:"active"`
}

func registerBulkCorrelationRules(m *Module) {
	uc := m.deps.EventProcessing.GetCorrelationRuleUsecase()

	Add(m, &mcp.Tool{
		Name:        "platform.correlation_rule.bulk_create",
		Title:       "Bulk create correlation rule across tenants (super admin only)",
		Description: "SUPER ADMIN ONLY. Installs the same correlation rule in every selected tenant. System-owned rules refuse per tenant.",
	}, Gate{Permission: "eventprocessing.write", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in bulkCorrelationRuleCreateInput) (any, error) {
			tids, err := resolveBulkTenants(ctx, m, in.Selector, false)
			if err != nil {
				return nil, err
			}
			return bulkLoop(ctx, tids, func(ctx context.Context) error {
				return uc.Create(ctx, in.Rule.toDTO())
			}), nil
		})

	Add(m, &mcp.Tool{
		Name:        "platform.correlation_rule.bulk_update",
		Title:       "Bulk update correlation rule across tenants (super admin only)",
		Description: "SUPER ADMIN ONLY. Updates the correlation rule with matching relPath across tenants. System-owned rules refuse per tenant.",
	}, Gate{Permission: "eventprocessing.write", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in bulkCorrelationRuleUpdateInput) (any, error) {
			tids, err := resolveBulkTenants(ctx, m, in.Selector, false)
			if err != nil {
				return nil, err
			}
			return bulkLoop(ctx, tids, func(ctx context.Context) error {
				return uc.Update(ctx, in.Rule.toDTO())
			}), nil
		})

	Add(m, &mcp.Tool{
		Name:        "platform.correlation_rule.bulk_delete",
		Title:       "Bulk delete correlation rule across tenants (super admin only)",
		Description: "SUPER ADMIN ONLY. Deletes the correlation rule with matching relPath across tenants. System-owned rules refuse per tenant.",
	}, Gate{Permission: "eventprocessing.write", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in bulkCorrelationRuleRelPathInput) (any, error) {
			tids, err := resolveBulkTenants(ctx, m, in.Selector, false)
			if err != nil {
				return nil, err
			}
			return bulkLoop(ctx, tids, func(ctx context.Context) error {
				return uc.Delete(ctx, in.RelPath)
			}), nil
		})

	Add(m, &mcp.Tool{
		Name:        "platform.correlation_rule.bulk_set_active",
		Title:       "Bulk activate/deactivate correlation rule across tenants (super admin only)",
		Description: "SUPER ADMIN ONLY. Enables or disables the correlation rule per tenant.",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "eventprocessing.write", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in bulkCorrelationRuleActivateInput) (any, error) {
			tids, err := resolveBulkTenants(ctx, m, in.Selector, false)
			if err != nil {
				return nil, err
			}
			return bulkLoop(ctx, tids, func(ctx context.Context) error {
				_, err := uc.SetActive(ctx, in.RelPath, in.Active)
				return err
			}), nil
		})
}

// ---- platform.compliance.framework.bulk.* ----------------------------------

type bulkFrameworkUpsertInput struct {
	Selector    bulkSelector                `json:"selector"`
	Key         string                      `json:"key"`
	Name        string                      `json:"name"`
	Description string                      `json:"description,omitempty"`
	Source      string                      `json:"source,omitempty"`
	Sections    []compdomain.FrameworkSection `json:"sections,omitempty"`
}

type bulkFrameworkKeyInput struct {
	Selector bulkSelector `json:"selector"`
	Key      string       `json:"key"`
}

func registerBulkComplianceFrameworks(m *Module) {
	uc := m.deps.Compliance.GetFrameworkUsecase()

	toFramework := func(in bulkFrameworkUpsertInput) compdomain.Framework {
		return compdomain.Framework{
			Key: in.Key, Name: in.Name, Description: in.Description,
			Source: in.Source, Sections: in.Sections,
		}
	}

	Add(m, &mcp.Tool{
		Name:        "platform.compliance.framework.bulk_create",
		Title:       "Bulk create compliance framework across tenants (super admin only)",
		Description: "SUPER ADMIN ONLY. Adds the same custom framework in every selected tenant.",
	}, Gate{Permission: "compliance.write", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in bulkFrameworkUpsertInput) (any, error) {
			tids, err := resolveBulkTenants(ctx, m, in.Selector, false)
			if err != nil {
				return nil, err
			}
			return bulkLoop(ctx, tids, func(ctx context.Context) error {
				_, err := uc.CreateFramework(ctx, toFramework(in))
				return err
			}), nil
		})

	Add(m, &mcp.Tool{
		Name:        "platform.compliance.framework.bulk_update",
		Title:       "Bulk update compliance framework across tenants (super admin only)",
		Description: "SUPER ADMIN ONLY. Updates the framework across tenants. Vendor/system frameworks refuse per tenant.",
	}, Gate{Permission: "compliance.write", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in bulkFrameworkUpsertInput) (any, error) {
			tids, err := resolveBulkTenants(ctx, m, in.Selector, false)
			if err != nil {
				return nil, err
			}
			return bulkLoop(ctx, tids, func(ctx context.Context) error {
				_, err := uc.UpdateFramework(ctx, toFramework(in))
				return err
			}), nil
		})

	Add(m, &mcp.Tool{
		Name:        "platform.compliance.framework.bulk_delete",
		Title:       "Bulk delete compliance framework across tenants (super admin only)",
		Description: "SUPER ADMIN ONLY. Removes the framework overlay across tenants. System-shipped frameworks refuse per tenant.",
	}, Gate{Permission: "compliance.write", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in bulkFrameworkKeyInput) (any, error) {
			tids, err := resolveBulkTenants(ctx, m, in.Selector, false)
			if err != nil {
				return nil, err
			}
			return bulkLoop(ctx, tids, func(ctx context.Context) error {
				return uc.DeleteFramework(ctx, in.Key)
			}), nil
		})
}

// ---- platform.compliance.control.bulk.* ------------------------------------

type bulkControlUpsertInput struct {
	Selector    bulkSelector          `json:"selector"`
	ID          string                `json:"id"`
	Family      string                `json:"family,omitempty"`
	FamilyName  string                `json:"family_name,omitempty"`
	Name        string                `json:"name"`
	Scope       compdomain.ControlScope `json:"scope,omitempty"`
	Statement   string                `json:"statement,omitempty"`
	Remediation string                `json:"remediation,omitempty"`
	Strategy    compdomain.CheckStrategy `json:"strategy,omitempty"`
	Checks      []compdomain.Check     `json:"checks,omitempty"`
	Source      string                `json:"source,omitempty"`
}

type bulkControlIDInput struct {
	Selector bulkSelector `json:"selector"`
	ID       string       `json:"id"`
}

func registerBulkComplianceControls(m *Module) {
	uc := m.deps.Compliance.GetFrameworkUsecase()

	toControl := func(in bulkControlUpsertInput) compdomain.Control {
		return compdomain.Control{
			ID: in.ID, Family: in.Family, FamilyName: in.FamilyName, Name: in.Name,
			Scope: in.Scope, Statement: in.Statement, Remediation: in.Remediation,
			Strategy: in.Strategy, Checks: in.Checks, Source: in.Source,
		}
	}

	Add(m, &mcp.Tool{
		Name:        "platform.compliance.control.bulk_create",
		Title:       "Bulk create compliance control across tenants (super admin only)",
		Description: "SUPER ADMIN ONLY. Adds the same custom control across selected tenants.",
	}, Gate{Permission: "compliance.write", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in bulkControlUpsertInput) (any, error) {
			tids, err := resolveBulkTenants(ctx, m, in.Selector, false)
			if err != nil {
				return nil, err
			}
			return bulkLoop(ctx, tids, func(ctx context.Context) error {
				_, err := uc.CreateControl(ctx, toControl(in))
				return err
			}), nil
		})

	Add(m, &mcp.Tool{
		Name:        "platform.compliance.control.bulk_update",
		Title:       "Bulk update compliance control across tenants (super admin only)",
		Description: "SUPER ADMIN ONLY. Updates the control across selected tenants. System controls refuse per tenant.",
	}, Gate{Permission: "compliance.write", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in bulkControlUpsertInput) (any, error) {
			tids, err := resolveBulkTenants(ctx, m, in.Selector, false)
			if err != nil {
				return nil, err
			}
			return bulkLoop(ctx, tids, func(ctx context.Context) error {
				_, err := uc.UpdateControl(ctx, toControl(in))
				return err
			}), nil
		})

	Add(m, &mcp.Tool{
		Name:        "platform.compliance.control.bulk_delete",
		Title:       "Bulk delete compliance control across tenants (super admin only)",
		Description: "SUPER ADMIN ONLY. Deletes the control overlay across selected tenants. System controls refuse per tenant.",
	}, Gate{Permission: "compliance.write", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in bulkControlIDInput) (any, error) {
			tids, err := resolveBulkTenants(ctx, m, in.Selector, false)
			if err != nil {
				return nil, err
			}
			return bulkLoop(ctx, tids, func(ctx context.Context) error {
				return uc.DeleteControl(ctx, in.ID)
			}), nil
		})
}

// ---- platform.config.smtp.bulk.* -------------------------------------------

type bulkSMTPField struct {
	Key   string `json:"key" jsonschema:"e.g. utmstack.mail.host, utmstack.mail.port, utmstack.mail.username, utmstack.mail.password, utmstack.mail.from"`
	Value string `json:"value"`
}

type bulkSMTPUpdateInput struct {
	Selector bulkSelector    `json:"selector"`
	Fields   []bulkSMTPField `json:"fields"`
}

func registerBulkSMTP(m *Module) {
	uc := m.deps.AppConfig.Usecase()

	// ponytail: strips DefaultTenantID from AllTenants — the platform plane
	// SMTP must not be silently overwritten.
	Add(m, &mcp.Tool{
		Name:        "platform.config.smtp.bulk_update",
		Title:       "Bulk update SMTP config across tenants (super admin only)",
		Description: "SUPER ADMIN ONLY. Upserts one or more utmstack.mail.* fields (host, port, username, password, from, etc.) across the selected tenants. Password field is encrypted at rest. When all_tenants=true the platform default tenant is skipped to prevent silent overwrite.",
	}, Gate{Permission: "config.write", Platform: true},
		func(ctx context.Context, actor *authz.Actor, in bulkSMTPUpdateInput) (any, error) {
			tids, err := resolveBulkTenants(ctx, m, in.Selector, true)
			if err != nil {
				return nil, err
			}
			return bulkLoop(ctx, tids, func(ctx context.Context) error {
				for _, kv := range in.Fields {
					if _, err := uc.Update(ctx, actor.Email, kv.Key, appconfigdto.UpsertRequest{Value: kv.Value}); err != nil {
						return err
					}
				}
				return nil
			}), nil
		})
}

// ---- platform.branding.bulk.* ----------------------------------------------

type bulkBrandingUpdateInput struct {
	Selector       bulkSelector `json:"selector"`
	Enabled        bool         `json:"enabled"`
	ProductName    string       `json:"product_name,omitempty"`
	LogoURL        string       `json:"logo_url,omitempty"`
	LogoDarkURL    string       `json:"logo_dark_url,omitempty"`
	FaviconURL     string       `json:"favicon_url,omitempty"`
	ReportLogoURL  string       `json:"report_logo_url,omitempty"`
	ReportCoverURL string       `json:"report_cover_url,omitempty"`
	AccentColor    string       `json:"accent_color,omitempty"`
}

func registerBulkBranding(m *Module) {
	branding := m.deps.AppConfig.Branding()

	// ponytail: strips DefaultTenantID from AllTenants — same reason as SMTP.
	Add(m, &mcp.Tool{
		Name:        "platform.branding.bulk_update",
		Title:       "Bulk update branding across tenants (super admin only)",
		Description: "SUPER ADMIN ONLY. Applies the same branding (product name, colors, logo URLs, enabled flag) across selected tenants. Requires an Enterprise license per tenant. When all_tenants=true the platform default tenant is skipped.",
	}, Gate{Permission: "config.write", Platform: true},
		func(ctx context.Context, actor *authz.Actor, in bulkBrandingUpdateInput) (any, error) {
			tids, err := resolveBulkTenants(ctx, m, in.Selector, true)
			if err != nil {
				return nil, err
			}
			return bulkLoop(ctx, tids, func(ctx context.Context) error {
				_, err := branding.Update(ctx, actor.Email, appconfigdto.BrandingRequest{
					Enabled: in.Enabled, ProductName: in.ProductName,
					LogoURL: in.LogoURL, LogoDarkURL: in.LogoDarkURL, FaviconURL: in.FaviconURL,
					ReportLogoURL: in.ReportLogoURL, ReportCoverURL: in.ReportCoverURL,
					AccentColor: in.AccentColor,
				})
				return err
			}), nil
		})
}
