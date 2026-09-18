package main

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"

	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/tidwall/gjson"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAlertGroupingValue(t *testing.T) {
	eventLog, err := structpb.NewStruct(map[string]any{"eventCode": 4625, "isFailure": true, "accounts": []any{"alice", "bob"}})
	if err != nil {
		t.Fatal(err)
	}
	alert := &plugins.Alert{
		Adversary: &plugins.Side{Ip: "203.0.113.7"},
		Target:    &plugins.Side{Host: "dc01"},
		Events: []*plugins.Event{
			{Action: "previous", Origin: &plugins.Side{User: "previous-user"}},
			{Action: "login", Origin: &plugins.Side{User: "alice"}, Log: eventLog.Fields},
		},
	}
	serialized, err := utils.ProtoMessageToString(alert)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ field, want string }{
		{"adversary.ip", "203.0.113.7"}, {"target.host", "dc01"},
		{"lastEvent.action", "login"}, {"lastEvent.origin.user", "alice"},
		{"lastEvent.log.eventCode", "4625"}, {"lastEvent.log.isFailure", "true"},
		{"lastEvent.log.accounts.1", "bob"},
	} {
		t.Run(tc.field, func(t *testing.T) {
			got := alertGroupingValue(*serialized, tc.field)
			if !got.Exists() || got.String() != tc.want {
				t.Fatalf("got %v, want %s", got, tc.want)
			}
		})
	}
	for _, field := range []string{"lastEvent.log.missing", "origin.ip"} {
		if alertGroupingValue(*serialized, field).Exists() {
			t.Errorf("unexpected value for %s", field)
		}
	}
}
func TestAlertGroupingValueWithoutEvents(t *testing.T) {
	for _, input := range []string{`{}`, `{"events":[]}`, `{"events":[null]}`} {
		if alertGroupingValue(input, "lastEvent.origin.ip").Exists() {
			t.Errorf("unexpected lastEvent for %s", input)
		}
	}
}

func TestScalarGroupingValue(t *testing.T) {
	for _, tc := range []struct {
		input string
		valid bool
	}{
		{`"alice"`, true}, {`0`, true}, {`false`, true}, {`true`, true},
		{`null`, false}, {`[]`, false}, {`["alice"]`, false}, {`{"user":"alice"}`, false},
	} {
		if _, ok := scalarGroupingValue(gjson.Parse(tc.input)); ok != tc.valid {
			t.Errorf("scalarGroupingValue(%s) = %v, want %v", tc.input, ok, tc.valid)
		}
	}
}

// Exercise the same SDK query construction used by both production callers.
// Empty indices keep this unit test offline: live mapping/.keyword resolution
// remains a separate OpenSearch integration check.
func TestAlertGroupingIndexedTerms(t *testing.T) {
	log, err := structpb.NewStruct(map[string]any{
		"accounts": []any{map[string]any{"id": "first"}, map[string]any{"id": "selected"}},
		"attempts": 0, "blocked": false, "object": map[string]any{"nested": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	alert := &plugins.Alert{
		Name:      "Synthetic grouping contract",
		Adversary: &plugins.Side{Ip: "203.0.113.7"},
		Events: []*plugins.Event{
			{Action: "previous-action", Origin: &plugins.Side{User: "previous-user", Ip: "192.0.2.2"}},
			{Action: "final-action", Origin: &plugins.Side{User: "final-user"}, Log: log.Fields},
		},
	}
	wire, err := utils.ProtoMessageToString(alert)
	if err != nil {
		t.Fatal(err)
	}
	builder := sdkos.NewBoolBuilder(context.Background(), nil, "grouping-query-test")
	fields := []string{
		"lastEvent.action.keyword", "lastEvent.origin.user", "adversary.ip",
		"lastEvent.log.accounts.1.id.keyword", "lastEvent.log.attempts", "lastEvent.log.blocked",
		"lastEvent.origin.ip", // Only present on the older event: do not fall back.
		"lastEvent.log.accounts", "lastEvent.log.object", "lastEvent.log.missing",
	}
	if !addAlertGroupingTerms(builder, *wire, fields) {
		t.Fatal("expected usable grouping terms")
	}
	query, errors := builder.BuildWithErrors()
	if len(errors) != 0 {
		t.Fatal(errors)
	}
	encoded, err := json.Marshal(query)
	if err != nil {
		t.Fatal(err)
	}
	terms := map[string]any{}
	for _, clause := range query.Bool.Filter {
		for field, term := range clause.Term {
			if strings.HasPrefix(field, "events.") {
				t.Errorf("wire path leaked into indexed query: %s", field)
			}
			terms[field] = term["value"]
		}
	}
	want := map[string]any{
		"lastEvent.action": "final-action", "lastEvent.origin.user": "final-user",
		"adversary.ip": "203.0.113.7", "lastEvent.log.accounts.id": "selected",
		"lastEvent.log.attempts": float64(0), "lastEvent.log.blocked": false,
	}
	if !reflect.DeepEqual(terms, want) {
		t.Fatalf("indexed terms differ: got %s, want %v", encoded, want)
	}
}

func TestAlertGroupingDoesNotEnableNameOnlyQuery(t *testing.T) {
	for _, wire := range []string{
		`{}`, `{"events":[]}`,
		`{"events":[{"action":"older"},{}]}`,
		`{"events":[{"log":{"items":[1],"object":{"id":"x"},"empty":null}}]}`,
	} {
		builder := sdkos.NewBoolBuilder(context.Background(), nil, "grouping-query-test")
		builder.FilterTerm("name", "Synthetic grouping contract")
		if addAlertGroupingTerms(builder, wire, []string{
			"lastEvent.action", "lastEvent.log.items", "lastEvent.log.object", "lastEvent.log.empty",
		}) {
			t.Fatalf("caller would execute a name-only query for %s", wire)
		}
		query := builder.Build()
		if len(query.Bool.Filter) != 1 {
			t.Fatalf("unexpected grouping term: %v", query)
		}
	}
}
