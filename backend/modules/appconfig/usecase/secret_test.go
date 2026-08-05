package usecase

import (
	"testing"

	"github.com/utmstack/utmstack/backend/modules/appconfig/domain"
	"github.com/utmstack/utmstack/backend/pkg/secret"
)

func newCipher(t *testing.T) *secret.Cipher {
	t.Helper()
	c, err := secret.NewCipher("test-encryption-key-0123456789")
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}
	return c
}

// The SMTP password is readable in process, where the mailer needs it, and never
// over the API — config.read is granted to every ordinary user, so returning it
// handed the instance's mail credential to anyone with an account.
func TestSecretsAreNeverReturned(t *testing.T) {
	cipher := newCipher(t)
	s := &service{cipher: cipher}

	const plain = "the-smtp-password"
	enc, err := cipher.Encrypt(plain)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	row := domain.Config{
		ConfParamShort:    "utmstack.mail.password",
		ConfParamValue:    enc,
		ConfParamDatatype: "password",
	}

	got := s.toResponse(row)

	if got.Value != "" {
		t.Errorf("value = %q, want it withheld", got.Value)
	}
	if got.Value == plain || got.Value == enc {
		t.Error("the stored secret was returned")
	}
	if !got.IsSecret {
		t.Error("isSecret = false, want the caller told it is one")
	}
	if !got.IsSet {
		t.Error("isSet = false; the UI needs to know a value is stored")
	}
}

// Everything that is not a secret still comes back: the UI renders timestamps
// from the date format and language.
func TestOrdinaryValuesAreStillReturned(t *testing.T) {
	s := &service{cipher: newCipher(t)}

	got := s.toResponse(domain.Config{
		ConfParamShort: "utmstack.time.dateformat",
		ConfParamValue: "yyyy-MM-dd",
	})

	if got.Value != "yyyy-MM-dd" {
		t.Errorf("value = %q, want it returned", got.Value)
	}
}
