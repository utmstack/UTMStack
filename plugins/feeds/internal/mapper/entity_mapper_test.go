package mapper

import (
	"errors"
	"testing"
)

func TestBuildEntityFlagsValuesThatDoNotFitTheType(t *testing.T) {
	m := NewEntityMapper()

	for _, c := range []struct{ entityType, value string }{
		{"hostname", "v11-enterprise-19"},
		{"file", `C:\Windows\explorer.exe`},
	} {
		_, _, err := m.BuildEntity(c.entityType, c.value, EntityEnrichmentContext{})
		if !errors.Is(err, ErrInvalidValue) {
			t.Errorf("%s %q: err = %v, want ErrInvalidValue", c.entityType, c.value, err)
		}
	}

	if _, _, err := m.BuildEntity("hostname", "mail.example.com", EntityEnrichmentContext{}); err != nil {
		t.Errorf("a well-formed hostname was rejected: %v", err)
	}
}
