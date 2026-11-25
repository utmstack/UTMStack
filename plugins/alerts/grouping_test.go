package main

import (
	"context"
	"encoding/json"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"github.com/utmstack/go-sdk/plugins"
	"google.golang.org/protobuf/encoding/protojson"
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
			{Action: "login", Origin: &plugins.Side{User: "alice"}, Event: eventLog.Fields},
		},
	}
	serialized, err := utils.ProtoMessageToString(alert)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ field, want string }{
		{"adversary.ip", "203.0.113.7"}, {"target.host", "dc01"},
		{"lastEvent.action", "login"}, {"lastEvent.origin.user", "alice"},
		{"lastEvent.event.eventCode", "4625"}, {"lastEvent.event.isFailure", "true"},
		{"lastEvent.event.accounts.1", "bob"},
	} {
		t.Run(tc.field, func(t *testing.T) {
			got := alertGroupingValue(*serialized, tc.field)
			if !got.Exists() || got.String() != tc.want {
				t.Fatalf("got %v, want %s", got, tc.want)
			}
		})
	}
	for _, field := range []string{"lastEvent.event.missing", "origin.ip"} {
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
			{Action: "final-action", Origin: &plugins.Side{User: "final-user"}, Event: log.Fields},
		},
	}
	wire, err := utils.ProtoMessageToString(alert)
	if err != nil {
		t.Fatal(err)
	}
	builder := sdkos.NewBoolBuilder(context.Background(), nil, "grouping-query-test")
	fields := []string{
		"lastEvent.action.keyword", "lastEvent.origin.user", "adversary.ip",
		"lastEvent.event.accounts.1.id.keyword", "lastEvent.event.attempts", "lastEvent.event.blocked",
		"lastEvent.origin.ip", // Only present on the older event: do not fall back.
		"lastEvent.event.accounts", "lastEvent.event.object", "lastEvent.event.missing",
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
		"adversary.ip": "203.0.113.7", "lastEvent.event.accounts.id": "selected",
		"lastEvent.event.attempts": float64(0), "lastEvent.event.blocked": false,
	}
	if !reflect.DeepEqual(terms, want) {
		t.Fatalf("indexed terms differ: got %s, want %v", encoded, want)
	}
}

func TestAlertGroupingDoesNotEnableNameOnlyQuery(t *testing.T) {
	for _, wire := range []string{
		`{}`, `{"events":[]}`,
		`{"events":[{"action":"older"},{}]}`,
		`{"events":[{"event":{"items":[1],"object":{"id":"x"},"empty":null}}]}`,
	} {
		builder := sdkos.NewBoolBuilder(context.Background(), nil, "grouping-query-test")
		builder.FilterTerm("name", "Synthetic grouping contract")
		if addAlertGroupingTerms(builder, wire, []string{
			"lastEvent.action", "lastEvent.event.items", "lastEvent.event.object", "lastEvent.event.empty",
		}) {
			t.Fatalf("caller would execute a name-only query for %s", wire)
		}
		query := builder.Build()
		if len(query.Bool.Filter) != 1 {
			t.Fatalf("unexpected grouping term: %v", query)
		}
	}
}

func TestBuildDuplicateFilterExcludesCompletedStatus(t *testing.T) {
	alert := &plugins.Alert{
		Name:       "Linux SSH Bruteforce Attack",
		DataSource: "web01",
		Adversary:  &plugins.Side{Ip: "203.0.113.42"},
	}
	wire, err := utils.ProtoMessageToString(alert)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	builder := sdkos.NewBoolBuilder(context.Background(), nil, "duplicate-filter-test")
	if !buildDuplicateFilter(builder, alert.Name, *wire, []string{"dataSource", "adversary.ip"}, now) {
		t.Fatal("expected buildDuplicateFilter to return true")
	}

	query, errs := builder.BuildWithErrors()
	if len(errs) != 0 {
		t.Fatalf("unexpected builder errors: %v", errs)
	}

	var foundMustNotCompleted bool
	for _, clause := range query.Bool.MustNot {
		if term, ok := clause.Term["status"]; ok {
			if term["value"] == statusCompleted || term["value"] == float64(statusCompleted) {
				foundMustNotCompleted = true
			}
		}
	}
	if !foundMustNotCompleted {
		encoded, _ := json.Marshal(query)
		t.Fatalf("expected must_not status=%d in duplicate filter, got: %s", statusCompleted, encoded)
	}

	terms := map[string]any{}
	for _, clause := range query.Bool.Filter {
		for field, term := range clause.Term {
			terms[field] = term["value"]
		}
	}
	wantTerms := map[string]any{
		"name":         "Linux SSH Bruteforce Attack",
		"dataSource":   "web01",
		"adversary.ip": "203.0.113.42",
	}
	if !reflect.DeepEqual(terms, wantTerms) {
		encoded, _ := json.Marshal(query)
		t.Fatalf("duplicate filter terms differ: got %v (%s), want %v", terms, encoded, wantTerms)
	}
}

func TestNoRuleGroupsByBareOrigin(t *testing.T) {
	err := filepath.WalkDir("../../rules", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".yml") {
			return err
		}
		blob, err := utils.ReadPbYaml(path)
		if err != nil {
			t.Errorf("%s: failed to read rule yaml: %v", path, err)
			return nil
		}
		r := new(plugins.Rule)
		if err := protojson.Unmarshal(blob, r); err != nil {
			t.Errorf("%s: failed to unmarshal rule: %v", path, err)
			return nil
		}
		r.Normalize()
		for _, field := range append(append([]string(nil), r.GroupBy...), r.DeduplicateBy...) {
			if strings.HasPrefix(field, "origin.") {
				t.Errorf("%s: invalid bare origin grouping field %q (alerts have adversary.*/target.* or lastEvent.origin.*)", path, field)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
