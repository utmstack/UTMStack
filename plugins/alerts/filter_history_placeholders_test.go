package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/types/known/structpb"
)

// fixtureHistoryPlaceholders is an opt-in, all-branch preflight. The SDK may not
// visit an OR branch when the parent count passes, so this is deliberately not
// applied to every positive rule fixture. It proves value availability only;
// technology tests must exercise SearchRequest.Execute for query/threshold proof.
func fixtureHistoryPlaceholders(searches []*plugins.SearchRequest, event string) []string {
	var issues []string
	var visit func([]*plugins.SearchRequest, string)
	visit = func(requests []*plugins.SearchRequest, path string) {
		for i, request := range requests {
			location := fmt.Sprintf("%s[%d]", path, i)
			for j, expression := range request.With {
				if expression.Value == nil {
					issues = append(issues, fmt.Sprintf("%s.with[%d]: nil value", location, j))
					continue
				}
				value := expression.Value.GetStringValue()
				if strings.HasPrefix(value, "{{.") && strings.HasSuffix(value, "}}") {
					// Match the pinned SDK's exact-token placeholder lookup.
					field := strings.ReplaceAll(strings.ReplaceAll(value, "{{.", ""), "}}", "")
					if gjson.Get(event, field).Value() == nil {
						issues = append(issues, fmt.Sprintf("%s.with[%d]: unresolved %s", location, j, value))
					}
				}
			}
			visit(request.Or, location+".or")
		}
	}
	visit(searches, "correlation")
	return issues
}

func TestHistoryPlaceholderPreflightIncludesNestedOr(t *testing.T) {
	with := func(field string) *plugins.Expression {
		return &plugins.Expression{Field: "unused-indexed-key", Value: structpb.NewStringValue("{{." + field + "}}")}
	}
	requests := []*plugins.SearchRequest{{
		With: []*plugins.Expression{with("origin.user")},
		Or: []*plugins.SearchRequest{{Or: []*plugins.SearchRequest{{
			With: []*plugins.Expression{with("origin.ip"), with("log.zero"), with("log.false")},
		}}}},
	}}
	issues := fixtureHistoryPlaceholders(requests, `{"origin":{"user":"synthetic"},"log":{"zero":0,"false":false}}`)
	if len(issues) != 1 || !strings.Contains(issues[0], "correlation[0].or[0].or[0].with[0]") || !strings.Contains(issues[0], "origin.ip") {
		t.Fatalf("expected only nested missing IP, got %v", issues)
	}
	if issues := fixtureHistoryPlaceholders(requests, `{"origin":{"user":"synthetic","ip":"192.0.2.1"},"log":{"zero":0,"false":false}}`); len(issues) != 0 {
		t.Fatal(issues)
	}
}
