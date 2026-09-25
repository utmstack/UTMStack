package main

// These regressions preserve existing rule predicates while the Linux outcome
// producer changes. The raw helper models ordered extraction; SDK CEL is real.
// Actual playground execution is recorded separately with its own runtime pins.
import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestLinuxActionResultRuleConsumers(t *testing.T) {
	data, err := os.ReadFile("testdata/linux_action_result_rules.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name     string         `json:"name"`
		Rule     string         `json:"rule"`
		Raw      string         `json:"raw"`
		Match    bool           `json:"match"`
		Result   string         `json:"actionResult"`
		Preserve map[string]any `json:"preserve"`
	}
	if err = json.Unmarshal(data, &cases); err != nil || len(cases) == 0 {
		t.Fatalf("fixtures: %v", err)
	}
	cache := plugins.NewCELCache("linux-outcome-rule-consumers")
	positives, negatives := map[string]int{}, map[string]int{}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			blob, err := utils.ReadPbYaml("../../rules/" + tc.Rule)
			if err != nil {
				t.Fatal(err)
			}
			rule := new(plugins.Rule)
			if err = protojson.Unmarshal(blob, rule); err != nil {
				t.Fatal(err)
			}
			rule.Normalize()
			event := linuxActionResultNormalize(t, tc.Raw)
			if got := gjson.Get(event, "actionResult"); got.String() != tc.Result || (tc.Result == "" && got.Exists()) {
				t.Fatalf("outcome %s want %q", got.Raw, tc.Result)
			}
			for path, want := range tc.Preserve {
				if got := gjson.Get(event, path); !got.Exists() || !reflect.DeepEqual(got.Value(), want) {
					t.Errorf("%s=%v want %v", path, got.Value(), want)
				}
			}
			matched, err := cache.Eval(rule.Where, event)
			if err != nil || matched != tc.Match {
				t.Fatalf("%s predicate=%v want%v: %v", tc.Rule, matched, tc.Match, err)
			}
			if matched {
				positives[tc.Rule]++
				if gaps := fixtureHistoryPlaceholders(rule.Correlation, event); len(gaps) != 0 {
					t.Fatal(gaps)
				}
			} else {
				negatives[tc.Rule]++
			}
			if tc.Rule == "linux/bruteforce_attack.yml" {
				// This existing history counts matching journal-message records from the
				// collector. A USER_AUTH predicate match does not itself prove a history hit,
				// a final successful login, or an alert; keep those claims separate.
				if len(rule.Correlation) != 1 {
					t.Fatal("unexpected brute-force history shape")
				}
				history := rule.Correlation[0]
				if history.IndexPattern != "v11-log-linux-*" || history.Within != "15m" || history.Count != 10 || len(history.With) != 2 {
					t.Fatal("brute-force history scope changed")
				}
				if history.With[0].Field != "dataSource.keyword" || history.With[0].Operator != "filter_term" || history.With[0].Value.GetStringValue() != "{{.dataSource}}" {
					t.Fatal("collector identity changed")
				}
				if history.With[1].Field != "log.message" || history.With[1].Operator != "filter_match" || history.With[1].Value.GetStringValue() != "Failed password" {
					t.Fatal("message history scope changed")
				}
			} else if len(rule.Correlation) != 0 {
				t.Fatal("unexpected history dependency")
			}
		})
	}
	for name := range positives {
		if positives[name] == 0 || negatives[name] == 0 {
			t.Fatalf("%s missing positive/negative controls", name)
		}
	}
	if len(positives) != 3 {
		t.Fatalf("expected three exercised rules, got %d", len(positives))
	}
}
