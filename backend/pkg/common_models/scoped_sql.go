package common_models

import (
	"fmt"
	"strings"
)

// ScopeSQL wraps a caller-supplied SELECT with tenant-scoped logs/alerts CTEs.
// Both the log explorer's SQL mode and the store.search_sql MCP tool run
// queries this way, so the scoping lives in one place.
//
// FetchQuery binds no parameters, so the tenant is inlined; sqlQuote escapes
// the two characters that can end a ClickHouse string literal. Callers must
// first run the text through the SQL guard (loganalyzer usecase ValidateSQL).
//
// The CTEs only shadow the bare names `logs` and `alerts`. A query that names
// the physical table, a system table or a table function would walk around
// them — which is why the guard runs first.
func ScopeSQL(query, tenant, logsTable, alertsTable string, page, size int) string {
	if page < 1 {
		page = 1
	}
	q := strings.TrimSpace(strings.TrimRight(strings.TrimSpace(query), ";"))
	return fmt.Sprintf(
		"WITH logs AS (SELECT * FROM %s WHERE tenantId = %s), alerts AS (SELECT * FROM %s WHERE tenantId = %s) "+
			"SELECT * FROM (%s) LIMIT %d OFFSET %d",
		logsTable, sqlQuote(tenant), alertsTable, sqlQuote(tenant), q, size, (page-1)*size,
	)
}

// sqlQuote renders a ClickHouse string literal: backslash and quote are the
// two characters that can end a literal early.
func sqlQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	return "'" + s + "'"
}
