package main

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	sdkos "github.com/threatwinds/go-sdk/os"
)

const (
	alertIndexPattern = "v11-alert-*"
	falsePositiveTag  = "False positive"

	statusAutomaticReview = 1
	statusOpen            = 2
	statusCompleted       = 5

	tagRuleOpenObservation      = "This alert has been evaluated by the tag rules engine"
	tagRuleCompletedObservation = "Status changed to completed because alert was tagged as False positive"
)

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

// applyTagRulesAndRelease evaluates each rule's conditions against the freshly
// indexed alert via id-scoped update_by_query calls, then flips any remaining
func applyTagRulesAndRelease(ctx context.Context, alertID string, rules []RuleSnapshot) {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in applyTagRulesAndRelease", nil, map[string]any{
				"panic":   r,
				"alertId": alertID,
				"process": "plugin_com.utmstack.alerts",
			})
		}
	}()

	if alertID == "" {
		return
	}

	refreshCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	if err := sdkos.RefreshIndex(refreshCtx, alertIndexPattern); err != nil {
		_ = catcher.Error("tag rules: refresh index failed", err, map[string]any{
			"alertId": alertID,
			"process": "plugin_com.utmstack.alerts",
		})
	}
	cancel()

	for _, rule := range rules {
		applyRule(ctx, alertID, rule)
	}

	releaseAlertToOpen(ctx, alertID)
}

func applyRule(ctx context.Context, alertID string, rule RuleSnapshot) {
	mustClauses := []map[string]any{
		{"term": map[string]any{"id.keyword": alertID}},
		{"term": map[string]any{"status": statusAutomaticReview}},
	}
	for _, cond := range rule.Conditions {
		mustClauses = append(mustClauses, cond.toQueryClause())
	}

	body := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must": mustClauses,
			},
		},
		"script": map[string]any{
			"source": applyTagRuleScript,
			"lang":   "painless",
			"params": map[string]any{
				"tagsForInsert":        rule.TagNames,
				"ruleId":               rule.ID,
				"falsePositiveTag":     falsePositiveTag,
				"completedCode":        statusCompleted,
				"completedLabel":       "Completed",
				"completedObservation": tagRuleCompletedObservation,
			},
		},
	}

	callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if _, err := sdkos.UpdateByQuery(callCtx, []string{alertIndexPattern}, body); err != nil {
		_ = catcher.Error("tag rules: apply rule failed", err, map[string]any{
			"alertId": alertID,
			"rule":    rule.Name,
			"process": "plugin_com.utmstack.alerts",
		})
	}
}

func releaseAlertToOpen(ctx context.Context, alertID string) {
	body := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must": []map[string]any{
					{"term": map[string]any{"id.keyword": alertID}},
					{"term": map[string]any{"status": statusAutomaticReview}},
				},
			},
		},
		"script": map[string]any{
			"source": releaseToOpenScript,
			"lang":   "painless",
			"params": map[string]any{
				"openCode":        statusOpen,
				"openLabel":       "Open",
				"openObservation": tagRuleOpenObservation,
			},
		},
	}

	callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if _, err := sdkos.UpdateByQuery(callCtx, []string{alertIndexPattern}, body); err != nil {
		_ = catcher.Error("tag rules: release to open failed", err, map[string]any{
			"alertId": alertID,
			"process": "plugin_com.utmstack.alerts",
		})
	}
}
