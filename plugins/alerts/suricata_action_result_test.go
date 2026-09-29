package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

// The ordered normalization model uses the reviewed SDK's CEL and Event schema.
// Isolated playground cases separately execute raw JSON and the actual plugins.
func TestSuricataActionResult(t *testing.T) {
	var cases []struct {
		Name     string            `json:"name"`
		Raw      string            `json:"raw"`
		Result   string            `json:"result"`
		Expected map[string]string `json:"expected"`
		Absent   []string          `json:"absent"`
	}
	data, err := os.ReadFile("testdata/suricata_action_result.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 18 {
		t.Fatal("final verdict and flow state classes are missing")
	}
	cache := plugins.NewCELCache("suricata-final-outcome")
	rules := map[string]*plugins.Rule{}
	for _, name := range []string{"high_severity_suricata_alerts_were_detected", "medium_severity_suricata_alerts_were_detected"} {
		b, err := utils.ReadPbYaml(filepath.Join("../..", "rules/suricata", name+".yml"))
		if err != nil {
			t.Fatal(err)
		}
		rule := new(plugins.Rule)
		if err := protojson.Unmarshal(b, rule); err != nil {
			t.Fatal(err)
		}
		rule.Normalize()
		rules[name] = rule
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			parsed, err := fixtureJSON(tc.Raw)
			if err != nil {
				t.Fatal(err)
			}
			fixture := Fixture{Filter: "suricata/suricata.yml", Input: map[string]any{
				"raw": tc.Raw, "dataType": "suricata", "dataSource": "synthetic-suricata", "log": parsed,
			}}
			out, issues, err := normalize("../..", fixture, cache)
			if err != nil || len(issues) != 0 {
				t.Fatalf("normalize: %v / %v", err, issues)
			}
			if got := gjson.Get(out, "actionResult").String(); got != tc.Result {
				t.Errorf("actionResult = %q; want %q", got, tc.Result)
			}
			for path, want := range tc.Expected {
				if got := gjson.Get(out, path); !got.Exists() || got.String() != want {
					t.Errorf("%s = %q; want %q", path, got.String(), want)
				}
			}
			for _, path := range tc.Absent {
				if gjson.Get(out, path).Exists() {
					t.Errorf("unexpected %s", path)
				}
			}
			for _, value := range []string{"success", "failure", "denied"} {
				got, err := cache.Eval(`equals("actionResult","`+value+`")`, out)
				if err != nil || got != (tc.Result == value) {
					t.Errorf("predicate %s = %v (%v)", value, got, err)
				}
			}
			if tc.Name == "allowed-alert-is-not-final-verdict" ||
				tc.Name == "allowed-alert-final-drop" || tc.Name == "blocked-alert-no-verdict" {
				wantHigh := tc.Name == "allowed-alert-is-not-final-verdict"
				wantMedium := tc.Name == "allowed-alert-final-drop"
				for name, want := range map[string]bool{
					"high_severity_suricata_alerts_were_detected":   wantHigh,
					"medium_severity_suricata_alerts_were_detected": wantMedium,
				} {
					got, err := cache.Eval(rules[name].Where, out)
					if err != nil || got != want {
						t.Errorf("%s predicate = %v (%v); want %v", name, got, err, want)
					}
				}
			}
		})
	}
}
