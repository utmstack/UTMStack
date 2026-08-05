package repository

import "testing"

// What comes back from here goes into ORDER BY, which GORM treats as raw SQL. A
// field it does not recognise must never reach the query: in Postgres an ORDER
// BY can hold a subquery, and a subquery reads whatever it names regardless of
// the tenant the rest of the statement is bound to.
func TestSortIsAnAllowlist(t *testing.T) {
	injections := []string{
		"(CASE WHEN (SELECT password_hash FROM jhi_user LIMIT 1) LIKE 'a%' THEN 1 ELSE 2 END)",
		"(SELECT api_key FROM api_keys LIMIT 1)",
		"id; DROP TABLE utm_notification",
		"created_at DESC, (SELECT 1)",
		"tenant_id", // a real column, but not one to order strangers' data by
		"",
	}

	for _, in := range injections {
		if got := parseSortParam(in); got != defaultOrder {
			t.Errorf("parseSortParam(%q) = %q, want the default %q", in, got, defaultOrder)
		}
	}
}

func TestSortAcceptsTheColumnsTheUISends(t *testing.T) {
	cases := map[string]string{
		"created_at,desc": "created_at DESC",
		"created_at,DESC": "created_at DESC",
		"created_at":      "created_at ASC",
		"source,asc":      "source ASC",
		" status , desc ": "status DESC",
		"id,sideways":     "id ASC",
	}

	for in, want := range cases {
		if got := parseSortParam(in); got != want {
			t.Errorf("parseSortParam(%q) = %q, want %q", in, got, want)
		}
	}
}
