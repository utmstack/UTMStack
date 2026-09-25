package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/tidwall/gjson"
)

// linuxActionResultNormalize runs fabricated collector/journal JSON through the
// committed raw extraction model and ordered filter steps. It does not execute
// the closed EventProcessor, threat-intelligence lookup, or alert creation.
func linuxActionResultNormalize(t *testing.T, raw string) string {
	t.Helper()
	out, issues, err := normalize("../..", Fixture{Filter: "linux/linux.yml", Raw: &raw, DataType: "linux", DataSource: "synthetic-collector"}, plugins.NewCELCache("linux-raw-outcome"))
	if err != nil || len(issues) != 0 {
		t.Fatalf("raw extraction model: %v; CEL issues: %v", err, issues)
	}
	return out
}

func TestLinuxActionResult(t *testing.T) {
	b, err := os.ReadFile("testdata/linux_action_result.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name     string         `json:"name"`
		Raw      string         `json:"raw"`
		Result   string         `json:"result"`
		Preserve map[string]any `json:"preserve"`
		Expected map[string]any `json:"expected"`
		Absent   []string       `json:"absent"`
		IP       *bool          `json:"ipEligible"`
	}
	if err = json.Unmarshal(b, &cases); err != nil || len(cases) == 0 {
		t.Fatalf("cases: %v", err)
	}
	cache := plugins.NewCELCache("linux-result-predicates")
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			out := linuxActionResultNormalize(t, tc.Raw)
			got := gjson.Get(out, "actionResult")
			if got.String() != tc.Result || (tc.Result == "" && got.Exists()) {
				t.Fatalf("actionResult = %s, want %q", got.Raw, tc.Result)
			}
			for _, result := range []string{"success", "failure", "denied"} {
				onlyResult, resultErr := cache.Eval(`equals("actionResult","`+result+`")`, out)
				if resultErr != nil || onlyResult != (tc.Result == result) {
					t.Errorf("SDK result predicate = %v, %v", onlyResult, resultErr)
				}
				match, err := cache.Eval(`equals("actionResult","`+result+`") && (inCIDR("origin.ip","0.0.0.0/0") || inCIDR("origin.ip","::/0"))`, out)
				want := tc.Result == result && tc.IP != nil && *tc.IP
				if err != nil || match != want {
					t.Errorf("SDK %s-and-IP predicate = %v, %v", result, match, err)
				}
			}
			for path, want := range tc.Preserve {
				got := gjson.Get(out, path)
				if !got.Exists() || !reflect.DeepEqual(got.Value(), want) {
					t.Errorf("vendor field %s = %v, want %v", path, got.Value(), want)
				}
			}
			for path, want := range tc.Expected {
				if got := gjson.Get(out, path); !got.Exists() || !reflect.DeepEqual(got.Value(), want) {
					t.Errorf("standard field %s = %v, want %v", path, got.Value(), want)
				}
			}
			for _, path := range tc.Absent {
				if gjson.Get(out, path).Exists() {
					t.Errorf("unexpected field %s", path)
				}
			}
		})
	}
}
