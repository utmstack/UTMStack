package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

// registerAlerts exposes the alerts module's usecases as MCP tools.
//
// Tool naming follows `<module>.<resource>.<verb>` (snake_case for verbs).
// Annotations.ReadOnlyHint=true means the tool does not mutate state;
// destructiveHint is left at its default (true) for write tools so MCP
// clients can request explicit confirmation.
func registerAlerts(m *Module) {
	registerAlertActions(m)
	registerAlertTags(m)
	registerAlertTagRules(m)
	registerAlertAdversary(m)
}

// ------------------------------------------------------------------
// alerts.* — operate on individual alert documents (live in OpenSearch).
// ------------------------------------------------------------------

type alertsUpdateStatusInput struct {
	AlertIDs            []string `json:"alert_ids" jsonschema:"OpenSearch _id values of the alerts to update"`
	Status              int      `json:"status" jsonschema:"Target status code (UTMStack convention: 0=open 1=in-review 2=closed)"`
	StatusObservation   string   `json:"status_observation,omitempty" jsonschema:"Optional analyst note attached to the status change"`
	AddFalsePositiveTag bool     `json:"add_false_positive_tag,omitempty" jsonschema:"If true, also apply the FALSE_POSITIVE tag (used when closing as FP)"`
}

type alertsUpdateNotesInput struct {
	AlertID string `json:"alert_id" jsonschema:"OpenSearch _id of the alert"`
	Notes   string `json:"notes" jsonschema:"Analyst notes to attach (replaces existing notes)"`
}

type alertsUpdateTagsInput struct {
	AlertIDs   []string `json:"alert_ids" jsonschema:"Alert _id values to tag"`
	Tags       []string `json:"tags" jsonschema:"Tag names to apply"`
	CreateRule bool     `json:"create_rule,omitempty" jsonschema:"If true, also persist a tag rule so future matching alerts are auto-tagged"`
}

type alertsConvertToIncidentInput struct {
	AlertIDs       []string `json:"alert_ids" jsonschema:"_id values of the alerts to attach to the incident"`
	IncidentName   string   `json:"incident_name" jsonschema:"Display name for the new/target incident"`
	IncidentID     int      `json:"incident_id" jsonschema:"Numeric ID of the incident to attach to (0 = create new)"`
	IncidentSource string   `json:"incident_source,omitempty"`
}

type alertsEmptyInput struct{}

func registerAlertActions(m *Module) {
	uc := m.deps.Alerts.GetAlertUsecase()

	Add(m, &mcp.Tool{
		Name:        "alerts.update_status",
		Title:       "Update alert status",
		Description: "Change the status of one or more alerts. Requires the alerts.write permission. Idempotent — repeating the same call has no additional effect.",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "alerts.write"},
		func(ctx context.Context, actor *authz.Actor, in alertsUpdateStatusInput) (any, error) {
			if len(in.AlertIDs) == 0 {
				return nil, fmt.Errorf("alert_ids: must provide at least one alert id")
			}
			err := uc.UpdateStatus(ctx, actor.Login, dto.UpdateAlertStatusRequest{
				AlertIDs:            in.AlertIDs,
				Status:              in.Status,
				StatusObservation:   in.StatusObservation,
				AddFalsePositiveTag: in.AddFalsePositiveTag,
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"updated": len(in.AlertIDs)}, nil
		})

	Add(m, &mcp.Tool{
		Name:        "alerts.update_notes",
		Title:       "Replace alert notes",
		Description: "Replace the analyst notes attached to an alert. Requires alerts.write.",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "alerts.write"},
		func(ctx context.Context, actor *authz.Actor, in alertsUpdateNotesInput) (any, error) {
			if in.AlertID == "" {
				return nil, fmt.Errorf("alert_id is required")
			}
			if err := uc.UpdateNotes(ctx, actor.Login, in.AlertID, in.Notes); err != nil {
				return nil, err
			}
			return map[string]any{"alert_id": in.AlertID}, nil
		})

	Add(m, &mcp.Tool{
		Name:        "alerts.update_tags",
		Title:       "Set tags on alerts",
		Description: "Replace the tag list on one or more alerts. If create_rule is true, also persists a tag rule so future matching alerts receive the same tags automatically. Requires alerts.write.",
	}, Gate{Permission: "alerts.write"},
		func(ctx context.Context, actor *authz.Actor, in alertsUpdateTagsInput) (any, error) {
			if len(in.AlertIDs) == 0 {
				return nil, fmt.Errorf("alert_ids: must provide at least one alert id")
			}
			err := uc.UpdateTags(ctx, actor.Login, dto.UpdateAlertTagsRequest{
				AlertIDs:   in.AlertIDs,
				Tags:       in.Tags,
				CreateRule: in.CreateRule,
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"updated": len(in.AlertIDs)}, nil
		})

	Add(m, &mcp.Tool{
		Name:        "alerts.convert_to_incident",
		Title:       "Attach alerts to an incident",
		Description: "Convert/attach a set of alerts to an incident — either by creating a new one (incident_id=0) or by appending to an existing one. Requires alerts.write.",
	}, Gate{Permission: "alerts.write"},
		func(ctx context.Context, actor *authz.Actor, in alertsConvertToIncidentInput) (any, error) {
			if len(in.AlertIDs) == 0 || in.IncidentName == "" {
				return nil, fmt.Errorf("alert_ids and incident_name are required")
			}
			err := uc.ConvertToIncident(ctx, actor.Login, dto.ConvertToIncidentRequest{
				AlertIDs:       in.AlertIDs,
				IncidentName:   in.IncidentName,
				IncidentID:     in.IncidentID,
				IncidentSource: in.IncidentSource,
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"alerts_attached": len(in.AlertIDs), "incident_id": in.IncidentID}, nil
		})

	readOnly := true
	Add(m, &mcp.Tool{
		Name:        "alerts.count_open",
		Title:       "Count open alerts",
		Description: "Return the current count of alerts in the OPEN status across all indices.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: readOnly},
	}, Gate{Permission: "alerts.read"},
		func(ctx context.Context, _ *authz.Actor, _ alertsEmptyInput) (any, error) {
			return uc.CountOpenAlerts(ctx)
		})
}

// ------------------------------------------------------------------
// alert_tags.* — CRUD for the user-defined tag catalog (Postgres-backed).
// ------------------------------------------------------------------

type alertTagsCreateInput struct {
	TagName  string `json:"tag_name" jsonschema:"Tag name (unique)"`
	TagColor string `json:"tag_color,omitempty" jsonschema:"Optional CSS color (e.g. #ff0000)"`
}
type alertTagsUpdateInput struct {
	ID          uint64 `json:"id"`
	TagName     string `json:"tag_name"`
	TagColor    string `json:"tag_color,omitempty"`
	SystemOwner bool   `json:"system_owner,omitempty"`
}
type alertTagsListInput struct {
	TagName     *string `json:"tag_name,omitempty" jsonschema:"Filter by partial tag name"`
	SystemOwner *bool   `json:"system_owner,omitempty"`
	Page        int     `json:"page,omitempty"`
	Size        int     `json:"size,omitempty"`
}
type alertTagsByIDInput struct {
	ID uint64 `json:"id"`
}

func registerAlertTags(m *Module) {
	uc := m.deps.Alerts.GetAlertTagUsecase()

	Add(m, &mcp.Tool{
		Name:        "alert_tags.create",
		Title:       "Create alert tag",
		Description: "Create a new alert tag in the user catalog.",
	}, Gate{Permission: "alerts.write"},
		func(ctx context.Context, _ *authz.Actor, in alertTagsCreateInput) (any, error) {
			return uc.Create(ctx, dto.CreateAlertTagRequest{TagName: in.TagName, TagColor: in.TagColor})
		})

	Add(m, &mcp.Tool{
		Name:        "alert_tags.update",
		Title:       "Update alert tag",
		Description: "Update an existing alert tag.",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "alerts.write"},
		func(ctx context.Context, _ *authz.Actor, in alertTagsUpdateInput) (any, error) {
			return uc.Update(ctx, dto.UpdateAlertTagRequest{
				ID: in.ID, TagName: in.TagName, TagColor: in.TagColor, SystemOwner: in.SystemOwner,
			})
		})

	Add(m, &mcp.Tool{
		Name:        "alert_tags.list",
		Title:       "List alert tags",
		Description: "List alert tags from the catalog, paginated.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "alerts.read"},
		func(ctx context.Context, _ *authz.Actor, in alertTagsListInput) (any, error) {
			items, total, err := uc.List(ctx, dto.AlertTagFilters{
				TagName: in.TagName, SystemOwner: in.SystemOwner,
				Page: in.Page, Size: clampPageSize(in.Size),
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name:        "alert_tags.get",
		Title:       "Get alert tag by ID",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "alerts.read"},
		func(ctx context.Context, _ *authz.Actor, in alertTagsByIDInput) (any, error) {
			return uc.GetByID(ctx, in.ID)
		})

	Add(m, &mcp.Tool{
		Name:        "alert_tags.delete",
		Title:       "Delete alert tag",
		Description: "Delete an alert tag from the catalog. Existing alerts that reference the tag retain it.",
	}, Gate{Permission: "alerts.write"},
		func(ctx context.Context, _ *authz.Actor, in alertTagsByIDInput) (any, error) {
			if err := uc.Delete(ctx, in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})
}

// ------------------------------------------------------------------
// alert_tag_rules.* — auto-tagging rules.
// ------------------------------------------------------------------

type tagRuleRef struct {
	ID          uint64 `json:"id"`
	TagName     string `json:"tag_name,omitempty"`
	TagColor    string `json:"tag_color,omitempty"`
	SystemOwner bool   `json:"system_owner,omitempty"`
}
type alertTagRulesCreateInput struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Conditions  []common_models.FilterType `json:"conditions" jsonschema:"Filter DSL conditions; see mcp://utmstack/docs/filter-operators"`
	Tags        []tagRuleRef               `json:"tags"`
}
type alertTagRulesUpdateInput struct {
	ID          uint64                     `json:"id"`
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Conditions  []common_models.FilterType `json:"conditions"`
	Tags        []tagRuleRef               `json:"tags"`
}
type alertTagRulesListInput struct {
	Name           *string `json:"name,omitempty"`
	ConditionField *string `json:"condition_field,omitempty"`
	ConditionValue *string `json:"condition_value,omitempty"`
	RuleActive     *bool   `json:"active,omitempty"`
	TagIDs         []int64 `json:"tag_ids,omitempty"`
	Page           int     `json:"page,omitempty"`
	Size           int     `json:"size,omitempty"`
}
type alertTagRulesByIDInput struct {
	ID uint64 `json:"id"`
}

func registerAlertTagRules(m *Module) {
	uc := m.deps.Alerts.GetAlertTagRuleUsecase()

	Add(m, &mcp.Tool{
		Name:        "alert_tag_rules.create",
		Title:       "Create alert tag rule",
		Description: "Create a rule that auto-applies tags to alerts matching the given conditions.",
	}, Gate{Permission: "alerts.write"},
		func(ctx context.Context, _ *authz.Actor, in alertTagRulesCreateInput) (any, error) {
			return uc.Create(ctx, dto.CreateAlertTagRuleRequest{
				Name: in.Name, Description: in.Description,
				Conditions: in.Conditions, Tags: toTagRefs(in.Tags),
			})
		})

	Add(m, &mcp.Tool{
		Name:        "alert_tag_rules.update",
		Title:       "Update alert tag rule",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "alerts.write"},
		func(ctx context.Context, _ *authz.Actor, in alertTagRulesUpdateInput) (any, error) {
			return uc.Update(ctx, dto.UpdateAlertTagRuleRequest{
				ID: in.ID, Name: in.Name, Description: in.Description,
				Conditions: in.Conditions, Tags: toTagRefs(in.Tags),
			})
		})

	Add(m, &mcp.Tool{
		Name:        "alert_tag_rules.list",
		Title:       "List alert tag rules",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "alerts.read"},
		func(ctx context.Context, _ *authz.Actor, in alertTagRulesListInput) (any, error) {
			items, total, err := uc.List(ctx, dto.AlertTagRuleFilters{
				Name: in.Name, ConditionField: in.ConditionField, ConditionValue: in.ConditionValue,
				RuleActive: in.RuleActive, TagIDs: in.TagIDs,
				Page: in.Page, Size: clampPageSize(in.Size),
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name:        "alert_tag_rules.get",
		Title:       "Get alert tag rule by ID",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "alerts.read"},
		func(ctx context.Context, _ *authz.Actor, in alertTagRulesByIDInput) (any, error) {
			return uc.GetByID(ctx, in.ID)
		})

	Add(m, &mcp.Tool{
		Name:        "alert_tag_rules.delete",
		Title:       "Delete alert tag rule",
		Description: "Soft-delete a tag rule. Already-tagged alerts retain their tags.",
	}, Gate{Permission: "alerts.write"},
		func(ctx context.Context, _ *authz.Actor, in alertTagRulesByIDInput) (any, error) {
			if err := uc.Delete(ctx, in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})
}

// ------------------------------------------------------------------
// adversary.* — grouped adversary view across alerts.
// ------------------------------------------------------------------

type adversarySearchInput struct {
	Filters []common_models.FilterType `json:"filters,omitempty" jsonschema:"Optional structured filters (see filter-operators doc). When empty, returns all adversaries within the default time window."`
}

func registerAlertAdversary(m *Module) {
	uc := m.deps.Alerts.GetAdversaryUsecase()
	Add(m, &mcp.Tool{
		Name:        "alerts.search_adversary",
		Title:       "Search alerts grouped by adversary",
		Description: "Return alerts grouped by adversary host. Each entry surfaces the parent alert plus its correlated children.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "alerts.read"},
		func(ctx context.Context, _ *authz.Actor, in adversarySearchInput) (any, error) {
			return uc.FetchAdversaryAlerts(ctx, in.Filters)
		})
}

// ------------------------------------------------------------------
// helpers
// ------------------------------------------------------------------

func toTagRefs(in []tagRuleRef) []dto.AlertTagRuleTagRef {
	out := make([]dto.AlertTagRuleTagRef, len(in))
	for i, t := range in {
		out[i] = dto.AlertTagRuleTagRef{
			ID: t.ID, TagName: t.TagName, TagColor: t.TagColor, SystemOwner: t.SystemOwner,
		}
	}
	return out
}
