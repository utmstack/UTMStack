package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	ospkg "github.com/utmstack/utmstack/backend/pkg/opensearch"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

// TODO(module-15): wire IndexPatternRegistry.Alerts here when all consumers are ported.
// alertIndex is the OpenSearch index pattern for alert documents.
// After Java Liquibase changeset 20241227001, 'alert-*' was prefixed to 'v11-alert-*'.
// TODO(module-15): wire IndexPatternRegistry.Alerts here when all consumers are ported.
const alertIndex = "v11-alert-*"

// falsePositiveTag is the canonical false-positive tag string.
const falsePositiveTag = "False positive"

// ---------------------------------------------------------------------------
// Painless script constants — all are static string literals; NO string concat.
// ---------------------------------------------------------------------------

const updateStatusScript = `
ctx._source.status = params.status;
ctx._source.statusLabel = params.statusLabel;
ctx._source.statusObservation = params.statusObservation;
`

const updateTagsScript = `ctx._source.tags = params.tags;`

const updateNotesScript = `ctx._source.notes = params.notes;`

// addFalsePositiveScript adds the "False positive" tag without disturbing others.
const addFalsePositiveScript = `
if (ctx._source.tags == null) { ctx._source.tags = []; }
if (!ctx._source.tags.contains(params.tag)) { ctx._source.tags.add(params.tag); }
`

const convertToIncidentScript = `
ctx._source.isIncident = true;
if (ctx._source.incidentDetail == null) { ctx._source.incidentDetail = [:]; }
ctx._source.incidentDetail.incidentName = params.incidentName;
ctx._source.incidentDetail.incidentId   = params.incidentId;
ctx._source.incidentDetail.creationDate = params.creationDate;
ctx._source.incidentDetail.createdBy    = params.createdBy;
ctx._source.incidentDetail.source       = params.source;
`

// applyTagRuleScript applies tags from a tag rule and optionally completes the alert.
// NOTE: Painless writes 'TagRulesApplied' (capital T) — BUG-003 ported verbatim.
const applyTagRuleScript = `
if (!ctx._source.containsKey('tags') || ctx._source.tags == null || ctx._source.tags.empty) {
  ctx._source.tags = new ArrayList();
}
ctx._source.tags.addAll(params.tagsForInsert);
ctx._source.tags = ctx._source.tags.stream().distinct().collect(Collectors.toList());

if (!ctx._source.containsKey('TagRulesApplied') || ctx._source.TagRulesApplied == null || ctx._source.TagRulesApplied.empty) {
  ctx._source.TagRulesApplied = new ArrayList();
}
ctx._source.TagRulesApplied.add(params.ruleId);
ctx._source.TagRulesApplied = ctx._source.TagRulesApplied.stream().distinct().collect(Collectors.toList());

if (ctx._source.tags.contains(params.falsePositiveTag)) {
  ctx._source.status = params.completedCode;
  ctx._source.statusLabel = params.completedLabel;
  ctx._source.statusObservation = params.completedObservation;
}
`

const releaseToOpenScript = `
ctx._source.status = params.openCode;
ctx._source.statusLabel = params.openLabel;
ctx._source.statusObservation = params.openObservation;
`

const assignAssetGroupsScript = `
if (params.mapping.containsKey(ctx._source.dataSource)) {
  ctx._source.assetGroupId = params.mapping[ctx._source.dataSource].id;
  ctx._source.assetGroupName = params.mapping[ctx._source.dataSource].name;
}
`

// ---------------------------------------------------------------------------
// osAlertRepo implements connectors.AlertRepository using pkg/opensearch.
// ---------------------------------------------------------------------------

type osAlertRepo struct {
	client *ospkg.Client
}

func NewOSAlertRepository(client *ospkg.Client) connectors.AlertRepository {
	return &osAlertRepo{client: client}
}

func idOrParentFilter(alertIDs []string) map[string]any {
	return map[string]any{
		"bool": map[string]any{
			"should": []map[string]any{
				ospkg.TermsQuery("id.keyword", alertIDs),
				ospkg.TermsQuery("parentId.keyword", alertIDs),
			},
			"minimum_should_match": 1,
		},
	}
}

func idOnlyFilter(alertIDs []string) map[string]any {
	return ospkg.TermsQuery("id.keyword", alertIDs)
}

func (r *osAlertRepo) UpdateStatus(ctx context.Context, alertIDs []string, status int, statusLabel, observation string) error {
	script := ospkg.Script{
		Source: updateStatusScript,
		Params: map[string]any{
			"status":            status,
			"statusLabel":       statusLabel,
			"statusObservation": observation,
		},
	}
	return r.client.UpdateByQuery(ctx, alertIndex, idOrParentFilter(alertIDs), script)
}

func (r *osAlertRepo) UpdateStatusAndTag(ctx context.Context, alertIDs []string) error {
	script := ospkg.Script{
		Source: addFalsePositiveScript,
		Params: map[string]any{
			"tag": falsePositiveTag,
		},
	}
	return r.client.UpdateByQuery(ctx, alertIndex, idOrParentFilter(alertIDs), script)
}

func (r *osAlertRepo) UpdateNotes(ctx context.Context, alertID, notes string) error {
	script := ospkg.Script{
		Source: updateNotesScript,
		Params: map[string]any{
			"notes": notes,
		},
	}
	filter := ospkg.TermQuery("id.keyword", alertID)
	return r.client.UpdateByQuery(ctx, alertIndex, filter, script)
}

func (r *osAlertRepo) UpdateTags(ctx context.Context, alertIDs []string, tags []string) error {
	var tagsParam any
	if len(tags) > 0 {
		tagsParam = tags
	}
	script := ospkg.Script{
		Source: updateTagsScript,
		Params: map[string]any{
			"tags": tagsParam,
		},
	}
	return r.client.UpdateByQuery(ctx, alertIndex, idOnlyFilter(alertIDs), script)
}

func (r *osAlertRepo) ConvertToIncident(ctx context.Context, alertIDs []string, name string, id int, createdAt time.Time, createdBy, source string) error {
	filter := map[string]any{
		"bool": map[string]any{
			"must": []map[string]any{
				idOnlyFilter(alertIDs),
				ospkg.TermQuery("isIncident", false),
			},
		},
	}
	script := ospkg.Script{
		Source: convertToIncidentScript,
		Params: map[string]any{
			"incidentName": name,
			"incidentId":   id,
			"creationDate": createdAt.UTC().Format(time.RFC3339),
			"createdBy":    createdBy,
			"source":       source,
		},
	}
	return r.client.UpdateByQuery(ctx, alertIndex, filter, script)
}

func (r *osAlertRepo) CountOpenAlerts(ctx context.Context) (int64, error) {
	query := map[string]any{
		"bool": map[string]any{
			"must": []map[string]any{
				ospkg.TermQuery("status", int(domain.AlertStatusOpen)),
			},
			"must_not": []map[string]any{
				{
					"term": map[string]any{
						"tags": falsePositiveTag,
					},
				},
				{
					"exists": map[string]any{
						"field": "parentId.keyword",
					},
				},
			},
		},
	}
	return r.client.Count(ctx, alertIndex, query)
}

func (r *osAlertRepo) CountByStatus(ctx context.Context, status int) (int64, error) {
	query := ospkg.TermQuery("status", status)
	return r.client.Count(ctx, alertIndex, query)
}

func (r *osAlertRepo) SearchByIDs(ctx context.Context, alertIDs []string) ([]domain.UtmAlert, error) {
	query := idOnlyFilter(alertIDs)
	raws, err := r.client.Search(ctx, alertIndex, query, 10000)
	if err != nil {
		return nil, err
	}
	alerts := make([]domain.UtmAlert, 0, len(raws))
	for _, raw := range raws {
		var a domain.UtmAlert
		if err := json.Unmarshal(raw, &a); err != nil {
			continue
		}
		alerts = append(alerts, a)
	}
	return alerts, nil
}

func filterTypeToQuery(f common_models.FilterType) map[string]any {
	switch f.Operator {
	case common_models.OpEquals:
		return ospkg.TermQuery(f.Field, f.Value)
	case common_models.OpNotEquals:
		return map[string]any{
			"bool": map[string]any{
				"must_not": ospkg.TermQuery(f.Field, f.Value),
			},
		}
	case common_models.OpIn, common_models.OpInOr:
		vals := toAnySlice(f.Value)
		return map[string]any{
			"terms": map[string]any{
				f.Field: vals,
			},
		}
	case common_models.OpNotExists:
		return map[string]any{
			"bool": map[string]any{
				"must_not": map[string]any{
					"exists": map[string]any{"field": f.Field},
				},
			},
		}
	case common_models.OpExists:
		return map[string]any{
			"exists": map[string]any{"field": f.Field},
		}
	case common_models.OpNotContains:
		return map[string]any{
			"bool": map[string]any{
				"must_not": ospkg.MatchQuery(f.Field, f.Value),
			},
		}
	default:
		return ospkg.MatchQuery(f.Field, f.Value)
	}
}

func toAnySlice(v any) []any {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case []any:
		return t
	case []string:
		out := make([]any, len(t))
		for i, s := range t {
			out[i] = s
		}
		return out
	default:
		return []any{v}
	}
}

func (r *osAlertRepo) ApplyTagRule(ctx context.Context, rule domain.UtmAlertTagRule, tagsForInsert []string, evaluationStart time.Time) error {
	// Start with the base must clauses: status + timestamp range.
	mustClauses := []map[string]any{
		ospkg.TermQuery("status", int(domain.AlertStatusAutomaticReview)),
		{
			"range": map[string]any{
				"@timestamp": map[string]any{
					"lte": evaluationStart.UTC().Format(time.RFC3339),
				},
			},
		},
	}

	// Parse rule conditions and append each as an additional must clause.
	if rule.RuleConditions != "" {
		var conditions []common_models.FilterType
		if err := json.Unmarshal([]byte(rule.RuleConditions), &conditions); err == nil {
			for _, cond := range conditions {
				mustClauses = append(mustClauses, filterTypeToQuery(cond))
			}
		}
	}

	filter := map[string]any{
		"bool": map[string]any{
			"must": mustClauses,
		},
	}

	script := ospkg.Script{
		Source: applyTagRuleScript,
		Params: map[string]any{
			"tagsForInsert":        tagsForInsert,
			"ruleId":               rule.ID,
			"falsePositiveTag":     falsePositiveTag,
			"completedCode":        int(domain.AlertStatusCompleted),
			"completedLabel":       domain.StatusName(domain.AlertStatusCompleted),
			"completedObservation": "Status changed to completed because alert was tagged as False positive",
		},
	}
	return r.client.UpdateByQuery(ctx, alertIndex, filter, script)
}

func (r *osAlertRepo) ReleaseToOpen(ctx context.Context, evaluationStart time.Time) error {
	filter := map[string]any{
		"bool": map[string]any{
			"must": []map[string]any{
				ospkg.TermQuery("status", int(domain.AlertStatusAutomaticReview)),
				{
					"range": map[string]any{
						"@timestamp": map[string]any{
							"lte": evaluationStart.UTC().Format(time.RFC3339),
						},
					},
				},
			},
		},
	}
	script := ospkg.Script{
		Source: releaseToOpenScript,
		Params: map[string]any{
			"openCode":        int(domain.AlertStatusOpen),
			"openLabel":       domain.StatusName(domain.AlertStatusOpen),
			"openObservation": "This alert has been evaluated by the tag rules engine",
		},
	}
	return r.client.UpdateByQuery(ctx, alertIndex, filter, script)
}

func (r *osAlertRepo) SearchByStatusAndTimestamp(ctx context.Context, status int, evaluationStart time.Time, size int) ([]domain.UtmAlert, error) {
	query := map[string]any{
		"bool": map[string]any{
			"must": []map[string]any{
				ospkg.TermQuery("status", status),
				{
					"range": map[string]any{
						"@timestamp": map[string]any{
							"lte": evaluationStart.UTC().Format(time.RFC3339),
						},
					},
				},
			},
		},
	}
	raws, err := r.client.Search(ctx, alertIndex, query, size)
	if err != nil {
		return nil, err
	}
	alerts := make([]domain.UtmAlert, 0, len(raws))
	for _, raw := range raws {
		var a domain.UtmAlert
		if err := json.Unmarshal(raw, &a); err != nil {
			continue
		}
		alerts = append(alerts, a)
	}
	return alerts, nil
}

func (r *osAlertRepo) SearchByRuleConditionsAndTimestamp(ctx context.Context, rule domain.UtmAlertTagRule, evaluationStart time.Time, size int) ([]domain.UtmAlert, error) {
	mustClauses := []map[string]any{
		ospkg.TermQuery("status", int(domain.AlertStatusAutomaticReview)),
		{
			"range": map[string]any{
				"@timestamp": map[string]any{
					"lte": evaluationStart.UTC().Format(time.RFC3339),
				},
			},
		},
	}
	// Apply rule conditions — same logic as ApplyTagRule.
	if rule.RuleConditions != "" {
		var conditions []common_models.FilterType
		if err := json.Unmarshal([]byte(rule.RuleConditions), &conditions); err == nil {
			for _, cond := range conditions {
				mustClauses = append(mustClauses, filterTypeToQuery(cond))
			}
		}
	}
	query := map[string]any{"bool": map[string]any{"must": mustClauses}}
	raws, err := r.client.Search(ctx, alertIndex, query, size)
	if err != nil {
		return nil, err
	}
	alerts := make([]domain.UtmAlert, 0, len(raws))
	for _, raw := range raws {
		var a domain.UtmAlert
		if err := json.Unmarshal(raw, &a); err != nil {
			continue
		}
		alerts = append(alerts, a)
	}
	return alerts, nil
}

func (r *osAlertRepo) AssignAssetGroups(ctx context.Context, mapping map[string]connectors.AssetGroupRef) error {
	filter := ospkg.TermQuery("status", int(domain.AlertStatusAutomaticReview))

	// Convert mapping to a plain map[string]any for Painless params.
	pmapping := make(map[string]any, len(mapping))
	for k, v := range mapping {
		pmapping[k] = map[string]any{"id": v.ID, "name": v.Name}
	}

	script := ospkg.Script{
		Source: assignAssetGroupsScript,
		Params: map[string]any{
			"mapping": pmapping,
		},
	}
	return r.client.UpdateByQuery(ctx, alertIndex, filter, script)
}
