package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
)

var alertFields = []string{"@timestamp", "id", "name", "severity", "dataSource", "dataType", "origin.host"}

func TestUnknownFieldMessage(t *testing.T) {
	tests := []struct {
		name string
		spec domain.Spec
		want string // substring; empty means the spec is fine
	}{
		{
			name: "a real dimension passes",
			spec: domain.Spec{Dataset: "alerts", Chart: domain.ChartCategory, Dimension: "dataSource"},
		},
		{
			name: "snake_case guess points at the camelCase field",
			spec: domain.Spec{Dataset: "alerts", Chart: domain.ChartCategory, Dimension: "data_source"},
			want: `did you mean "dataSource"`,
		},
		{
			name: "an invented field is named, with no suggestion",
			spec: domain.Spec{Dataset: "alerts", Chart: domain.ChartCategory, Dimension: "agent.name"},
			want: `dimension "agent.name" is not a field of the alerts dataset`,
		},
		{
			name: "a filter on an unknown field is caught",
			spec: domain.Spec{Dataset: "alerts", Chart: domain.ChartMetric, Filters: []domain.Filter{{Field: "sev", Op: "eq", Value: "high"}}},
			want: `filter field "sev"`,
		},
		{
			name: "table columns are checked",
			spec: domain.Spec{Dataset: "alerts", Chart: domain.ChartTable, Columns: []string{"name", "hostname"}},
			want: `column "hostname"`,
		},
		{
			name: "a metric with no fields has nothing to check",
			spec: domain.Spec{Dataset: "alerts", Chart: domain.ChartMetric},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unknownFieldMessage(tt.spec, alertFields)
			if tt.want == "" {
				if got != "" {
					t.Fatalf("expected the spec to pass, got %q", got)
				}
				return
			}
			if !strings.Contains(got, tt.want) {
				t.Fatalf("message %q does not contain %q", got, tt.want)
			}
			if !strings.Contains(got, "store.dataset.fields") {
				t.Fatalf("message %q should tell the author where the valid names are", got)
			}
		})
	}
}

func TestInventedFieldGetsNoSuggestion(t *testing.T) {
	got := unknownFieldMessage(domain.Spec{Dataset: "logs", Chart: domain.ChartCategory, Dimension: "agent.name"}, alertFields)
	if strings.Contains(got, "did you mean") {
		t.Fatalf("no real field resembles agent.name, but got %q", got)
	}
}

func TestWidgetIsNotSavedBlindWhenTheStoreIsUnavailable(t *testing.T) {
	m := &Module{deps: &Deps{}}

	err := m.checkWidgetSpec(context.Background(), `{"dataset":"alerts","chart":"metric","metric":{"agg":"count"}}`)

	if err == nil || !strings.Contains(err.Error(), "not saved") {
		t.Fatalf("a widget that cannot be checked must be refused, got %v", err)
	}
}

func TestMalformedSpecIsRefusedBeforeTheStoreIsTouched(t *testing.T) {
	m := &Module{deps: &Deps{}}

	for name, spec := range map[string]string{
		"not json":        `{`,
		"unknown dataset": `{"dataset":"nope","chart":"metric"}`,
		"category no dim": `{"dataset":"alerts","chart":"category"}`,
		"unsupported agg": `{"dataset":"alerts","chart":"metric","metric":{"agg":"sum"}}`,
	} {
		if err := m.checkWidgetSpec(context.Background(), spec); err == nil || strings.Contains(err.Error(), "not available") {
			t.Errorf("%s: expected a spec error before any store access, got %v", name, err)
		}
	}
}
