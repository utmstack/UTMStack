package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

func registerSOAR(m *Module) {
	registerSOARRules(m)
	registerSOARTemplates(m)
	registerSOARExecutions(m)
	registerSOARVariables(m)
	registerSOARActions(m)
	registerSOARActionCommands(m)
	registerSOARJobs(m)
	registerSOARAgents(m)
}

// ---- soar.rule.* -----------------------------------------------------------

type soarRuleCreateInput struct {
	Name           string              `json:"name"`
	Description    string              `json:"description,omitempty"`
	Conditions     []dto.FilterVM      `json:"conditions"`
	Commands       []dto.FlowCommandVM `json:"commands"`
	Active         bool                `json:"active"`
	AgentPlatform  string              `json:"agent_platform"`
	DefaultAgent   string              `json:"default_agent,omitempty"`
	Shell          string              `json:"shell,omitempty"`
	ExcludedAgents []string            `json:"excluded_agents,omitempty"`
}

type soarRuleUpdateInput struct {
	RelPath        string              `json:"rel_path"`
	ID             *int64              `json:"id,omitempty"`
	Name           string              `json:"name"`
	Description    string              `json:"description,omitempty"`
	Conditions     []dto.FilterVM      `json:"conditions"`
	Commands       []dto.FlowCommandVM `json:"commands"`
	Active         bool                `json:"active"`
	AgentPlatform  string              `json:"agent_platform"`
	DefaultAgent   string              `json:"default_agent,omitempty"`
	Shell          string              `json:"shell,omitempty"`
	ExcludedAgents []string            `json:"excluded_agents,omitempty"`
}

type soarRuleRelPathInput struct {
	RelPath string `json:"rel_path"`
}

type soarRuleSetEnabledInput struct {
	RelPath string `json:"rel_path"`
	Enabled bool   `json:"enabled"`
}

type soarRuleListInput struct {
	RuleName      string `json:"name,omitempty"`
	RuleActive    *bool  `json:"active,omitempty"`
	AgentPlatform string `json:"agent_platform,omitempty"`
	CreatedBy     string `json:"created_by,omitempty"`
	SystemOwner   *bool  `json:"system_owner,omitempty"`
	Page          int    `json:"page,omitempty"`
	Size          int    `json:"size,omitempty"`
}

func registerSOARRules(m *Module) {
	uc := m.deps.SOAR.GetRuleUsecase()

	Add(m, &mcp.Tool{
		Name: "soar.rule.create", Title: "Create SOAR rule",
	}, Gate{Permission: "soar.write"},
		func(ctx context.Context, actor *authz.Actor, in soarRuleCreateInput) (any, error) {
			if len(in.Conditions) == 0 || len(in.Commands) == 0 {
				return nil, fmt.Errorf("conditions and commands are required")
			}
			active := in.Active
			return uc.Create(ctx, dto.CreateRuleRequest{
				Name: in.Name, Description: in.Description,
				Conditions: in.Conditions, Commands: in.Commands, Active: &active,
				AgentPlatform: in.AgentPlatform, DefaultAgent: in.DefaultAgent,
				Shell: in.Shell, ExcludedAgents: in.ExcludedAgents,
			}, actor.Email)
		})

	Add(m, &mcp.Tool{
		Name: "soar.rule.update", Title: "Update SOAR rule",
	}, Gate{Permission: "soar.write"},
		func(ctx context.Context, actor *authz.Actor, in soarRuleUpdateInput) (any, error) {
			active := in.Active
			return uc.Update(ctx, in.RelPath, dto.UpdateRuleRequest{
				ID: in.ID, Name: in.Name, Description: in.Description,
				Conditions: in.Conditions, Commands: in.Commands, Active: &active,
				AgentPlatform: in.AgentPlatform, DefaultAgent: in.DefaultAgent,
				Shell: in.Shell, ExcludedAgents: in.ExcludedAgents,
			}, actor.Email)
		})

	Add(m, &mcp.Tool{
		Name: "soar.rule.get", Title: "Get SOAR rule",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(ctx context.Context, _ *authz.Actor, in soarRuleRelPathInput) (any, error) {
			return uc.Get(ctx, in.RelPath)
		})

	Add(m, &mcp.Tool{
		Name: "soar.rule.delete", Title: "Delete SOAR rule",
	}, Gate{Permission: "soar.write"},
		func(ctx context.Context, _ *authz.Actor, in soarRuleRelPathInput) (any, error) {
			if err := uc.Delete(ctx, in.RelPath); err != nil {
				return nil, err
			}
			return map[string]any{"rel_path": in.RelPath, "deleted": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "soar.rule.set_enabled", Title: "Enable/disable SOAR rule",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "soar.write"},
		func(ctx context.Context, _ *authz.Actor, in soarRuleSetEnabledInput) (any, error) {
			if err := uc.SetEnabled(ctx, in.RelPath, in.Enabled); err != nil {
				return nil, err
			}
			return map[string]any{"rel_path": in.RelPath, "enabled": in.Enabled}, nil
		})

	Add(m, &mcp.Tool{
		Name: "soar.rule.list", Title: "List SOAR rules",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(ctx context.Context, _ *authz.Actor, in soarRuleListInput) (any, error) {
			f := dto.RuleFilters{
				RuleName: in.RuleName, RuleActive: in.RuleActive,
				AgentPlatform: in.AgentPlatform, CreatedBy: in.CreatedBy, SystemOwner: in.SystemOwner,
				Params: database.Params{Page: in.Page, Size: clampPageSize(in.Size)},
			}
			return uc.List(ctx, f)
		})

	Add(m, &mcp.Tool{
		Name: "soar.rule.resolve_filter_values", Title: "Suggest filter values for rule editor",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(ctx context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			return uc.ResolveFilterValues(ctx)
		})
}

// ---- soar.template.* -------------------------------------------------------

type soarTemplateListInput struct {
	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
	Command     string `json:"command,omitempty"`
	SystemOwner *bool  `json:"system_owner,omitempty"`
	Page        int    `json:"page,omitempty"`
	Size        int    `json:"size,omitempty"`
}

func registerSOARTemplates(m *Module) {
	uc := m.deps.SOAR.GetTemplateUsecase()
	Add(m, &mcp.Tool{
		Name: "soar.template.list", Title: "List rule templates",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(ctx context.Context, _ *authz.Actor, in soarTemplateListInput) (any, error) {
			return uc.List(ctx, dto.TemplateFilters{
				Label: in.Label, Description: in.Description, Command: in.Command, SystemOwner: in.SystemOwner,
				Params: database.Params{Page: in.Page, Size: clampPageSize(in.Size)},
			})
		})
}

// ---- soar.execution.* ------------------------------------------------------

type soarExecutionListInput struct {
	RulePath string `json:"rule_path,omitempty"`
	AlertID  string `json:"alert_id,omitempty"`
	Agent    string `json:"agent,omitempty"`
	Status   string `json:"execution_status,omitempty"`
	DateGTE  string `json:"date_gte,omitempty"`
	DateLTE  string `json:"date_lte,omitempty"`
	Page     int    `json:"page,omitempty"`
	Size     int    `json:"size,omitempty"`
}

func registerSOARExecutions(m *Module) {
	uc := m.deps.SOAR.GetExecutionUsecase()
	Add(m, &mcp.Tool{
		Name: "soar.execution.list", Title: "List rule executions",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(ctx context.Context, _ *authz.Actor, in soarExecutionListInput) (any, error) {
			return uc.List(ctx, dto.ExecutionFilters{
				RulePath: in.RulePath, AlertID: in.AlertID, Agent: in.Agent,
				ExecutionDateGTE: in.DateGTE, ExecutionDateLTE: in.DateLTE,
				Params: database.Params{Page: in.Page, Size: clampPageSize(in.Size)},
			})
		})
}

// ---- soar.variable.* -------------------------------------------------------

type soarVariableCreateInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Value       *string `json:"value,omitempty"`
	IsSecret    bool    `json:"is_secret,omitempty"`
}

type soarVariableUpdateInput struct {
	ID          int64   `json:"id"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Value       *string `json:"value,omitempty"`
	IsSecret    bool    `json:"is_secret,omitempty"`
}

type soarVariableListInput struct {
	Name *string `json:"name,omitempty"`
	Page int     `json:"page,omitempty"`
	Size int     `json:"size,omitempty"`
}

type idInt64Input struct {
	ID int64 `json:"id"`
}

func registerSOARVariables(m *Module) {
	uc := m.deps.SOAR.GetVariableUsecase()

	Add(m, &mcp.Tool{
		Name: "soar.variable.create", Title: "Create incident variable",
	}, Gate{Permission: "soar.write"},
		func(_ context.Context, actor *authz.Actor, in soarVariableCreateInput) (any, error) {
			return uc.Create(dto.CreateVariableRequest{
				VariableName: in.Name, VariableDescription: in.Description, VariableValue: in.Value, IsSecret: in.IsSecret,
			}, actor.Email)
		})

	Add(m, &mcp.Tool{
		Name: "soar.variable.update", Title: "Update incident variable",
	}, Gate{Permission: "soar.write"},
		func(_ context.Context, actor *authz.Actor, in soarVariableUpdateInput) (any, error) {
			return uc.Update(dto.UpdateVariableRequest{
				ID: in.ID, VariableName: in.Name, VariableDescription: in.Description, VariableValue: in.Value, IsSecret: in.IsSecret,
			}, actor.Email)
		})

	Add(m, &mcp.Tool{
		Name: "soar.variable.get", Title: "Get incident variable",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(_ context.Context, _ *authz.Actor, in idInt64Input) (any, error) {
			return uc.FindByID(in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "soar.variable.list", Title: "List incident variables",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(_ context.Context, _ *authz.Actor, in soarVariableListInput) (any, error) {
			items, total, err := uc.FindAll(dto.VariableFilter{
				VariableName: in.Name,
				Params:       database.Params{Page: in.Page, Size: clampPageSize(in.Size)},
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "soar.variable.delete", Title: "Delete incident variable",
	}, Gate{Permission: "soar.write"},
		func(_ context.Context, _ *authz.Actor, in idInt64Input) (any, error) {
			if err := uc.Delete(in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})
}

// ---- soar.action.* / soar.action_command.* / soar.job.* --------------------

type soarActionCreateInput struct {
	ActionCommand     *string `json:"action_command,omitempty"`
	ActionDescription *string `json:"action_description,omitempty"`
	ActionParams      *string `json:"action_params,omitempty"`
	ActionType        *int    `json:"action_type,omitempty"`
	ActionEditable    bool    `json:"action_editable,omitempty"`
}

type soarActionUpdateInput struct {
	ID                int64   `json:"id"`
	ActionCommand     *string `json:"action_command,omitempty"`
	ActionDescription *string `json:"action_description,omitempty"`
	ActionParams      *string `json:"action_params,omitempty"`
	ActionType        *int    `json:"action_type,omitempty"`
	ActionEditable    bool    `json:"action_editable,omitempty"`
}

type soarActionListInput struct {
	ActionCommand  *string `json:"action_command,omitempty"`
	ActionType     *int    `json:"action_type,omitempty"`
	ActionEditable *bool   `json:"action_editable,omitempty"`
	Page           int     `json:"page,omitempty"`
	Size           int     `json:"size,omitempty"`
}

func registerSOARActions(m *Module) {
	uc := m.deps.SOAR.GetActionUsecase()

	Add(m, &mcp.Tool{
		Name: "soar.action.create", Title: "Create incident action",
	}, Gate{Permission: "soar.write"},
		func(_ context.Context, actor *authz.Actor, in soarActionCreateInput) (any, error) {
			return uc.Create(dto.CreateActionRequest{
				ActionCommand: in.ActionCommand, ActionDescription: in.ActionDescription,
				ActionParams: in.ActionParams, ActionType: in.ActionType, ActionEditable: in.ActionEditable,
			}, actor.Email)
		})

	Add(m, &mcp.Tool{
		Name: "soar.action.update", Title: "Update incident action",
	}, Gate{Permission: "soar.write"},
		func(_ context.Context, actor *authz.Actor, in soarActionUpdateInput) (any, error) {
			return uc.Update(dto.UpdateActionRequest{
				ID: in.ID, ActionCommand: in.ActionCommand, ActionDescription: in.ActionDescription,
				ActionParams: in.ActionParams, ActionType: in.ActionType, ActionEditable: in.ActionEditable,
			}, actor.Email)
		})

	Add(m, &mcp.Tool{
		Name: "soar.action.get", Title: "Get incident action",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(_ context.Context, _ *authz.Actor, in idInt64Input) (any, error) {
			return uc.FindByID(in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "soar.action.list", Title: "List incident actions",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(_ context.Context, _ *authz.Actor, in soarActionListInput) (any, error) {
			items, total, err := uc.FindAll(dto.ActionFilter{
				ActionCommand: in.ActionCommand, ActionType: in.ActionType, ActionEditable: in.ActionEditable,
				Params: database.Params{Page: in.Page, Size: clampPageSize(in.Size)},
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "soar.action.delete", Title: "Delete incident action",
	}, Gate{Permission: "soar.write"},
		func(_ context.Context, _ *authz.Actor, in idInt64Input) (any, error) {
			if err := uc.Delete(in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})
}

type soarActionCommandCreateInput struct {
	ActionID   int64   `json:"action_id"`
	OsPlatform *string `json:"os_platform,omitempty"`
	Command    *string `json:"command,omitempty"`
}

type soarActionCommandUpdateInput struct {
	ID         int64   `json:"id"`
	ActionID   int64   `json:"action_id"`
	OsPlatform *string `json:"os_platform,omitempty"`
	Command    *string `json:"command,omitempty"`
}

type soarActionCommandListInput struct {
	ActionID   *int64  `json:"action_id,omitempty"`
	OsPlatform *string `json:"os_platform,omitempty"`
	Command    *string `json:"command,omitempty"`
	Page       int     `json:"page,omitempty"`
	Size       int     `json:"size,omitempty"`
}

func registerSOARActionCommands(m *Module) {
	uc := m.deps.SOAR.GetActionCommandUsecase()

	Add(m, &mcp.Tool{
		Name: "soar.action_command.create", Title: "Create action command",
	}, Gate{Permission: "soar.write"},
		func(_ context.Context, _ *authz.Actor, in soarActionCommandCreateInput) (any, error) {
			return uc.Create(dto.CreateActionCommandRequest{
				ActionID: in.ActionID, OsPlatform: in.OsPlatform, Command: in.Command,
			})
		})

	Add(m, &mcp.Tool{
		Name: "soar.action_command.update", Title: "Update action command",
	}, Gate{Permission: "soar.write"},
		func(_ context.Context, _ *authz.Actor, in soarActionCommandUpdateInput) (any, error) {
			return uc.Update(dto.UpdateActionCommandRequest{
				ID: in.ID, ActionID: in.ActionID, OsPlatform: in.OsPlatform, Command: in.Command,
			})
		})

	Add(m, &mcp.Tool{
		Name: "soar.action_command.get", Title: "Get action command",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(_ context.Context, _ *authz.Actor, in idInt64Input) (any, error) {
			return uc.FindByID(in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "soar.action_command.list", Title: "List action commands",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(_ context.Context, _ *authz.Actor, in soarActionCommandListInput) (any, error) {
			items, total, err := uc.FindAll(dto.ActionCommandFilter{
				ActionID: in.ActionID, OsPlatform: in.OsPlatform, Command: in.Command,
				Params: database.Params{Page: in.Page, Size: clampPageSize(in.Size)},
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "soar.action_command.delete", Title: "Delete action command",
	}, Gate{Permission: "soar.write"},
		func(_ context.Context, _ *authz.Actor, in idInt64Input) (any, error) {
			if err := uc.Delete(in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})
}

type soarJobCreateInput struct {
	ActionID   *int64  `json:"action_id,omitempty"`
	Params     *string `json:"params,omitempty"`
	Agent      *string `json:"agent,omitempty"`
	Status     *int    `json:"status,omitempty"`
	OriginID   string  `json:"origin_id"`
	OriginType string  `json:"origin_type"`
}

type soarJobListInput struct {
	ActionID   *int64  `json:"action_id,omitempty"`
	Agent      *string `json:"agent,omitempty"`
	Status     *int    `json:"status,omitempty"`
	OriginID   *int    `json:"origin_id,omitempty"`
	OriginType *string `json:"origin_type,omitempty"`
	Page       int     `json:"page,omitempty"`
	Size       int     `json:"size,omitempty"`
}

func registerSOARJobs(m *Module) {
	uc := m.deps.SOAR.GetJobUsecase()

	Add(m, &mcp.Tool{
		Name:        "soar.job.create",
		Title:       "Create SOAR job",
		Description: "Schedules an action command for execution on a target agent. DESTRUCTIVE: triggers a real command.",
		Annotations: &mcp.ToolAnnotations{},
	}, Gate{Permission: "soar.write"},
		func(_ context.Context, actor *authz.Actor, in soarJobCreateInput) (any, error) {
			return uc.Create(dto.CreateJobRequest{
				ActionID: in.ActionID, Params: in.Params, Agent: in.Agent, Status: in.Status,
				OriginID: in.OriginID, OriginType: in.OriginType,
			}, actor.Email)
		})

	Add(m, &mcp.Tool{
		Name: "soar.job.get", Title: "Get SOAR job",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(_ context.Context, _ *authz.Actor, in idInt64Input) (any, error) {
			return uc.FindByID(in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "soar.job.list", Title: "List SOAR jobs",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(_ context.Context, _ *authz.Actor, in soarJobListInput) (any, error) {
			items, total, err := uc.FindAll(dto.JobFilter{
				ActionID: in.ActionID, Agent: in.Agent, Status: in.Status,
				OriginID: in.OriginID, OriginType: in.OriginType,
				Params: database.Params{Page: in.Page, Size: clampPageSize(in.Size)},
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "soar.job.count", Title: "Count SOAR jobs",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(_ context.Context, _ *authz.Actor, in soarJobListInput) (any, error) {
			c, err := uc.Count(dto.JobFilter{
				ActionID: in.ActionID, Agent: in.Agent, Status: in.Status,
				OriginID: in.OriginID, OriginType: in.OriginType,
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"count": c}, nil
		})

	Add(m, &mcp.Tool{
		Name: "soar.job.delete", Title: "Delete SOAR job",
	}, Gate{Permission: "soar.write"},
		func(_ context.Context, _ *authz.Actor, in idInt64Input) (any, error) {
			if err := uc.Delete(in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})
}

// ---- soar.agent.* ----------------------------------------------------------

type soarAgentListInput struct {
	Platform string `json:"platform" jsonschema:"e.g. windows | linux | macos"`
}

func registerSOARAgents(m *Module) {
	uc := m.deps.SOAR.GetAgentUsecase()
	Add(m, &mcp.Tool{
		Name: "soar.agent.list_by_platform", Title: "List agents on a platform",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(ctx context.Context, _ *authz.Actor, in soarAgentListInput) (any, error) {
			return uc.ListByPlatform(ctx, in.Platform)
		})
}
