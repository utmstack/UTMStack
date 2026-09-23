package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/tidwall/gjson"
)

// Fabricated CEF task-status payloads exercise the ordered filter model and
// the v11 SDK's CEL predicates. The isolated parser is replayed separately.
func TestBitdefenderActionResultRaw(t *testing.T) {
	var cases []struct {
		Name       string            `json:"name"`
		Raw        string            `json:"raw"`
		DataSource string            `json:"dataSource"`
		Result     string            `json:"result"`
		Expected   map[string]string `json:"expected"`
		Absent     []string          `json:"absent"`
		Rule       string            `json:"rule"`
		Match      bool              `json:"match"`
	}
	data, err := os.ReadFile("testdata/bitdefender_action_result.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 12 {
		t.Fatal("task outcome and contradiction classes are missing")
	}
	config := bitdefConfig(t)
	rules := bitdefRules(t)
	cache := plugins.NewCELCache("bitdefender-task-outcome")
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			out := bitdefParse(t, config, tc.Raw, tc.DataSource, cache)
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
					t.Errorf("outcome predicate %s = %v (%v)", value, got, err)
				}
			}
			if tc.Rule != "" {
				rule := rules["suspicious_exclusions_added"]
				if rule == nil {
					t.Fatal("shipped exclusion rule missing")
				}
				got, err := cache.Eval(rule.Where, out)
				if err != nil || got != tc.Match {
					t.Errorf("rule %s = %v (%v); want %v", tc.Rule, got, err, tc.Match)
				}
			}
		})
	}
}
