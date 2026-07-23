package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	adauditdto "github.com/utmstack/utmstack/backend/modules/adaudit/dto"
	appconfigdomain "github.com/utmstack/utmstack/backend/modules/appconfig/domain"
	appconfigdto "github.com/utmstack/utmstack/backend/modules/appconfig/dto"
	notifdomain "github.com/utmstack/utmstack/backend/modules/notifications/domain"
	notifdto "github.com/utmstack/utmstack/backend/modules/notifications/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

// This file groups the small/single-purpose tool registrations to avoid one
// file per module: audit (2), adaudit (1), notifications (7), appconfig (7),
// billing (2), socai (1) = 20 tools.

// ---- audit.* ---------------------------------------------------------------

type auditListInput struct {
	PageNumber  int    `json:"page_number,omitempty"`
	PageSize    int    `json:"page_size,omitempty"`
	SearchQuery string `json:"search_query,omitempty" jsonschema:"DSL e.g. user_login.equals.alex&event_type.equals.MCP_TOOL_CALL"`
	SortBy      string `json:"sort_by,omitempty"`
}

type auditIDInput struct {
	ID uint64 `json:"id"`
}

func registerAudit(m *Module) {
	uc := m.deps.Audit.Usecase()

	Add(m, &mcp.Tool{
		Name: "audit.list", Title: "List audit log entries",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "audit.read"},
		func(ctx context.Context, _ *authz.Actor, in auditListInput) (any, error) {
			return uc.List(ctx, common_models.ListRequest{
				PageNumber: in.PageNumber, PageSize: clampPageSize(in.PageSize),
				SearchQuery: in.SearchQuery, SortBy: in.SortBy,
			})
		})

	Add(m, &mcp.Tool{
		Name: "audit.get", Title: "Get audit entry",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "audit.read"},
		func(ctx context.Context, _ *authz.Actor, in auditIDInput) (any, error) {
			return uc.Get(ctx, in.ID)
		})
}

// ---- adaudit.* -------------------------------------------------------------

type adauditListInput struct {
	Search   string `json:"search,omitempty"`
	TenantID string `json:"tenant_id,omitempty"`
	Active   *bool  `json:"active,omitempty"`
	Page     int    `json:"page,omitempty"`
	Size     int    `json:"size,omitempty"`
}

func registerADAudit(m *Module) {
	uc := m.deps.ADAudit.GetUsecase()

	Add(m, &mcp.Tool{
		Name: "adaudit.users.list", Title: "List Active Directory users",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "adaudit.read"},
		func(ctx context.Context, _ *authz.Actor, in adauditListInput) (any, error) {
			f := adauditdto.ADUserFilter{
				Search: in.Search, TenantID: in.TenantID, Active: in.Active,
			}
			f.Page = in.Page
			f.Size = clampPageSize(in.Size)
			return uc.List(ctx, f)
		})
}

// ---- notifications.* -------------------------------------------------------

type notifListInput struct {
	Source string `json:"source,omitempty"`
	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"`
	Read   *bool  `json:"read,omitempty"`
	Page   int    `json:"page,omitempty"`
	Size   int    `json:"size,omitempty"`
	Sort   string `json:"sort,omitempty"`
}

type notifIDInput struct {
	ID int64 `json:"id"`
}

type notifMarkReadInput struct {
	ID   int64 `json:"id"`
	Read bool  `json:"read"`
}

type notifUpdateStatusInput struct {
	ID     int64  `json:"id"`
	Status string `json:"status" jsonschema:"ACTIVE | HIDDEN | DELETED"`
}

func registerNotifications(m *Module) {
	uc := m.deps.Notifications.GetNotificationUsecase()

	Add(m, &mcp.Tool{
		Name: "notifications.list", Title: "List notifications",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{},
		func(ctx context.Context, _ *authz.Actor, in notifListInput) (any, error) {
			q := notifdto.NotificationListQuery{Read: in.Read, Sort: in.Sort}
			q.Page = in.Page
			q.Size = clampPageSize(in.Size)
			if in.Source != "" {
				s := notifdomain.NotificationSource(in.Source)
				q.Source = &s
			}
			if in.Type != "" {
				t := notifdomain.NotificationType(in.Type)
				q.Type = &t
			}
			if in.Status != "" {
				st := notifdomain.NotificationStatus(in.Status)
				q.Status = &st
			}
			items, total, err := uc.List(ctx, q)
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "notifications.get", Title: "Get notification",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{},
		func(ctx context.Context, _ *authz.Actor, in notifIDInput) (any, error) {
			return uc.GetByID(ctx, in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "notifications.unread_count", Title: "Count unread notifications",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{},
		func(ctx context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			c, err := uc.CountUnread(ctx)
			if err != nil {
				return nil, err
			}
			return map[string]any{"count": c}, nil
		})

	Add(m, &mcp.Tool{
		Name: "notifications.mark_read", Title: "Mark notification read/unread",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{},
		func(ctx context.Context, _ *authz.Actor, in notifMarkReadInput) (any, error) {
			return uc.MarkRead(ctx, in.ID, in.Read)
		})

	Add(m, &mcp.Tool{
		Name: "notifications.mark_all_read", Title: "Mark all notifications read",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{},
		func(ctx context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			if err := uc.MarkAllRead(ctx); err != nil {
				return nil, err
			}
			return map[string]any{"marked": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "notifications.update_status", Title: "Update notification status",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{},
		func(ctx context.Context, _ *authz.Actor, in notifUpdateStatusInput) (any, error) {
			return uc.UpdateStatus(ctx, in.ID, notifdomain.NotificationStatus(in.Status))
		})

	Add(m, &mcp.Tool{
		Name: "notifications.delete", Title: "Delete notification",
	}, Gate{},
		func(ctx context.Context, _ *authz.Actor, in notifIDInput) (any, error) {
			if err := uc.Delete(ctx, in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})
}

// ---- config.* / branding.* -------------------------------------------------

type configKeyInput struct {
	Key string `json:"key"`
}

type configUpdateInput struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	IsSecret    *bool  `json:"is_secret,omitempty"`
	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
}

type configCheckMailInput struct {
	Configs []appconfigdomain.MailConfig `json:"configs"`
}

type brandingUpdateInput struct {
	Enabled        bool   `json:"enabled"`
	ProductName    string `json:"product_name,omitempty"`
	LogoURL        string `json:"logo_url,omitempty"`
	LogoDarkURL    string `json:"logo_dark_url,omitempty"`
	FaviconURL     string `json:"favicon_url,omitempty"`
	ReportLogoURL  string `json:"report_logo_url,omitempty"`
	ReportCoverURL string `json:"report_cover_url,omitempty"`
}

func registerAppConfig(m *Module) {
	uc := m.deps.AppConfig.Usecase()
	branding := m.deps.AppConfig.Branding()

	Add(m, &mcp.Tool{
		Name: "config.list", Title: "List configs",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "config.read"},
		func(ctx context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			return uc.List(ctx)
		})

	Add(m, &mcp.Tool{
		Name: "config.get", Title: "Get config by key",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "config.read"},
		func(ctx context.Context, _ *authz.Actor, in configKeyInput) (any, error) {
			return uc.Get(ctx, in.Key)
		})

	Add(m, &mcp.Tool{
		Name: "config.update", Title: "Update config value",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "config.write"},
		func(ctx context.Context, actor *authz.Actor, in configUpdateInput) (any, error) {
			return uc.Update(ctx, actor.Login, in.Key, appconfigdto.UpsertRequest{
				Value: in.Value, IsSecret: in.IsSecret, Label: in.Label, Description: in.Description,
			})
		})

	Add(m, &mcp.Tool{
		Name: "config.check_mail", Title: "Test SMTP configuration",
	}, Gate{Permission: "config.write"},
		func(ctx context.Context, _ *authz.Actor, in configCheckMailInput) (any, error) {
			if err := uc.CheckMail(ctx, in.Configs); err != nil {
				return nil, err
			}
			return map[string]any{"ok": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "branding.get", Title: "Get branding (admin view)",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "config.read"},
		func(ctx context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			return branding.Get(ctx)
		})

	Add(m, &mcp.Tool{
		Name: "branding.update", Title: "Update branding (MSSP)",
	}, Gate{Permission: "config.write", MSSP: true},
		func(ctx context.Context, actor *authz.Actor, in brandingUpdateInput) (any, error) {
			return branding.Update(ctx, actor.Login, appconfigdto.BrandingRequest{
				Enabled: in.Enabled, ProductName: in.ProductName,
				LogoURL: in.LogoURL, LogoDarkURL: in.LogoDarkURL, FaviconURL: in.FaviconURL,
				ReportLogoURL: in.ReportLogoURL, ReportCoverURL: in.ReportCoverURL,
			})
		})

	Add(m, &mcp.Tool{
		Name: "branding.get_public", Title: "Get public branding (login-page view)",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{},
		func(ctx context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			return branding.GetPublic(ctx)
		})
}

// ---- billing.* -------------------------------------------------------------

func registerBilling(m *Module) {
	Add(m, &mcp.Tool{
		Name: "billing.version", Title: "Server version info",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{},
		func(_ context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			return m.deps.Billing.Version().Info()
		})

	Add(m, &mcp.Tool{
		Name: "billing.license", Title: "Active license",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{},
		func(_ context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			return m.deps.Billing.License().Current(), nil
		})
}

// ---- socai.* ---------------------------------------------------------------

type socaiAnalyzeInput struct {
	Alert any `json:"alert" jsonschema:"Alert payload (any JSON object) forwarded to the external SOC AI service"`
}

func registerSOCAI(m *Module) {
	client := m.deps.SOCAI.Client()
	Add(m, &mcp.Tool{
		Name:        "socai.analyze_alert",
		Title:       "Forward alert to SOC AI for analysis",
		Description: "POSTs the alert payload to the configured SOC AI service. Returns its async ack; analysis results flow back through the regular alerts/incidents pipeline.",
		Annotations: &mcp.ToolAnnotations{},
	}, Gate{},
		func(ctx context.Context, _ *authz.Actor, in socaiAnalyzeInput) (any, error) {
			if in.Alert == nil {
				return nil, fmt.Errorf("alert is required")
			}
			raw, err := json.Marshal(in.Alert)
			if err != nil {
				return nil, fmt.Errorf("marshal alert: %w", err)
			}
			status, body, err := client.Analyze(ctx, raw)
			if err != nil {
				return nil, err
			}
			return map[string]any{"status": status, "body": json.RawMessage(body)}, nil
		})
}
