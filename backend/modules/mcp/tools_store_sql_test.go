package mcp

import (
	"strings"
	"testing"
)

const (
	testLogs   = "utmstack.logs"
	testAlerts = "utmstack.alerts"
)

func TestStoreSQLScopesTheBareDatasetNames(t *testing.T) {
	got, err := storeScopedSQL("SELECT dataSource, count() FROM logs GROUP BY dataSource;", testLogs, testAlerts, 1, 50)
	if err != nil {
		t.Fatalf("a plain query over the scoped datasets must pass: %v", err)
	}
	if !strings.Contains(got, "WITH logs AS (SELECT * FROM utmstack.logs WHERE tenantId = ?)") {
		t.Fatalf("query was not wrapped in the tenant-scoped CTEs: %s", got)
	}
}

// Anything that names a table directly reads around the tenant scoping.
func TestStoreSQLRefusesWhatWalksAroundTheTenantScope(t *testing.T) {
	for name, query := range map[string]string{
		"physical table":   "SELECT * FROM utmstack.logs",
		"system table":     "SELECT * FROM system.users",
		"table function":   "SELECT * FROM url('http://example.com/x', CSV)",
		"second statement": "SELECT 1; DROP TABLE utmstack.logs",
		"comment":          "SELECT * FROM logs -- and more",
		"not a select":     "DROP TABLE logs",
	} {
		if _, err := storeScopedSQL(query, testLogs, testAlerts, 1, 50); err == nil {
			t.Errorf("%s: %q should have been refused", name, query)
		}
	}
}
