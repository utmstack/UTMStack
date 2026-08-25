package auth

import "testing"

// The stored value is "tenant\x00key". What matters is that a presented key
// resolves to the tenant it was issued for and to nothing else.
func TestMatchConnector(t *testing.T) {
	cases := []struct {
		name       string
		stored     string
		presented  string
		wantTenant string
		wantOK     bool
	}{
		{"key of its tenant", "tenant-a\x00secret", "secret", "tenant-a", true},
		{"wrong key", "tenant-a\x00secret", "guess", "", false},
		{"empty key against a real one", "tenant-a\x00secret", "", "", false},
		{"the tenant is not a key", "tenant-a\x00secret", "tenant-a", "", false},

		// An agent-manager that has not been upgraded yet writes the key alone.
		// It authenticates, into no tenant in particular.
		{"value from an older writer", "secret", "secret", "", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tenant, ok := matchConnector(c.stored, c.presented)
			if ok != c.wantOK {
				t.Fatalf("ok = %v, want %v", ok, c.wantOK)
			}
			if tenant != c.wantTenant {
				t.Errorf("tenant = %q, want %q", tenant, c.wantTenant)
			}
		})
	}
}

// getKey returns the tenant and the key apart; the cache holds them joined, and
// matchConnector splits them again. Nothing in the type system keeps those
// three in step, so the round trip is asserted here: a separator changed on one
// side alone would otherwise authenticate into the wrong tenant, or into none.
func TestConnectorRoundTripsThroughTheCacheFormat(t *testing.T) {
	cases := []struct {
		name string
		auth ConnectorAuth
	}{
		{"an ordinary connector", ConnectorAuth{Key: "secret", TenantID: "tenant-a"}},
		{"the default tenant is empty", ConnectorAuth{Key: "secret", TenantID: ""}},
		{"a key that looks like a separator", ConnectorAuth{Key: "a\x00b", TenantID: "tenant-a"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stored := c.auth.TenantID + "\x00" + c.auth.Key

			tenant, ok := matchConnector(stored, c.auth.Key)
			if !ok {
				t.Fatalf("the key it was issued did not authenticate")
			}
			if tenant != c.auth.TenantID {
				t.Fatalf("tenant = %q, want %q", tenant, c.auth.TenantID)
			}

			if _, ok := matchConnector(stored, c.auth.Key+"x"); ok {
				t.Fatalf("a key that is not the issued one authenticated")
			}
		})
	}
}
