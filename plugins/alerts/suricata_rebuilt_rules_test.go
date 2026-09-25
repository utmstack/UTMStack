package main

// The rebuilt DNS tunneling, ICMP tunneling and threat-intelligence rules read
// fields that Suricata EVE records carry after the filter: log.eventType, the
// dns object (version 2 and version 3 layouts), ICMP flow counters and alert
// metadata. Fabricated EVE records go through the ordered filter model (the JSON
// step is modeled by fixtureJSON) and the real SDK CEL. deduplicateBy keys are
// resolved the way grouping.go resolves them on an alert built with
// adversary=origin, as these rules declare.
import (
	"encoding/json"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

var suricataRebuiltRules = []string{
	"nids/suricata/dns_tunneling_detection.yml",
	"nids/suricata/icmp_tunneling_detection.yml",
	"nids/suricata/threat_intelligence_iocs.yml",
}

func TestSuricataRebuiltRules(t *testing.T) {
	data, err := os.ReadFile("testdata/suricata_rebuilt_rules.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name  string   `json:"name"`
		Rule  string   `json:"rule"`
		Raw   string   `json:"raw"`
		Match bool     `json:"match"`
		Keys  []string `json:"deduplicateKeys"`
	}
	if err = json.Unmarshal(data, &cases); err != nil || len(cases) == 0 {
		t.Fatalf("fixtures: %v", err)
	}
	rules := map[string]*plugins.Rule{}
	for _, path := range suricataRebuiltRules {
		blob, err := utils.ReadPbYaml("../../rules/" + path)
		if err != nil {
			t.Fatal(err)
		}
		rule := new(plugins.Rule)
		if err = protojson.Unmarshal(blob, rule); err != nil {
			t.Fatal(err)
		}
		rule.Normalize()
		if len(rule.Correlation) != 0 || len(rule.GroupBy) != 0 || len(rule.DeduplicateBy) < 2 ||
			rule.DeduplicateBy[0] != "dataSource" || rule.Adversary != "origin" {
			t.Fatalf("%s: want adversary origin, deduplicateBy starting with dataSource, no groupBy and no history", path)
		}
		rules[path] = rule
	}
	cache := plugins.NewCELCache("suricata-rebuilt-rules")
	positives, negatives := map[string]int{}, map[string]int{}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			rule, ok := rules[tc.Rule]
			if !ok {
				t.Fatalf("unknown rule %s", tc.Rule)
			}
			parsed, err := fixtureJSON(tc.Raw)
			if err != nil {
				t.Fatal(err)
			}
			fixture := Fixture{Filter: "suricata/suricata.yml", Input: map[string]any{
				"raw": tc.Raw, "dataType": "suricata", "dataSource": "synthetic-suricata", "log": parsed,
			}}
			event, issues, err := normalize("../..", fixture, cache)
			if err != nil || len(issues) != 0 {
				t.Fatalf("normalize: %v / %v", err, issues)
			}
			matched, err := cache.Eval(rule.Where, event)
			if err != nil || matched != tc.Match {
				t.Fatalf("%s predicate=%v want %v: %v\n%s", tc.Rule, matched, tc.Match, err, event)
			}
			if !matched {
				negatives[tc.Rule]++
				return
			}
			positives[tc.Rule]++
			side := func(path string) string {
				if v := gjson.Get(event, path); v.Exists() {
					return v.Raw
				}
				return "{}"
			}
			alert := `{"name":` + strconv.Quote(rule.Name) + `,"dataSource":"synthetic-suricata","adversary":` +
				side("origin") + `,"target":` + side("target") + `,"events":[` + event + `]}`
			var resolved []string
			for _, field := range rule.DeduplicateBy {
				if _, ok := scalarGroupingValue(alertGroupingValue(alert, field)); ok {
					resolved = append(resolved, field)
				}
			}
			if !reflect.DeepEqual(resolved, tc.Keys) {
				t.Errorf("deduplicateBy resolves %v, want %v", resolved, tc.Keys)
			}
		})
	}
	for _, path := range suricataRebuiltRules {
		if positives[path] == 0 || negatives[path] == 0 {
			t.Errorf("%s: missing positive or negative cases", path)
		}
	}
}
