package repository

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/utmstack/utmstack/backend/pkg/authz"
)

func hasTenantFilter(t *testing.T, q map[string]any, want string) bool {
	t.Helper()
	b, err := json.Marshal(q)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var probe struct {
		Bool struct {
			Filter []map[string]map[string]any `json:"filter"`
		} `json:"bool"`
		Term map[string]any `json:"term"`
	}
	if err := json.Unmarshal(b, &probe); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if v, ok := probe.Term["tenantId"]; ok && v == want {
		return true
	}
	for _, f := range probe.Bool.Filter {
		if v, ok := f["term"]["tenantId"]; ok && v == want {
			return true
		}
	}
	return false
}

// Every read in this module goes through scopeTenantQuery, so this is the one
// place the isolation is decided.
func TestScopeTenantQueryAddsTheActingTenant(t *testing.T) {
	ctx := authz.WithTenantID(context.Background(), "tenant-a")

	t.Run("wraps an existing query", func(t *testing.T) {
		got := scopeTenantQuery(ctx, map[string]any{"match_all": map[string]any{}})
		if !hasTenantFilter(t, got, "tenant-a") {
			b, _ := json.Marshal(got)
			t.Fatalf("no tenant filter in %s", b)
		}
	})

	t.Run("stands alone when there is no query", func(t *testing.T) {
		got := scopeTenantQuery(ctx, nil)
		if !hasTenantFilter(t, got, "tenant-a") {
			b, _ := json.Marshal(got)
			t.Fatalf("no tenant filter in %s", b)
		}
	})

	// A caller-supplied filter must not be able to displace the tenant one: it
	// goes in must, the tenant goes in filter, and both have to hold.
	t.Run("keeps the caller's own query", func(t *testing.T) {
		got := scopeTenantQuery(ctx, map[string]any{"term": map[string]any{"severity": "high"}})
		b, _ := json.Marshal(got)
		if !hasTenantFilter(t, got, "tenant-a") {
			t.Fatalf("no tenant filter in %s", b)
		}
		if !containsSeverity(string(b)) {
			t.Fatalf("the caller's query was dropped: %s", b)
		}
	})
}

// On-prem has no tenant and must keep working unfiltered.
func TestScopeTenantQueryIsInertWithoutATenant(t *testing.T) {
	in := map[string]any{"match_all": map[string]any{}}
	got := scopeTenantQuery(context.Background(), in)

	b, _ := json.Marshal(got)
	if string(b) != `{"match_all":{}}` {
		t.Fatalf("query = %s, want it untouched", b)
	}
}

func containsSeverity(s string) bool {
	return len(s) > 0 && (indexOf(s, `"severity"`) >= 0)
}

func indexOf(h, n string) int {
	for i := 0; i+len(n) <= len(h); i++ {
		if h[i:i+len(n)] == n {
			return i
		}
	}
	return -1
}
