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
