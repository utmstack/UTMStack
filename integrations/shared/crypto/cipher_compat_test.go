package crypto

import "testing"

// A value the backend encrypted (backend/pkg/secret) with the key "compat-key". The containers read
// what the backend writes, so the two ciphers have to stay compatible.
func TestItReadsWhatTheBackendEncrypts(t *testing.T) {
	const fromBackend = "2/hPKzsmwtwtbIXPJiB5qMPxpqXVYKL8/BWd7iiLhJ89cL622FRRL9CZvVQQ8JXA"

	got, err := NewCipher("compat-key").Decrypt(fromBackend)
	if err != nil {
		t.Fatal(err)
	}

	if want := "plugins:\n  aws:\n    tenants: []\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
