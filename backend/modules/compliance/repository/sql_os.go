package repository

import (
	"context"
	"encoding/json"
	"fmt"

	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
)

type osSQL struct{}

// NewOpenSearchSQL returns a runner that executes compliance check SQL against
// the OpenSearch SQL plugin (_plugins/_sql) using the SDK's configured client.
func NewOpenSearchSQL() connectors.OpenSearchSQL { return &osSQL{} }

func (r *osSQL) RunCheck(ctx context.Context, sql string) (int64, error) {
	data, status, err := sdkos.DoRequest(ctx, "POST", "/_plugins/_sql", map[string]string{"query": sql})
	if err != nil {
		return 0, err
	}
	if status < 200 || status >= 300 {
		snippet := string(data)
		if len(snippet) > 300 {
			snippet = snippet[:300]
		}
		return 0, fmt.Errorf("opensearch sql returned %d: %s", status, snippet)
	}

	var resp struct {
		Datarows [][]any `json:"datarows"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, err
	}
	// A single numeric cell (SELECT count(*)) is the count itself; otherwise the
	// number of rows is the hit count (e.g. GROUP BY ... HAVING listing offenders).
	if len(resp.Datarows) == 1 && len(resp.Datarows[0]) == 1 {
		if n, ok := toInt64(resp.Datarows[0][0]); ok {
			return n, nil
		}
	}
	return int64(len(resp.Datarows)), nil
}

func toInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	default:
		return 0, false
	}
}
