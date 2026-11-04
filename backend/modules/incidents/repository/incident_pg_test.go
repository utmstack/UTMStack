package repository

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsNameTaken(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"unrelated", errors.New("boom"), false},
		{"name constraint", &pgconn.PgError{Code: "23505", ConstraintName: "uniq_incident_tenant_name"}, true},
		{"wrapped name constraint", fmt.Errorf("create: %w", &pgconn.PgError{Code: "23505", ConstraintName: "uniq_incident_tenant_name"}), true},
		{"other unique constraint", &pgconn.PgError{Code: "23505", ConstraintName: "incidents_pkey"}, false},
		{"other pg error", &pgconn.PgError{Code: "23503", ConstraintName: "uniq_incident_tenant_name"}, false},
	}
	for _, c := range cases {
		if got := isNameTaken(c.err); got != c.want {
			t.Errorf("%s: isNameTaken = %v, want %v", c.name, got, c.want)
		}
	}
}
