package common_models

import (
	"errors"
	"testing"
)

func TestValidateFieldAcceptsRealFieldNames(t *testing.T) {
	for _, f := range []string{"@timestamp", "dataSource", "log.eventDataTargetUserSid", "origin.ip", "log.http-method", "log.a:b", "lastEvent.log.x_y"} {
		if err := ValidateField(f); err != nil {
			t.Errorf("%q refused: %v", f, err)
		}
	}
}

func TestValidateFieldRefusesWhatQuotingHasToSurvive(t *testing.T) {
	for name, f := range map[string]string{
		"empty":        "",
		"trailing dot": "log.",
		"leading dot":  ".log",
		"double dot":   "log..x",
		"backtick":     "log.a`b",
		"backslash":    "log.a\\",
		"injection":    "x\\`,(SELECT 42) AS y -- ",
		"control char": "log.a\nb",
		"too long":     string(make([]byte, 300)),
	} {
		err := ValidateField(f)
		if !errors.Is(err, ErrInvalidField) {
			t.Errorf("%s: err = %v, want ErrInvalidField", name, err)
		}
	}
}

func TestToStoreFiltersRefusesABadFieldButNotATextSearch(t *testing.T) {
	if _, err := ToStoreFilters([]FilterType{{Field: "log..x", Operator: OpEquals, Value: "1"}}); !errors.Is(err, ErrInvalidField) {
		t.Errorf("err = %v", err)
	}
	if _, err := ToStoreFilters([]FilterType{{Field: "", Operator: OpIsInFields, Value: "needle"}}); err != nil {
		t.Errorf("a text search was refused: %v", err)
	}
}
