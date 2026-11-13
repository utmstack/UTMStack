package mcp

import (
	"strings"
	"testing"

	loganalyzer_usecase "github.com/utmstack/utmstack/backend/modules/loganalyzer/usecase"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

const (
	testLogs   = "utmstack.logs"
	testAlerts = "utmstack.alerts"
	testTenant = "8f1c1b8e-0000-4000-8000-00000000000a"
)

func TestStoreSQLScopesTheBareDatasetNames(t *testing.T) {
	got := common_models.ScopeSQL("SELECT dataSource, count() FROM logs GROUP BY dataSource;", testTenant, testLogs, testAlerts, 1, 50)
	if !strings.Contains(got, "WITH logs AS (SELECT * FROM utmstack.logs WHERE tenantId = '") {
		t.Fatalf("query was not wrapped in the tenant-scoped CTEs: %s", got)
	}
	if !strings.Contains(got, testTenant) {
		t.Fatalf("the session tenant is missing from the CTEs: %s", got)
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
		if err := loganalyzer_usecase.ValidateSQL(query); err == nil {
			t.Errorf("%s: %q should have been refused", name, query)
		}
	}
}
