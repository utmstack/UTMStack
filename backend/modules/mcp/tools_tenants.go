package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	tenantdom "github.com/utmstack/utmstack/backend/modules/tenant/domain"
	tenantdto "github.com/utmstack/utmstack/backend/modules/tenant/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

// parseID keeps the schema readable ("id: string") and the handler tolerant.
// Every id-carrying tool below routes through it. ponytail: uuid.UUID is a
// [16]byte and infers to array-of-bytes in the SDK schema — not a shape any
// client can produce. String in, uuid out, parse errors surfaced as-is.
func parseID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("id must be a uuid: %w", err)
	}
	return id, nil
}

// Tenant CRUD is fleet-level: only a super admin (an admin of the default
// tenant, per authz.IsPlatform) may see or change other tenants. The Platform
// gate mirrors the REST platform-only group in modules/tenant/routes.go.

type tenantListInput struct {
	Name   string `json:"name,omitempty"`
	Domain string `json:"domain,omitempty"`
	Status string `json:"status,omitempty" jsonschema:"ACTIVE | SUSPENDED | TERMINATED"`
	Page   int    `json:"page,omitempty"`
	Size   int    `json:"size,omitempty"`
}

type tenantIDInput struct {
	ID string `json:"id" jsonschema:"tenant uuid"`
}

type tenantCreateInput struct {
	Name       string `json:"name"`
	Domain     string `json:"domain"`
	AdminEmail string `json:"admin_email"`
}

type tenantUpdateInput struct {
	ID     string `json:"id" jsonschema:"tenant uuid"`
	Name   string `json:"name,omitempty"`
	Domain string `json:"domain,omitempty"`
	Status string `json:"status,omitempty" jsonschema:"ACTIVE | SUSPENDED | TERMINATED"`
	// MaxAIRequests: nil = leave alone, non-nil = set. Clearing the cap needs
	// the REST tri-state (absent / null / number); MCP tools can add that flag
	// when someone actually needs it. ponytail: single case, YAGNI on the rest.
	MaxAIRequests *int `json:"max_ai_requests,omitempty"`
}

type tenantSupportInput struct {
	ID            string `json:"id" jsonschema:"tenant uuid"`
	SupportAccess string `json:"support_access" jsonschema:"NONE | READ | FULL"`
}

func registerTenants(m *Module) {
	if m.deps == nil || m.deps.Tenant == nil {
		return
	}
	uc := m.deps.Tenant.GetTenantUsecase()

	Add(m, &mcp.Tool{
		Name: "tenants.list", Title: "List tenants (super admin)",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "tenant.read", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in tenantListInput) (any, error) {
			f := tenantdto.Filter{
				Name:   in.Name,
				Domain: in.Domain,
				Status: tenantdom.TenantStatus(in.Status),
				Page:   in.Page,
				Size:   clampPageSize(in.Size),
			}
			items, total, err := uc.List(ctx, f)
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "tenants.get", Title: "Get tenant by id (super admin)",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "tenant.read", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in tenantIDInput) (any, error) {
			id, err := parseID(in.ID)
			if err != nil {
				return nil, err
			}
			return uc.GetByID(ctx, id)
		})

	Add(m, &mcp.Tool{
		Name: "tenants.create", Title: "Create tenant (super admin)",
	}, Gate{Permission: "tenant.write", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in tenantCreateInput) (any, error) {
			return uc.Create(ctx, tenantdto.CreateRequest{
				Name:       in.Name,
				Domain:     in.Domain,
				AdminEmail: in.AdminEmail,
			})
		})

	Add(m, &mcp.Tool{
		Name: "tenants.update", Title: "Update tenant (super admin)",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "tenant.write", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in tenantUpdateInput) (any, error) {
			id, err := parseID(in.ID)
			if err != nil {
				return nil, err
			}
			req := tenantdto.UpdateRequest{
				Name:   in.Name,
				Domain: in.Domain,
				Status: tenantdom.TenantStatus(in.Status),
			}
			if in.MaxAIRequests != nil {
				raw, _ := json.Marshal(*in.MaxAIRequests)
				req.MaxAIRequests = raw
			}
			return uc.Update(ctx, id, req)
		})

	Add(m, &mcp.Tool{
		Name: "tenants.terminate", Title: "Terminate tenant (super admin)",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "tenant.write", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in tenantIDInput) (any, error) {
			id, err := parseID(in.ID)
			if err != nil {
				return nil, err
			}
			if err := uc.Terminate(ctx, id); err != nil {
				return nil, err
			}
			return map[string]any{"id": id, "terminated": true}, nil
		})

	Add(m, &mcp.Tool{
		Name:  "tenants.set_support_access",
		Title: "Set support-access grant for a tenant (super admin)",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "tenant.write", Platform: true},
		func(ctx context.Context, _ *authz.Actor, in tenantSupportInput) (any, error) {
			id, err := parseID(in.ID)
			if err != nil {
				return nil, err
			}
			return uc.SetSupportAccess(ctx, id, tenantdom.SupportAccess(in.SupportAccess))
		})
}
