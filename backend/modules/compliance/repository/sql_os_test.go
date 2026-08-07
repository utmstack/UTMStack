package repository

import "testing"

// TestScopeSQLToTenant checks that the tenant predicate is appended without
// disturbing GROUP BY / ORDER BY / LIMIT tails, and that an empty tenant
// short-circuits to the input (on-prem/global behaviour).
func TestScopeSQLToTenant(t *testing.T) {
	const tid = "abc-123"
	cases := []struct{ name, in, want string }{
		{
			name: "existing WHERE",
			in:   `SELECT count(*) AS c FROM v11-log-wineventlog-* WHERE (event_id = '4729')`,
			want: `SELECT count(*) AS c FROM v11-log-wineventlog-* WHERE (event_id = '4729') AND tenantId = 'abc-123'`,
		},
		{
			name: "no WHERE, no tail",
			in:   `SELECT count(*) FROM v11-alert-*`,
			want: `SELECT count(*) FROM v11-alert-* WHERE tenantId = 'abc-123'`,
		},
		{
			name: "GROUP BY tail",
			in:   `SELECT host, count(*) FROM v11-log-* WHERE (severity = 5) GROUP BY host`,
			want: `SELECT host, count(*) FROM v11-log-* WHERE (severity = 5) AND tenantId = 'abc-123' GROUP BY host`,
		},
		{
			name: "no WHERE, LIMIT tail",
			in:   `SELECT * FROM v11-log-* LIMIT 10`,
			want: `SELECT * FROM v11-log-* WHERE tenantId = 'abc-123' LIMIT 10`,
		},
	}
	for _, c := range cases {
		got, err := scopeSQLToTenant(c.in, tid)
		if err != nil {
			t.Fatalf("%s: unexpected err: %v", c.name, err)
		}
		if got != c.want {
			t.Errorf("%s:\n got:  %s\n want: %s", c.name, got, c.want)
		}
	}

	// Empty tenant returns the input unchanged.
	unchanged := `SELECT count(*) FROM x WHERE y = 1`
	if got, _ := scopeSQLToTenant(unchanged, ""); got != unchanged {
		t.Errorf("empty tenant should not rewrite; got %q", got)
	}
}
