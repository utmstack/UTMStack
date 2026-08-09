package mcp

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

func toRaw(v any) json.RawMessage {
	if v == nil {
		return nil
	}
	b, _ := json.Marshal(v) // v came from json.Unmarshal into any, cannot fail
	return b
}

func registerEventProcessing(m *Module) {
	registerEPRegexPatterns(m)
	registerEPCorrelationRules(m)
	registerEPFilters(m)
	registerEPIngestionStats(m)
}

// ---- regex_pattern.* -------------------------------------------------------

type epRegexListInput struct {
	Search string `json:"search,omitempty"`
	Page   int    `json:"page,omitempty"`
	Size   int    `json:"size,omitempty"`
}

type epRegexIDInput struct {
	PatternID string `json:"pattern_id"`
}

func registerEPRegexPatterns(m *Module) {
	uc := m.deps.EventProcessing.GetRegexPatternUsecase()

	// Read-only surface: patterns are a shared vocabulary seeded by the pipeline
	// bootstrap, so there are no create/update/delete tools.
	Add(m, &mcp.Tool{
		Name: "regex_pattern.list", Title: "List regex patterns",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "eventprocessing.read"},
		func(ctx context.Context, _ *authz.Actor, in epRegexListInput) (any, error) {
			return uc.List(ctx, dto.RegexPatternFilters{
				Search: in.Search, Page: in.Page, Size: clampPageSize(in.Size),
			})
		})

	Add(m, &mcp.Tool{
		Name: "regex_pattern.get", Title: "Get regex pattern",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "eventprocessing.read"},
		func(ctx context.Context, _ *authz.Actor, in epRegexIDInput) (any, error) {
			return uc.GetByID(ctx, in.PatternID)
		})
}

type epAssetNameInput struct {
	AssetName string `json:"asset_name"`
}


// ---- correlation_rule.* ----------------------------------------------------

type epRuleListInput struct {
	RuleName    string   `json:"rule_name,omitempty"`
	RuleActive  *bool    `json:"active,omitempty"`
	Categories  []string `json:"categories,omitempty"`
	Adversaries []string `json:"adversaries,omitempty"`
	SystemOwner *bool    `json:"system_owner,omitempty"`
	Search      string   `json:"search,omitempty"`
	Page        int      `json:"page,omitempty"`
	Size        int      `json:"size,omitempty"`
}

type epRuleRelPathInput struct {
	RelPath string `json:"rel_path"`
}

type epRuleSetActiveInput struct {
	RelPath string `json:"rel_path"`
	Active  bool   `json:"active"`
}

type epRulePropertyValuesInput struct {
	Property string `json:"property"`
	Value    string `json:"value,omitempty"`
}

// epRuleCreateInput mirrors dto.CreateCorrelationRuleRequest for MCP tool
// registration. The five DSL fields are typed as `any` so the generated JSON
// schema advertises "anything goes" instead of "array of uint8" (the default
// for json.RawMessage under reflection). Values are re-marshaled via toRaw
// before being handed to the usecase.
type epRuleCreateInput struct {
	Name            string `json:"name"`
	Adversary       string `json:"adversary,omitempty"`
	Confidentiality int    `json:"confidentiality"`
	Integrity       int    `json:"integrity"`
	Availability    int    `json:"availability"`
	Category        string `json:"category,omitempty"`
	Technique       string `json:"technique,omitempty"`
	Description     string `json:"description,omitempty"`

	References    any `json:"references,omitempty" jsonschema:"Rule references (any JSON)"`
	Definition    any `json:"definition" jsonschema:"Rule definition DSL (any JSON)"`
	GroupBy       any `json:"groupBy,omitempty" jsonschema:"Group-by DSL (any JSON)"`
	DeduplicateBy any `json:"deduplicateBy,omitempty" jsonschema:"Deduplicate-by DSL (any JSON)"`
	Correlation   any `json:"correlation,omitempty" jsonschema:"Correlation DSL (any JSON)"`

	RuleActive bool              `json:"ruleActive"`
	DataTypes  []dto.DataTypeRef `json:"dataTypes,omitempty"`
}

type epRuleUpdateInput struct {
	RelPath string `json:"relPath"`
	epRuleCreateInput
}

func (in epRuleCreateInput) toDTO() dto.CreateCorrelationRuleRequest {
	return dto.CreateCorrelationRuleRequest{
		RuleName:            in.Name,
		RuleAdversary:       in.Adversary,
		RuleConfidentiality: in.Confidentiality,
		RuleIntegrity:       in.Integrity,
		RuleAvailability:    in.Availability,
		RuleCategory:        in.Category,
		RuleTechnique:       in.Technique,
		RuleDescription:     in.Description,
		RuleReferencesDef:   toRaw(in.References),
		RuleDefinitionDef:   toRaw(in.Definition),
		RuleGroupByDef:      toRaw(in.GroupBy),
		DeduplicateByDef:    toRaw(in.DeduplicateBy),
		CorrelationDef:      toRaw(in.Correlation),
		RuleActive:          in.RuleActive,
		DataTypes:           in.DataTypes,
	}
}

func (in epRuleUpdateInput) toDTO() dto.UpdateCorrelationRuleRequest {
	c := in.epRuleCreateInput.toDTO()
	return dto.UpdateCorrelationRuleRequest{
		RelPath:             in.RelPath,
		RuleName:            c.RuleName,
		RuleAdversary:       c.RuleAdversary,
		RuleConfidentiality: c.RuleConfidentiality,
		RuleIntegrity:       c.RuleIntegrity,
		RuleAvailability:    c.RuleAvailability,
		RuleCategory:        c.RuleCategory,
		RuleTechnique:       c.RuleTechnique,
		RuleDescription:     c.RuleDescription,
		RuleReferencesDef:   c.RuleReferencesDef,
		RuleDefinitionDef:   c.RuleDefinitionDef,
		RuleGroupByDef:      c.RuleGroupByDef,
		DeduplicateByDef:    c.DeduplicateByDef,
		CorrelationDef:      c.CorrelationDef,
		RuleActive:          c.RuleActive,
		DataTypes:           c.DataTypes,
	}
}

func registerEPCorrelationRules(m *Module) {
	uc := m.deps.EventProcessing.GetCorrelationRuleUsecase()

	Add(m, &mcp.Tool{
		Name: "correlation_rule.create", Title: "Create correlation rule",
	}, Gate{Permission: "eventprocessing.write"},
		func(ctx context.Context, _ *authz.Actor, in epRuleCreateInput) (any, error) {
			if err := uc.Create(ctx, in.toDTO()); err != nil {
				return nil, err
			}
			return map[string]any{"name": in.Name, "created": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "correlation_rule.update", Title: "Update correlation rule",
	}, Gate{Permission: "eventprocessing.write"},
		func(ctx context.Context, _ *authz.Actor, in epRuleUpdateInput) (any, error) {
			if err := uc.Update(ctx, in.toDTO()); err != nil {
				return nil, err
			}
			return map[string]any{"rel_path": in.RelPath, "updated": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "correlation_rule.list", Title: "List correlation rules",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "eventprocessing.read"},
		func(ctx context.Context, _ *authz.Actor, in epRuleListInput) (any, error) {
			return uc.List(ctx, dto.CorrelationRuleFilters{
				RuleName: in.RuleName, RuleActive: in.RuleActive,
				RuleCategory: in.Categories, RuleAdversary: in.Adversaries,
				SystemOwner: in.SystemOwner, Search: in.Search,
				Page: in.Page, Size: clampPageSize(in.Size),
			})
		})

	Add(m, &mcp.Tool{
		Name: "correlation_rule.get", Title: "Get correlation rule",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "eventprocessing.read"},
		func(ctx context.Context, _ *authz.Actor, in epRuleRelPathInput) (any, error) {
			return uc.GetByRelPath(ctx, in.RelPath)
		})

	Add(m, &mcp.Tool{
		Name: "correlation_rule.delete", Title: "Delete correlation rule",
	}, Gate{Permission: "eventprocessing.write"},
		func(ctx context.Context, _ *authz.Actor, in epRuleRelPathInput) (any, error) {
			if err := uc.Delete(ctx, in.RelPath); err != nil {
				return nil, err
			}
			return map[string]any{"rel_path": in.RelPath, "deleted": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "correlation_rule.set_active", Title: "Activate or deactivate correlation rule",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "eventprocessing.write"},
		func(ctx context.Context, _ *authz.Actor, in epRuleSetActiveInput) (any, error) {
			changed, err := uc.SetActive(ctx, in.RelPath, in.Active)
			if err != nil {
				return nil, err
			}
			return map[string]any{"rel_path": in.RelPath, "active": in.Active, "changed": changed}, nil
		})

	Add(m, &mcp.Tool{
		Name: "correlation_rule.search_property_values", Title: "Search distinct property values",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "eventprocessing.read"},
		func(ctx context.Context, _ *authz.Actor, in epRulePropertyValuesInput) (any, error) {
			return uc.FindDistinctPropertyValues(ctx, in.Property, in.Value)
		})
}

// ---- filter.* --------------------------------------------------------------

type epFilterUpsertInput struct {
	RelPath string `json:"rel_path"`
	Content string `json:"content"`
}

type epFilterListInput struct {
	RelPathContains *string `json:"rel_path_contains,omitempty"`
	Active          *bool   `json:"active,omitempty"`
	System          *bool   `json:"system,omitempty"`
	Page            int     `json:"page,omitempty"`
	Size            int     `json:"size,omitempty"`
}

func registerEPFilters(m *Module) {
	uc := m.deps.EventProcessing.GetFilterUsecase()

	Add(m, &mcp.Tool{
		Name: "filter.create", Title: "Create logstash filter",
	}, Gate{Permission: "eventprocessing.write"},
		func(ctx context.Context, _ *authz.Actor, in epFilterUpsertInput) (any, error) {
			return uc.Create(ctx, dto.CreateFilterRequest{RelPath: in.RelPath, Content: in.Content})
		})

	Add(m, &mcp.Tool{
		Name: "filter.update", Title: "Update logstash filter",
	}, Gate{Permission: "eventprocessing.write"},
		func(ctx context.Context, _ *authz.Actor, in epFilterUpsertInput) (any, error) {
			return uc.Update(ctx, dto.UpdateFilterRequest{RelPath: in.RelPath, Content: in.Content})
		})

	Add(m, &mcp.Tool{
		Name: "filter.list", Title: "List logstash filters",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "eventprocessing.read"},
		func(ctx context.Context, _ *authz.Actor, in epFilterListInput) (any, error) {
			items, total, err := uc.List(ctx, dto.FilterFilters{
				RelPathContains: in.RelPathContains, IsActiveEq: in.Active, SystemEq: in.System,
				Page: in.Page, Size: clampPageSize(in.Size),
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "filter.get", Title: "Get logstash filter",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "eventprocessing.read"},
		func(ctx context.Context, _ *authz.Actor, in epRuleRelPathInput) (any, error) {
			return uc.GetByRelPath(ctx, in.RelPath)
		})

	Add(m, &mcp.Tool{
		Name: "filter.delete", Title: "Delete logstash filter",
	}, Gate{Permission: "eventprocessing.write"},
		func(ctx context.Context, _ *authz.Actor, in epRuleRelPathInput) (any, error) {
			if err := uc.Delete(ctx, in.RelPath); err != nil {
				return nil, err
			}
			return map[string]any{"rel_path": in.RelPath, "deleted": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "filter.set_active", Title: "Activate or deactivate filter",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "eventprocessing.write"},
		func(ctx context.Context, _ *authz.Actor, in epRuleSetActiveInput) (any, error) {
			if err := uc.SetActive(ctx, in.RelPath, in.Active); err != nil {
				return nil, err
			}
			return map[string]any{"rel_path": in.RelPath, "active": in.Active}, nil
		})
}

// ---- ingestion_stats.* -----------------------------------------------------

type epIngestionTotalsInput struct {
	GroupBy string `json:"group_by,omitempty"`
	Status  string `json:"status,omitempty"`
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
	Top     int    `json:"top,omitempty"`
}

type epIngestionTimelineInput struct {
	GroupBy    string `json:"group_by,omitempty"`
	Status     string `json:"status,omitempty"`
	Interval   string `json:"interval,omitempty"`
	From       string `json:"from,omitempty"`
	To         string `json:"to,omitempty"`
	Top        int    `json:"top,omitempty"`
	DataSource string `json:"data_source,omitempty"`
}

func registerEPIngestionStats(m *Module) {
	uc := m.deps.EventProcessing.GetIngestionStatsUsecase()

	Add(m, &mcp.Tool{
		Name: "ingestion_stats.totals", Title: "Ingestion totals",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "eventprocessing.read"},
		func(ctx context.Context, _ *authz.Actor, in epIngestionTotalsInput) (any, error) {
			return uc.Totals(ctx, in.GroupBy, in.Status, in.From, in.To, in.Top)
		})

	Add(m, &mcp.Tool{
		Name: "ingestion_stats.timeline", Title: "Ingestion timeline",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "eventprocessing.read"},
		func(ctx context.Context, _ *authz.Actor, in epIngestionTimelineInput) (any, error) {
			return uc.Timeline(ctx, in.GroupBy, in.Status, in.Interval, in.From, in.To, in.Top, in.DataSource)
		})
}
