package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/tidwall/gjson"
)

// These fabricated raw records exercise the complete checked-in O365 filter
// model and the pinned SDK CEL. The EventProcessor playground separately runs
// the real parser; neither test claims customer alert indexing or notifications.
func TestO365AwarenessRawPredicatesAndGrouping(t *testing.T) {
	var cases []struct {
		Name     string         `json:"name"`
		Raw      string         `json:"raw"`
		Rule     string         `json:"rule"`
		Match    bool           `json:"match"`
		Expected map[string]any `json:"expected"`
		Absent   []string       `json:"absent"`
	}
	b, err := os.ReadFile("testdata/o365_awareness.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 50 {
		t.Fatalf("awareness fixture set unexpectedly small: %d", len(cases))
	}
	paths := map[string]*plugins.Rule{}
	for _, tc := range cases {
		if _, ok := paths[tc.Rule]; ok {
			continue
		}
		if !strings.HasPrefix(tc.Rule, "office365/") || !strings.HasSuffix(tc.Rule, ".yml") {
			t.Fatalf("unexpected rule path %s", tc.Rule)
		}
		stem := strings.TrimSuffix(strings.TrimPrefix(tc.Rule, "office365/"), ".yml")
		paths[tc.Rule] = o365OutcomeRule(t, stem)
	}
	if len(paths) != 8 {
		t.Fatalf("want eight shipped awareness rules, got %d", len(paths))
	}
	// One top-level alert per acting account and action; later changes become children, which the
	// rule flood guard does not count. Grouping per object passed 50 top-level alerts a day.
	want := []string{"lastEvent.tenantId", "lastEvent.log.OrganizationId", "dataSource", "adversary.user", "lastEvent.action"}
	for path, rule := range paths {
		if !reflect.DeepEqual(rule.GroupBy, want) {
			t.Fatalf("%s: grouping paths %v, want %v", path, rule.GroupBy, want)
		}
		if len(rule.DeduplicateBy) != 0 || len(rule.Correlation) != 0 || len(rule.AfterEvents) != 0 {
			t.Fatalf("%s: awareness alerts should retain every change without history thresholds or suppression", path)
		}
		if rule.Impact.GetConfidentiality() != 1 || rule.Impact.GetIntegrity() != 1 || rule.Impact.GetAvailability() != 0 || rule.Adversary != "origin" {
			t.Fatalf("%s: invalid awareness impact or attribution", path)
		}
	}
	cache := plugins.NewCELCache("o365-awareness")
	seen := map[string]bool{}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			if seen[tc.Name] {
				t.Fatal("duplicate fixture")
			}
			seen[tc.Name] = true
			event := o365ActionResultNormalize(t, tc.Raw)
			for path, want := range tc.Expected {
				got := gjson.Get(event, path)
				if !got.Exists() || !reflect.DeepEqual(got.Value(), want) {
					t.Errorf("normalized %s = %v, want %v", path, got.Value(), want)
				}
			}
			for _, path := range tc.Absent {
				if gjson.Get(event, path).Exists() {
					t.Errorf("unexpected normalized %s", path)
				}
			}
			matches := 0
			for path, rule := range paths {
				got, err := cache.Eval(rule.Where, event)
				if err != nil {
					t.Fatalf("%s: CEL: %v", path, err)
				}
				want := path == tc.Rule && tc.Match
				if got != want {
					t.Errorf("%s: predicate = %v, want %v", path, got, want)
				}
				if got {
					matches++
				}
			}
			if matches > 1 {
				t.Errorf("one administrative change matched %d awareness rules", matches)
			}
		})
	}
}
