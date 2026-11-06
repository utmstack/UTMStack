package usecase

import "testing"

// CFG-F27: a forbidden schema-qualified reference must be caught no matter
// how it's quoted or spaced, since scopedSQL's tenant isolation (sql_ch.go)
// only applies to the bare, unqualified "logs"/"alerts" names — any query
// that reaches the real table directly reads every tenant's data.
func TestValidateSQLCatchesQuotedAndSpacedQualifiedNames(t *testing.T) {
	cases := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		{"bare unqualified logs is fine", "SELECT * FROM logs", false},
		{"unquoted qualified name is blocked", "SELECT * FROM utmstack.logs", true},
		{"backtick-quoted qualified name is blocked", "SELECT * FROM `utmstack`.`logs`", true},
		{"double-quoted qualified name is blocked", `SELECT * FROM "utmstack"."logs"`, true},
		{"space before the dot is blocked", "SELECT * FROM `utmstack` .`logs`", true},
		{"space after the dot is blocked", "SELECT * FROM `utmstack`. `logs`", true},
		{"space on both sides of the dot is blocked", "SELECT * FROM `utmstack` . `logs`", true},
		{"mixed quoting with spacing is blocked", `SELECT * FROM "utmstack" . logs`, true},
		{"system schema with spacing is blocked", "SELECT * FROM `system` . query_log", true},
		{"information_schema is blocked regardless", "SELECT * FROM information_schema.tables", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSQL(tc.sql)
			if tc.wantErr && err == nil {
				t.Fatalf("ValidateSQL(%q) = nil, want an error", tc.sql)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("ValidateSQL(%q) = %v, want nil", tc.sql, err)
			}
		})
	}
}

func TestValidateSQLBasicGuards(t *testing.T) {
	if err := ValidateSQL(""); err == nil {
		t.Fatal("empty query should be rejected")
	}
	if err := ValidateSQL("DROP TABLE logs"); err == nil {
		t.Fatal("a query not starting with SELECT/WITH should be rejected")
	}
	if err := ValidateSQL("SELECT * FROM logs -- drop everything"); err == nil {
		t.Fatal("a line comment should be rejected")
	}
	if err := ValidateSQL("SELECT * FROM logs; SELECT * FROM alerts"); err == nil {
		t.Fatal("more than one statement should be rejected")
	}
	if err := ValidateSQL("SELECT * FROM logs WHERE message = 'normal value'"); err != nil {
		t.Fatalf("an ordinary query should pass, got %v", err)
	}
}
