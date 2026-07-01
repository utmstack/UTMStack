package usecase

import (
	"errors"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
)

func TestSanitizeVisualizationSQL(t *testing.T) {
	cases := []struct {
		name    string
		sql     string
		wantErr error // sentinel expected via errors.Is (nil means pass)
	}{
		{
			name: "valid templated SELECT",
			sql: `SELECT bucket, count(*) AS y FROM logs
WHERE {{dashboardFilters}}{{timeFilter}}
GROUP BY bucket
ORDER BY y DESC
LIMIT 100`,
		},
		{
			name:    "empty is rejected as required",
			sql:     "   ",
			wantErr: domain.ErrSQLQueryRequired,
		},
		{
			name:    "non-SELECT rejected",
			sql:     "UPDATE logs SET user='x' WHERE 1=1",
			wantErr: domain.ErrInvalidSQL,
		},
		{
			name:    "forbidden keyword rejected",
			sql:     "SELECT * FROM logs; DROP TABLE users",
			wantErr: domain.ErrInvalidSQL,
		},
		{
			name:    "line comment rejected",
			sql:     "SELECT * FROM logs -- inject",
			wantErr: domain.ErrInvalidSQL,
		},
		{
			name:    "block comment rejected",
			sql:     "SELECT /* trick */ * FROM logs",
			wantErr: domain.ErrInvalidSQL,
		},
		{
			name:    "disallowed function rejected",
			sql:     "SELECT LOAD_FILE('/etc/passwd') FROM logs",
			wantErr: domain.ErrInvalidSQL,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := &domain.Visualization{SQLQuery: tc.sql}
			err := sanitizeVisualizationSQL(v)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected errors.Is(%v), got %v", tc.wantErr, err)
			}
		})
	}
}
