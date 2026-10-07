package mcp

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

func registerSOAR(m *Module) {
	registerSOARRules(m)
	registerSOARExecutions(m)
	registerSOARVariables(m)
	registerSOARAgents(m)
	registerSOARPrompts(m)
}

// ---- soar.flow.* -----------------------------------------------------------
func mergeFlowUpdate(cur *dto.RuleResponse, in soarRuleUpdateInput) (dto.UpdateRuleRequest, error) {
	for _, id := range in.DeletedNodes {
		if _, ok := in.Nodes[id]; ok {
			return dto.UpdateRuleRequest{}, fmt.Errorf("node %q appears in both nodes and deleted_nodes", id)
		}
	}

	nodes := make(map[string]dto.FlowNodeVM, len(cur.Nodes))
	for id, n := range cur.Nodes {
		if _, del := in.Nodes[id]; del || nodeInList(in.DeletedNodes, id) {
			continue
		}
		n.OnSuccess = scrubEdges(n.OnSuccess, in.DeletedNodes)
		n.OnError = scrubEdges(n.OnError, in.DeletedNodes)
		nodes[id] = n
	}
	for id, n := range in.Nodes {
		nodes[id] = n
	}

	roots := cur.Roots
	if in.Roots != nil {
		roots = *in.Roots
	}
	roots = scrubEdges(roots, in.DeletedNodes)

	req := dto.UpdateRuleRequest{
		Name:        cur.Name,
		Description: cur.Description,
		Conditions:  cur.Conditions,
		Roots:       roots,
		Nodes:       nodes,
		MaxDepth:    cur.MaxDepth,
	}
	if in.Name != nil {
		req.Name = *in.Name
	}
	if in.Description != nil {
		req.Description = *in.Description
	}
	if in.Conditions != nil {
		req.Conditions = *in.Conditions
	}
	if in.MaxDepth != nil {
		req.MaxDepth = *in.MaxDepth
	}
	if in.Active != nil {
		req.Active = in.Active
	}
	return req, nil
}

func nodeInList(ids []string, id string) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

func scrubEdges(edges []string, deleted []string) []string {
	out := make([]string, 0, len(edges))
	for _, e := range edges {
		if !nodeInList(deleted, e) {
			out = append(out, e)
		}
	}
	return out
}

// ---- soar.rule.* -----------------------------------------------------------

type soarRuleCreateInput struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description,omitempty"`
	Conditions  []dto.FilterVM            `json:"conditions"`
	Roots       []string                  `json:"roots"`
	Nodes       map[string]dto.FlowNodeVM `json:"nodes"`
	MaxDepth    int                       `json:"max_depth,omitempty"`
	Active      bool                      `json:"active"`
}

type soarRuleUpdateInput struct {
	RelPath      string                    `json:"rel_path"`
	Name         *string                   `json:"name,omitempty" jsonschema:"New flow name; omit to keep current"`
	Description  *string                   `json:"description,omitempty" jsonschema:"New flow description; omit to keep current"`
	Conditions   *[]dto.FilterVM           `json:"conditions,omitempty" jsonschema:"New trigger conditions (full replace of the list); omit to keep current"`
	Roots        *[]string                 `json:"roots,omitempty" jsonschema:"New root node ids; omit to keep current (deleted node ids are always pruned)"`
	Nodes        map[string]dto.FlowNodeVM `json:"nodes,omitempty" jsonschema:"Nodes to add or overwrite, keyed by id; merged over the current nodes, not a replacement"`
	DeletedNodes []string                  `json:"deleted_nodes,omitempty" jsonschema:"Node ids to remove; their edges are scrubbed automatically. Must not overlap nodes"`
	MaxDepth     *int                      `json:"max_depth,omitempty" jsonschema:"New max depth; omit to keep current"`
	Active       *bool                     `json:"active,omitempty" jsonschema:"New enabled state; omit to keep current (or use soar.rule.set_enabled)"`
}

type soarRuleRelPathInput struct {
	RelPath string `json:"rel_path"`
}

type soarRuleSetEnabledInput struct {
	RelPath string `json:"rel_path"`
	Enabled bool   `json:"enabled"`
}

type soarRuleListInput struct {
	RuleName    string `json:"name,omitempty"`
	RuleActive  *bool  `json:"active,omitempty"`
	CreatedBy   string `json:"created_by,omitempty"`
	SystemOwner *bool  `json:"system_owner,omitempty"`
	Page        int    `json:"page,omitempty"`
	Size        int    `json:"size,omitempty"`
}

func registerSOARRules(m *Module) {
	uc := m.deps.SOAR.GetRuleUsecase()

	Add(m, &mcp.Tool{
		Name: "soar.flow.create", Title: "Create SOAR rule",
		Description: `Create a SOAR rule: a DAG of nodes that runs when an alert matches ALL trigger conditions. ` +
			`See mcp://utmstack/docs/soar-flow-guide for the full authoring guide. ` +
			`Node params shape depends on the executor type — params is a plain object: ` +
			`shell: (none, use node-level command/shell/platform/agent); ` +
			`http: {"method","url","headers","body"?}; ` +
			`conditional: {"conditions":[{"operator","field","value"?}]}; ` +
			`llm_enrich/llm_action: {"prompt"}; ` +
			`notify: {"message","type":"INFO"|"WARNING"|"ERROR"}; ` +
			`incident: {"name","description"}; ` +
			`mail: {"to","cc","subject","body"}. ` +
			`Live executor types: call soar.node_types. ` +
			`Create with active=false unless the user explicitly asks to enable it.`,
	}, Gate{Permission: "soar.write"},
		func(ctx context.Context, actor *authz.Actor, in soarRuleCreateInput) (any, error) {
			if len(in.Conditions) == 0 || len(in.Roots) == 0 || len(in.Nodes) == 0 {
				return nil, fmt.Errorf("conditions, roots, and nodes are required")
			}
			active := in.Active
			return uc.Create(ctx, dto.CreateRuleRequest{
				Name: in.Name, Description: in.Description,
				Conditions: in.Conditions, Roots: in.Roots, Nodes: in.Nodes,
				MaxDepth: in.MaxDepth, Active: &active,
			}, actor.Email)
		})

	Add(m, &mcp.Tool{
		Name: "soar.rule.update", Title: "Update SOAR rule",
		Description: "PARTIAL update: send only what changes. Omitted fields keep their current value; nodes are merged by id; deleted_nodes removes nodes and scrubs their edges. " +
			"Read the flow with soar.rule.get first, then send just the changed pieces. Never resend the whole flow.",
	}, Gate{Permission: "soar.write"},
		func(ctx context.Context, actor *authz.Actor, in soarRuleUpdateInput) (any, error) {
			cur, err := uc.Get(ctx, in.RelPath)
			if err != nil {
				return nil, err
			}
			req, err := mergeFlowUpdate(cur, in)
			if err != nil {
				return nil, err
			}
			return uc.Update(ctx, in.RelPath, req, actor.Email)
		})

	Add(m, &mcp.Tool{
		Name: "soar.flow.get", Title: "Get SOAR rule",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(ctx context.Context, _ *authz.Actor, in soarRuleRelPathInput) (any, error) {
			return uc.Get(ctx, in.RelPath)
		})

	Add(m, &mcp.Tool{
		Name: "soar.flow.delete", Title: "Delete SOAR rule",
	}, Gate{Permission: "soar.write"},
		func(ctx context.Context, _ *authz.Actor, in soarRuleRelPathInput) (any, error) {
			if err := uc.Delete(ctx, in.RelPath); err != nil {
				return nil, err
			}
			return map[string]any{"rel_path": in.RelPath, "deleted": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "soar.flow.set_enabled", Title: "Enable/disable SOAR rule",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "soar.write"},
		func(ctx context.Context, _ *authz.Actor, in soarRuleSetEnabledInput) (any, error) {
			if err := uc.SetEnabled(ctx, in.RelPath, in.Enabled); err != nil {
				return nil, err
			}
			return map[string]any{"rel_path": in.RelPath, "enabled": in.Enabled}, nil
		})

	Add(m, &mcp.Tool{
		Name: "soar.flow.list", Title: "List SOAR rules",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(ctx context.Context, _ *authz.Actor, in soarRuleListInput) (any, error) {
			f := dto.RuleFilters{
				RuleName: in.RuleName, RuleActive: in.RuleActive,
				CreatedBy: in.CreatedBy, SystemOwner: in.SystemOwner,
				Params: database.Params{Page: in.Page, Size: clampPageSize(in.Size)},
			}
			return uc.List(ctx, f)
		})

	Add(m, &mcp.Tool{
		Name: "soar.node_types", Title: "List SOAR flow node types",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(ctx context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			return map[string]any{
				"kinds":          []string{"executor", "enrichment"},
				"executor_types": m.deps.SOAR.GetExecutorTypes(),
			}, nil
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

// ---- prompts ---------------------------------------------------------------

func registerSOARPrompts(m *Module) {
	m.server.AddPrompt(&mcp.Prompt{
		Name: "soc.draft-soar-rule", Title: "Draft SOAR rule",
		Description: "Guides the model through SOAR flow construction: list existing flows to start from, list variables, propose filter conditions, then create the flow on confirmation.",
		Arguments: []*mcp.PromptArgument{{
			Name: "goal", Title: "Goal",
			Description: "Plain-language goal for the rule (e.g. 'isolate hosts on Mimikatz detections')",
		}},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		goal := req.Params.Arguments["goal"]
		if goal == "" {
			return nil, fmt.Errorf("argument 'goal' is required")
		}
		text := fmt.Sprintf(`Draft or modify a SOAR rule for this goal: %s

Follow these steps, showing results as you go:
1. Call soar.rule.list to find existing flows that could be a starting point.
2. If you will modify an existing flow, call soar.rule.get(rel_path) to read it.
3. Call soar.variable.list to see available incident variables.
4. Call soar.rule.resolve_filter_values to suggest valid filter fields/values.
5. Draft the change and present it to the user before touching anything.
6. Only after explicit user confirmation, apply it:
   - new flow → soar.rule.create with the full flow and active=false;
   - existing flow → soar.rule.update with PARTIAL fields only: omit what does
     not change, use nodes to add/overwrite by id and deleted_nodes to remove.
     Never resend the whole flow in an update.

Never enable a rule without asking first.`, goal)
		msg := &mcp.PromptMessage{Role: "user", Content: &mcp.TextContent{Text: text}}
		return &mcp.GetPromptResult{Messages: []*mcp.PromptMessage{msg}}, nil
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
				StartedAtGTE: in.DateGTE, StartedAtLTE: in.DateLTE,
				Params: database.Params{Page: in.Page, Size: clampPageSize(in.Size)},
			})
		})
}

// ---- soar.variable.* -------------------------------------------------------

type soarVariableCreateInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Value       string  `json:"value"`
	IsSecret    bool    `json:"is_secret,omitempty"`
}

type soarVariableUpdateInput struct {
	ID          uuid.UUID `json:"id"`
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Value       *string   `json:"value,omitempty"`
	IsSecret    bool      `json:"is_secret,omitempty"`
}

type soarVariableListInput struct {
	Name *string `json:"name,omitempty"`
	Page int     `json:"page,omitempty"`
	Size int     `json:"size,omitempty"`
}

type idUUIDInput struct {
	ID uuid.UUID `json:"id"`
}

func registerSOARVariables(m *Module) {
	uc := m.deps.SOAR.GetVariableUsecase()

	Add(m, &mcp.Tool{
		Name: "soar.variable.create", Title: "Create SOAR variable",
	}, Gate{Permission: "soar.write"},
		func(ctx context.Context, actor *authz.Actor, in soarVariableCreateInput) (any, error) {
			return uc.Create(ctx, dto.CreateVariableRequest{
				Name: in.Name, Description: in.Description, Value: in.Value, IsSecret: in.IsSecret,
			}, actor.Email)
		})

	Add(m, &mcp.Tool{
		Name: "soar.variable.update", Title: "Update SOAR variable",
	}, Gate{Permission: "soar.write"},
		func(ctx context.Context, actor *authz.Actor, in soarVariableUpdateInput) (any, error) {
			return uc.Update(ctx, dto.UpdateVariableRequest{
				ID: in.ID, Name: in.Name, Description: in.Description, Value: in.Value, IsSecret: in.IsSecret,
			}, actor.Email)
		})

	Add(m, &mcp.Tool{
		Name: "soar.variable.get", Title: "Get SOAR variable",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(ctx context.Context, _ *authz.Actor, in idUUIDInput) (any, error) {
			return uc.FindByID(ctx, in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "soar.variable.list", Title: "List SOAR variables",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "soar.read"},
		func(ctx context.Context, _ *authz.Actor, in soarVariableListInput) (any, error) {
			items, total, err := uc.FindAll(ctx, dto.VariableFilter{
				Name:   in.Name,
				Params: database.Params{Page: in.Page, Size: clampPageSize(in.Size)},
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "soar.variable.delete", Title: "Delete SOAR variable",
	}, Gate{Permission: "soar.write"},
		func(ctx context.Context, _ *authz.Actor, in idUUIDInput) (any, error) {
			if err := uc.Delete(ctx, in.ID); err != nil {
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
