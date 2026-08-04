package repository

import (
	"testing"

	"github.com/utmstack/utmstack/backend/pkg/secret"
)

func newTestCipher(t *testing.T) *secret.Cipher {
	t.Helper()
	c, err := secret.NewCipher("test-encryption-key-0123456789")
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}
	return c
}

func TestTfaSecretRoundTrip(t *testing.T) {
	c := newTestCipher(t)

	// base32, which is also valid base64 — the case that makes "try to decrypt
	// and see whether it worked" unusable.
	const plain = "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"

	stored, err := encryptTfaSecret(c, plain)
	if err != nil {
		t.Fatalf("encryptTfaSecret: %v", err)
	}
	if stored == plain {
		t.Fatal("secret was stored unchanged")
	}

	got, err := decryptTfaSecret(c, stored)
	if err != nil {
		t.Fatalf("decryptTfaSecret: %v", err)
	}
	if got != plain {
		t.Fatalf("round trip = %q, want %q", got, plain)
	}
}

// A secret written before encryption was introduced has to keep working, or
// every user with 2FA already enabled is locked out on deploy.
func TestTfaSecretPassesLegacyPlaintextThrough(t *testing.T) {
	c := newTestCipher(t)

	for _, legacy := range []string{
		"JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP", // TOTP secret
		"123456",                           // emailed login code
		"not base64 at all !!",
	} {
		got, err := decryptTfaSecret(c, legacy)
		if err != nil {
			t.Fatalf("decryptTfaSecret(%q): %v", legacy, err)
		}
		if got != legacy {
			t.Fatalf("decryptTfaSecret(%q) = %q, want it unchanged", legacy, got)
		}
	}
}

func TestTfaSecretHandlesEmpty(t *testing.T) {
	c := newTestCipher(t)

	stored, err := encryptTfaSecret(c, "")
	if err != nil {
		t.Fatalf("encryptTfaSecret(\"\"): %v", err)
	}
	if stored != "" {
		t.Fatalf("encryptTfaSecret(\"\") = %q, want empty so a cleared secret stays cleared", stored)
	}
}
