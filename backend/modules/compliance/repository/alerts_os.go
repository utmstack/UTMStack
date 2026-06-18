package repository

import (
	"context"
	"encoding/json"
	"fmt"

	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
)

const alertIndex = "v11-alert-*"

type osAlerts struct{}

// NewOpenSearchAlerts returns a counter over the alert index (activity dimension).
func NewOpenSearchAlerts() connectors.OpenSearchAlerts { return &osAlerts{} }

func (r *osAlerts) CountByRuleNames(ctx context.Context, ruleNames []string, sinceISO string) (int64, error) {
	if len(ruleNames) == 0 {
		return 0, nil
	}
	terms := make([]any, len(ruleNames))
	for i, n := range ruleNames {
		terms[i] = n
	}
	body := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []any{
					map[string]any{"terms": map[string]any{"name.keyword": terms}},
					map[string]any{"range": map[string]any{"@timestamp": map[string]any{"gte": sinceISO}}},
				},
			},
		},
	}
	data, status, err := sdkos.DoRequest(ctx, "POST", "/"+alertIndex+"/_count", body)
	if err != nil {
		return 0, err
	}
	if status < 200 || status >= 300 {
		return 0, fmt.Errorf("alert count returned %d", status)
	}
	var resp struct {
		Count int64 `json:"count"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, err
	}
	return resp.Count, nil
}
