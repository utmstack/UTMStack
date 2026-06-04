package dto

import (
	"testing"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
)

// On the wire, sensitive values must come out masked. Non-sensitive values
// pass through untouched.
func TestFromConfiguration_MasksSensitive(t *testing.T) {
	row := domain.UtmModuleGroupConfiguration{
		ConfKey:      "token",
		ConfValue:    "encrypted-blob",
		ConfDataType: domain.ConfTypePassword,
	}
	out := FromConfiguration(row, false)
	if out.ConfValue != domain.MaskedValue {
		t.Fatalf("expected masked value, got %q", out.ConfValue)
	}
}

func TestFromConfiguration_PassesThroughNonSensitive(t *testing.T) {
	row := domain.UtmModuleGroupConfiguration{
		ConfKey:      "host",
		ConfValue:    "example.com",
		ConfDataType: domain.ConfTypeText,
	}
	out := FromConfiguration(row, false)
	if out.ConfValue != "example.com" {
		t.Fatalf("expected plaintext, got %q", out.ConfValue)
	}
}

// reveal=true is the internal-key-gated path; it returns the stored value
// verbatim so event-processor can decrypt it.
func TestFromConfiguration_RevealKeepsSensitive(t *testing.T) {
	row := domain.UtmModuleGroupConfiguration{
		ConfKey:      "token",
		ConfValue:    "encrypted-blob",
		ConfDataType: domain.ConfTypePassword,
	}
	out := FromConfiguration(row, true)
	if out.ConfValue != "encrypted-blob" {
		t.Fatalf("expected raw value, got %q", out.ConfValue)
	}
}

// Empty sensitive values stay empty (not masked) — there's nothing to hide.
func TestFromConfiguration_EmptySensitiveStaysEmpty(t *testing.T) {
	row := domain.UtmModuleGroupConfiguration{
		ConfKey:      "token",
		ConfValue:    "",
		ConfDataType: domain.ConfTypePassword,
	}
	out := FromConfiguration(row, false)
	if out.ConfValue != "" {
		t.Fatalf("expected empty value to remain empty, got %q", out.ConfValue)
	}
}
