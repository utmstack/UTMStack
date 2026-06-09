package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

func registerCompliance(m *Module) {
	registerComplianceFrameworks(m)
	registerComplianceControls(m)
	registerComplianceReports(m)
	registerComplianceSchedules(m)
}

// ---- compliance.framework.* ------------------------------------------------

type complianceKeyInput struct {
	Key string `json:"key"`
}

type complianceFrameworkUpsertInput struct {
	Key         string                    `json:"key"`
	Name        string                    `json:"name"`
	Description string                    `json:"description,omitempty"`
	Source      string                    `json:"source,omitempty"`
	Sections    []domain.FrameworkSection `json:"sections,omitempty"`
}

type complianceFrameworkSetEnabledInput struct {
	Key     string `json:"key"`
	Enabled bool   `json:"enabled"`
}

func registerComplianceFrameworks(m *Module) {
	uc := m.deps.Compliance.GetFrameworkUsecase()

	Add(m, &mcp.Tool{
		Name: "compliance.framework.list", Title: "List compliance frameworks",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "compliance.read"},
		func(ctx context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			return uc.ListFrameworks(ctx), nil
		})

	Add(m, &mcp.Tool{
		Name: "compliance.framework.get", Title: "Get framework",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "compliance.read"},
		func(ctx context.Context, _ *authz.Actor, in complianceKeyInput) (any, error) {
			return uc.GetFramework(ctx, in.Key)
		})

	Add(m, &mcp.Tool{
		Name: "compliance.framework.create", Title: "Create custom framework",
	}, Gate{Permission: "compliance.write"},
		func(ctx context.Context, _ *authz.Actor, in complianceFrameworkUpsertInput) (any, error) {
			return uc.CreateFramework(ctx, domain.Framework{
				Key: in.Key, Name: in.Name, Description: in.Description, Source: in.Source, Sections: in.Sections,
			})
		})

	Add(m, &mcp.Tool{
		Name: "compliance.framework.update", Title: "Update framework (user overlay)",
	}, Gate{Permission: "compliance.write"},
		func(ctx context.Context, _ *authz.Actor, in complianceFrameworkUpsertInput) (any, error) {
			return uc.UpdateFramework(ctx, domain.Framework{
				Key: in.Key, Name: in.Name, Description: in.Description, Source: in.Source, Sections: in.Sections,
			})
		})

	Add(m, &mcp.Tool{
		Name: "compliance.framework.delete", Title: "Delete framework",
	}, Gate{Permission: "compliance.write"},
		func(ctx context.Context, _ *authz.Actor, in complianceKeyInput) (any, error) {
			if err := uc.DeleteFramework(ctx, in.Key); err != nil {
				return nil, err
			}
			return map[string]any{"key": in.Key, "deleted": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "compliance.framework.set_enabled", Title: "Enable or disable framework",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "compliance.write"},
		func(ctx context.Context, _ *authz.Actor, in complianceFrameworkSetEnabledInput) (any, error) {
			if err := uc.SetFrameworkEnabled(ctx, in.Key, in.Enabled); err != nil {
				return nil, err
			}
			return map[string]any{"key": in.Key, "enabled": in.Enabled}, nil
		})
}

// ---- compliance.control.* --------------------------------------------------

type complianceControlIDInput struct {
	ID string `json:"id"`
}

type complianceControlSetEnabledInput struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

type complianceControlUpsertInput struct {
	ID          string         `json:"id"`
	Family      string         `json:"family,omitempty"`
	FamilyName  string         `json:"family_name,omitempty"`
	Name        string         `json:"name"`
	Scope       string         `json:"scope,omitempty"`
	Statement   string         `json:"statement,omitempty"`
	Remediation string         `json:"remediation,omitempty"`
	Strategy    string         `json:"strategy,omitempty"`
	Checks      []domain.Check `json:"checks,omitempty"`
	Source      string         `json:"source,omitempty"`
}

func registerComplianceControls(m *Module) {
	uc := m.deps.Compliance.GetFrameworkUsecase()

	Add(m, &mcp.Tool{
		Name: "compliance.control.list", Title: "List controls",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "compliance.read"},
		func(ctx context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			return uc.ListControls(ctx), nil
		})

	Add(m, &mcp.Tool{
		Name: "compliance.control.get", Title: "Get control",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "compliance.read"},
		func(ctx context.Context, _ *authz.Actor, in complianceControlIDInput) (any, error) {
			return uc.GetControl(ctx, in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "compliance.control.create", Title: "Create control",
	}, Gate{Permission: "compliance.write"},
		func(ctx context.Context, _ *authz.Actor, in complianceControlUpsertInput) (any, error) {
			return uc.CreateControl(ctx, domain.Control{
				ID: in.ID, Family: in.Family, FamilyName: in.FamilyName, Name: in.Name,
				Scope: in.Scope, Statement: in.Statement, Remediation: in.Remediation,
				Strategy: in.Strategy, Checks: in.Checks, Source: in.Source,
			})
		})

	Add(m, &mcp.Tool{
		Name: "compliance.control.update", Title: "Update control",
	}, Gate{Permission: "compliance.write"},
		func(ctx context.Context, _ *authz.Actor, in complianceControlUpsertInput) (any, error) {
			return uc.UpdateControl(ctx, domain.Control{
				ID: in.ID, Family: in.Family, FamilyName: in.FamilyName, Name: in.Name,
				Scope: in.Scope, Statement: in.Statement, Remediation: in.Remediation,
				Strategy: in.Strategy, Checks: in.Checks, Source: in.Source,
			})
		})

	Add(m, &mcp.Tool{
		Name: "compliance.control.delete", Title: "Delete control",
	}, Gate{Permission: "compliance.write"},
		func(ctx context.Context, _ *authz.Actor, in complianceControlIDInput) (any, error) {
			if err := uc.DeleteControl(ctx, in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "compliance.control.set_enabled", Title: "Enable or disable control",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "compliance.write"},
		func(ctx context.Context, _ *authz.Actor, in complianceControlSetEnabledInput) (any, error) {
			if err := uc.SetControlEnabled(ctx, in.ID, in.Enabled); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "enabled": in.Enabled}, nil
		})
}

// ---- compliance.report.* ---------------------------------------------------

type complianceReportEvalInput struct {
	FrameworkKey string `json:"framework_key"`
}

type complianceReportListInput struct {
	FrameworkKey string `json:"framework_key"`
	Limit        int    `json:"limit,omitempty"`
}

type complianceReportIDInput struct {
	ID string `json:"id"`
}

func registerComplianceReports(m *Module) {
	uc := m.deps.Compliance.GetEvaluatorUsecase()

	Add(m, &mcp.Tool{
		Name: "compliance.report.evaluate", Title: "Evaluate framework (no snapshot)",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "compliance.read"},
		func(ctx context.Context, _ *authz.Actor, in complianceReportEvalInput) (any, error) {
			if in.FrameworkKey == "" {
				return nil, fmt.Errorf("framework_key is required")
			}
			return uc.EvaluateFramework(ctx, in.FrameworkKey)
		})

	Add(m, &mcp.Tool{
		Name: "compliance.report.generate", Title: "Generate compliance report (saves snapshot)",
	}, Gate{Permission: "compliance.write"},
		func(ctx context.Context, _ *authz.Actor, in complianceReportEvalInput) (any, error) {
			if in.FrameworkKey == "" {
				return nil, fmt.Errorf("framework_key is required")
			}
			return uc.GenerateReport(ctx, in.FrameworkKey)
		})

	Add(m, &mcp.Tool{
		Name: "compliance.report.list", Title: "List historical reports",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "compliance.read"},
		func(ctx context.Context, _ *authz.Actor, in complianceReportListInput) (any, error) {
			lim := in.Limit
			if lim <= 0 {
				lim = 25
			}
			return uc.ListReports(ctx, in.FrameworkKey, lim)
		})

	Add(m, &mcp.Tool{
		Name: "compliance.report.get", Title: "Get report snapshot",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "compliance.read"},
		func(ctx context.Context, _ *authz.Actor, in complianceReportIDInput) (any, error) {
			return uc.GetReport(ctx, in.ID)
		})
}

// ---- compliance.schedule.* -------------------------------------------------

type complianceScheduleCreateInput struct {
	FrameworkKey   string `json:"framework_key"`
	ScheduleString string `json:"schedule_string"`
	Recipients     string `json:"recipients,omitempty"`
}

type complianceScheduleUpdateInput struct {
	ID             int64  `json:"id"`
	FrameworkKey   string `json:"framework_key"`
	ScheduleString string `json:"schedule_string"`
	Recipients     string `json:"recipients,omitempty"`
}

type complianceScheduleListInput struct {
	FrameworkKey string `json:"framework_key,omitempty"`
	Page         int    `json:"page,omitempty"`
	Size         int    `json:"size,omitempty"`
}

func registerComplianceSchedules(m *Module) {
	uc := m.deps.Compliance.GetScheduleUsecase()

	Add(m, &mcp.Tool{
		Name: "compliance.schedule.create", Title: "Schedule recurring report",
	}, Gate{Permission: "compliance.write"},
		func(ctx context.Context, actor *authz.Actor, in complianceScheduleCreateInput) (any, error) {
			return uc.Create(ctx, int64(actor.UserID), dto.CreateScheduleRequest{
				FrameworkKey: in.FrameworkKey, ScheduleString: in.ScheduleString, Recipients: in.Recipients,
			})
		})

	Add(m, &mcp.Tool{
		Name: "compliance.schedule.update", Title: "Update report schedule",
	}, Gate{Permission: "compliance.write"},
		func(ctx context.Context, actor *authz.Actor, in complianceScheduleUpdateInput) (any, error) {
			return uc.Update(ctx, int64(actor.UserID), dto.UpdateScheduleRequest{
				ID: in.ID, FrameworkKey: in.FrameworkKey, ScheduleString: in.ScheduleString, Recipients: in.Recipients,
			})
		})

	Add(m, &mcp.Tool{
		Name: "compliance.schedule.list_by_user", Title: "List my report schedules",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "compliance.read"},
		func(ctx context.Context, actor *authz.Actor, in complianceScheduleListInput) (any, error) {
			items, total, err := uc.ListByUser(ctx, int64(actor.UserID), dto.ScheduleFilters{
				FrameworkKey: in.FrameworkKey, Page: in.Page, Size: clampPageSize(in.Size),
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "compliance.schedule.get", Title: "Get schedule",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "compliance.read"},
		func(ctx context.Context, _ *authz.Actor, in idInt64Input) (any, error) {
			return uc.GetByID(ctx, in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "compliance.schedule.delete", Title: "Delete schedule",
	}, Gate{Permission: "compliance.write"},
		func(ctx context.Context, _ *authz.Actor, in idInt64Input) (any, error) {
			if err := uc.Delete(ctx, in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})
}
