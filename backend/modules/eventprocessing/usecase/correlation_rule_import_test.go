package usecase

import (
	"errors"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
)

func TestValidateImportedRuleCorrelationSteps(t *testing.T) {
	base := func(correlation string) []byte {
		return []byte("dataTypes: [linux]\nname: r\nwhere: 'true'\n" + correlation)
	}

	cases := []struct {
		name        string
		correlation string
		wantInvalid bool
	}{
		{"v12 step without an index pattern", "correlation:\n  - with:\n      - {field: dataSource, operator: filter_term, value: x}\n    within: 1h\n    count: 2\n", false},
		{"step that still names an index pattern", "correlation:\n  - indexPattern: v11-log-*\n    within: 1h\n    count: 2\n", false},
		{"unparseable window", "correlation:\n  - within: soon\n    count: 2\n", true},
		{"unknown operator", "correlation:\n  - with:\n      - {field: a, operator: nope, value: x}\n    count: 2\n", true},
		{"condition without a field", "correlation:\n  - with:\n      - {operator: filter_term, value: x}\n    count: 2\n", true},
	}

	for _, c := range cases {
		rule, err := parseRuleYAML(base(c.correlation))
		if err != nil {
			t.Fatalf("%s: parse: %v", c.name, err)
		}
		err = validateImportedRule(rule)
		if c.wantInvalid != errors.Is(err, domain.ErrCorrelationRuleInvalidContent) {
			t.Errorf("%s: err = %v, wantInvalid %v", c.name, err, c.wantInvalid)
		}
	}
}
