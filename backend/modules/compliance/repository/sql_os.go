package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	sdkos "github.com/threatwinds/go-sdk/os"

	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

type osSQL struct{}

// NewOpenSearchSQL returns a runner that executes compliance check SQL against
// the OpenSearch SQL plugin (_plugins/_sql) using the SDK's configured client.
//
// ponytail: the OS SQL plugin is the only piece of compliance still on
// OpenSearch — CSV export and this runner are the two things `06cad527f`
// couldn't retire yet. Migrate check specs off free-form SQL onto a structured
// filter+rule shape and this file dies with alerts_os.go.
func NewOpenSearchSQL() connectors.OpenSearchSQL { return &osSQL{} }

func (r *osSQL) RunCheck(ctx context.Context, sql string) (int64, error) {
	scoped, err := scopeSQLToTenant(sql, authz.TenantIDFromContext(ctx))
	if err != nil {
		return 0, err
	}
	data, status, err := sdkos.DoRequest(ctx, "POST", "/_plugins/_sql", map[string]string{"query": scoped})
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

// scopeSQLToTenant appends a tenant predicate to a check's WHERE. Check SQL is
// trusted YAML input, not user input, so an unrecognised shape returns the
// original string unchanged. Empty tenant (on-prem/global) short-circuits.
//
// Handles the two shapes v11 checks emit: `... WHERE (<cond>)` and `... FROM
// <idx>` with no WHERE. GROUP BY / ORDER BY / LIMIT / HAVING clauses that
// follow are kept intact — the predicate is inserted just before them.
func scopeSQLToTenant(sql, tenantID string) (string, error) {
	if tenantID == "" {
		return sql, nil
	}
	pred := fmt.Sprintf(`tenantId = '%s'`, tenantID)

	// Insert into an existing WHERE.
	if loc := whereRe.FindStringIndex(sql); loc != nil {
		endOfWhere := len(sql)
		if tail := tailRe.FindStringIndex(sql); tail != nil && tail[0] > loc[1] {
			endOfWhere = tail[0]
		}
		before := strings.TrimRight(sql[:endOfWhere], " ")
		after := sql[endOfWhere:]
		sep := ""
		if after != "" && !strings.HasPrefix(after, " ") {
			sep = " "
		}
		return before + " AND " + pred + sep + after, nil
	}

	// No WHERE — inject one before the trailing clause, if any.
	if tail := tailRe.FindStringIndex(sql); tail != nil {
		before := strings.TrimRight(sql[:tail[0]], " ")
		return before + " WHERE " + pred + " " + sql[tail[0]:], nil
	}
	return strings.TrimRight(sql, " ") + " WHERE " + pred, nil
}

var (
	whereRe = regexp.MustCompile(`(?i)\bWHERE\b`)
	tailRe  = regexp.MustCompile(`(?i)\b(GROUP\s+BY|ORDER\s+BY|HAVING|LIMIT)\b`)
)
