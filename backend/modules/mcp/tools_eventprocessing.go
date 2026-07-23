package mcp

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

func registerEventProcessing(m *Module) {
	registerEPRegexPatterns(m)
	registerEPTenantConfigs(m)
	registerEPCorrelationRules(m)
	registerEPFilters(m)
	registerEPIngestionStats(m)
}

// ---- regex_pattern.* -------------------------------------------------------

type epRegexCreateInput struct {
	PatternID          string `json:"pattern_id"`
	PatternDescription string `json:"pattern_description,omitempty"`
	PatternDefinition  string `json:"pattern_definition"`
}

type epRegexUpdateInput struct {
	PatternID          string `json:"pattern_id"`
	PatternDescription string `json:"pattern_description,omitempty"`
	PatternDefinition  string `json:"pattern_definition"`
}

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

	Add(m, &mcp.Tool{
		Name: "regex_pattern.create", Title: "Create regex pattern",
	}, Gate{Permission: "eventprocessing.write"},
		func(ctx context.Context, _ *authz.Actor, in epRegexCreateInput) (any, error) {
			return uc.Create(ctx, dto.CreateRegexPatternRequest{
				PatternID: in.PatternID, PatternDescription: in.PatternDescription, PatternDefinition: in.PatternDefinition,
			})
		})

	Add(m, &mcp.Tool{
		Name: "regex_pattern.update", Title: "Update regex pattern",
	}, Gate{Permission: "eventprocessing.write"},
		func(ctx context.Context, _ *authz.Actor, in epRegexUpdateInput) (any, error) {
			return uc.Update(ctx, dto.UpdateRegexPatternRequest{
				PatternID: in.PatternID, PatternDescription: in.PatternDescription, PatternDefinition: in.PatternDefinition,
			})
		})

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

	Add(m, &mcp.Tool{
		Name: "regex_pattern.delete", Title: "Delete regex pattern",
	}, Gate{Permission: "eventprocessing.write"},
		func(ctx context.Context, _ *authz.Actor, in epRegexIDInput) (any, error) {
			if err := uc.Delete(ctx, in.PatternID); err != nil {
				return nil, err
			}
			return map[string]any{"pattern_id": in.PatternID, "deleted": true}, nil
		})
}

// ---- tenant_config.* -------------------------------------------------------

type epTenantConfigUpsertInput struct {
	AssetName            string   `json:"asset_name"`
	AssetHostnameList    []string `json:"asset_hostname_list,omitempty"`
	AssetIpList          []string `json:"asset_ip_list,omitempty"`
	AssetConfidentiality int      `json:"asset_confidentiality,omitempty"`
	AssetIntegrity       int      `json:"asset_integrity,omitempty"`
	AssetAvailability    int      `json:"asset_availability,omitempty"`
}

type epTenantConfigListInput struct {
	Search string `json:"search,omitempty"`
	Page   int    `json:"page,omitempty"`
	Size   int    `json:"size,omitempty"`
}

type epAssetNameInput struct {
	AssetName string `json:"asset_name"`
}

func registerEPTenantConfigs(m *Module) {
	uc := m.deps.EventProcessing.GetTenantConfigUsecase()

	Add(m, &mcp.Tool{
		Name: "tenant_config.create", Title: "Create tenant config",
	}, Gate{Permission: "eventprocessing.write"},
		func(ctx context.Context, _ *authz.Actor, in epTenantConfigUpsertInput) (any, error) {
			return uc.Create(ctx, dto.CreateTenantConfigRequest{
				AssetName: in.AssetName, AssetHostnameList: in.AssetHostnameList, AssetIpList: in.AssetIpList,
				AssetConfidentiality: in.AssetConfidentiality, AssetIntegrity: in.AssetIntegrity, AssetAvailability: in.AssetAvailability,
			})
		})

	Add(m, &mcp.Tool{
		Name: "tenant_config.update", Title: "Update tenant config",
	}, Gate{Permission: "eventprocessing.write"},
		func(ctx context.Context, _ *authz.Actor, in epTenantConfigUpsertInput) (any, error) {
			return uc.Update(ctx, dto.UpdateTenantConfigRequest{
				AssetName: in.AssetName, AssetHostnameList: in.AssetHostnameList, AssetIpList: in.AssetIpList,
				AssetConfidentiality: in.AssetConfidentiality, AssetIntegrity: in.AssetIntegrity, AssetAvailability: in.AssetAvailability,
			})
		})

	Add(m, &mcp.Tool{
		Name: "tenant_config.list", Title: "List tenant configs",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "eventprocessing.read"},
		func(ctx context.Context, _ *authz.Actor, in epTenantConfigListInput) (any, error) {
			return uc.List(ctx, dto.TenantConfigFilters{Search: in.Search, Page: in.Page, Size: clampPageSize(in.Size)})
		})

	Add(m, &mcp.Tool{
		Name: "tenant_config.get", Title: "Get tenant config",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "eventprocessing.read"},
		func(ctx context.Context, _ *authz.Actor, in epAssetNameInput) (any, error) {
			return uc.GetByID(ctx, in.AssetName)
		})

	Add(m, &mcp.Tool{
		Name: "tenant_config.delete", Title: "Delete tenant config",
	}, Gate{Permission: "eventprocessing.write"},
		func(ctx context.Context, _ *authz.Actor, in epAssetNameInput) (any, error) {
			if err := uc.Delete(ctx, in.AssetName); err != nil {
				return nil, err
			}
			return map[string]any{"asset_name": in.AssetName, "deleted": true}, nil
		})
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

// MCP-only input shapes: the underlying DTO uses json.RawMessage for these
// fields, which the SDK reflects as []byte → array<integer> — wrong for the
// automation client. These wrappers carry the real per-field shape.
type epCorrelationRuleCreateInput struct {
	RuleName            string            `json:"name" jsonschema:"Rule name"`
	RuleAdversary       string            `json:"adversary" jsonschema:"origin | target"`
	RuleConfidentiality int               `json:"confidentiality" jsonschema:"0-3"`
	RuleIntegrity       int               `json:"integrity" jsonschema:"0-3"`
	RuleAvailability    int               `json:"availability" jsonschema:"0-3"`
	RuleCategory        string            `json:"category"`
	RuleTechnique       string            `json:"technique"`
	RuleDescription     string            `json:"description"`
	RuleReferences      []any             `json:"references,omitempty" jsonschema:"Reference URLs or objects"`
	RuleDefinition      string            `json:"definition" jsonschema:"CEL expression, e.g. equals(\"log.channel\",\"Security\")"`
	RuleGroupBy         []string          `json:"groupBy,omitempty"`
	DeduplicateBy       []string          `json:"deduplicateBy,omitempty"`
	Correlation         any               `json:"correlation,omitempty" jsonschema:"Optional multi-step correlation object"`
	RuleActive          bool              `json:"ruleActive"`
	DataTypes           []dto.DataTypeRef `json:"dataTypes"`
}

type epCorrelationRuleUpdateInput struct {
	RelPath             string            `json:"relPath" jsonschema:"Rule identity (YAML relative path)"`
	RuleName            string            `json:"name"`
	RuleAdversary       string            `json:"adversary" jsonschema:"origin | target"`
	RuleConfidentiality int               `json:"confidentiality" jsonschema:"0-3"`
	RuleIntegrity       int               `json:"integrity" jsonschema:"0-3"`
	RuleAvailability    int               `json:"availability" jsonschema:"0-3"`
	RuleCategory        string            `json:"category"`
	RuleTechnique       string            `json:"technique"`
	RuleDescription     string            `json:"description"`
	RuleReferences      []any             `json:"references,omitempty"`
	RuleDefinition      string            `json:"definition" jsonschema:"CEL expression"`
	RuleGroupBy         []string          `json:"groupBy,omitempty"`
	DeduplicateBy       []string          `json:"deduplicateBy,omitempty"`
	Correlation         any               `json:"correlation,omitempty"`
	RuleActive          bool              `json:"ruleActive"`
	DataTypes           []dto.DataTypeRef `json:"dataTypes"`
}

func mustRaw(v any) json.RawMessage {
	if v == nil || reflect.ValueOf(v).IsZero() {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

func (in epCorrelationRuleCreateInput) toDTO() dto.CreateCorrelationRuleRequest {
	return dto.CreateCorrelationRuleRequest{
		RuleName:            in.RuleName,
		RuleAdversary:       in.RuleAdversary,
		RuleConfidentiality: in.RuleConfidentiality,
		RuleIntegrity:       in.RuleIntegrity,
		RuleAvailability:    in.RuleAvailability,
		RuleCategory:        in.RuleCategory,
		RuleTechnique:       in.RuleTechnique,
		RuleDescription:     in.RuleDescription,
		RuleReferencesDef:   mustRaw(in.RuleReferences),
		RuleDefinitionDef:   mustRaw(in.RuleDefinition),
		RuleGroupByDef:      mustRaw(in.RuleGroupBy),
		DeduplicateByDef:    mustRaw(in.DeduplicateBy),
		CorrelationDef:      mustRaw(in.Correlation),
		RuleActive:          in.RuleActive,
		DataTypes:           in.DataTypes,
	}
}

func (in epCorrelationRuleUpdateInput) toDTO() dto.UpdateCorrelationRuleRequest {
	return dto.UpdateCorrelationRuleRequest{
		RelPath:             in.RelPath,
		RuleName:            in.RuleName,
		RuleAdversary:       in.RuleAdversary,
		RuleConfidentiality: in.RuleConfidentiality,
		RuleIntegrity:       in.RuleIntegrity,
		RuleAvailability:    in.RuleAvailability,
		RuleCategory:        in.RuleCategory,
		RuleTechnique:       in.RuleTechnique,
		RuleDescription:     in.RuleDescription,
		RuleReferencesDef:   mustRaw(in.RuleReferences),
		RuleDefinitionDef:   mustRaw(in.RuleDefinition),
		RuleGroupByDef:      mustRaw(in.RuleGroupBy),
		DeduplicateByDef:    mustRaw(in.DeduplicateBy),
		CorrelationDef:      mustRaw(in.Correlation),
		RuleActive:          in.RuleActive,
		DataTypes:           in.DataTypes,
	}
}

func registerEPCorrelationRules(m *Module) {
	uc := m.deps.EventProcessing.GetCorrelationRuleUsecase()

	Add(m, &mcp.Tool{
		Name: "correlation_rule.create", Title: "Create correlation rule",
	}, Gate{Permission: "eventprocessing.write"},
		func(ctx context.Context, _ *authz.Actor, in epCorrelationRuleCreateInput) (any, error) {
			req := in.toDTO()
			if err := uc.Create(ctx, req); err != nil {
				return nil, err
			}
			return map[string]any{"name": req.RuleName, "created": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "correlation_rule.update", Title: "Update correlation rule",
	}, Gate{Permission: "eventprocessing.write"},
		func(ctx context.Context, _ *authz.Actor, in epCorrelationRuleUpdateInput) (any, error) {
			req := in.toDTO()
			if err := uc.Update(ctx, req); err != nil {
				return nil, err
			}
			return map[string]any{"rel_path": req.RelPath, "updated": true}, nil
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
	GroupBy  string `json:"group_by,omitempty"`
	Status   string `json:"status,omitempty"`
	Interval string `json:"interval,omitempty"`
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	Top      int    `json:"top,omitempty"`
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
