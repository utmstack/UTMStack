package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
)

// Fabricated native-agent payloads check the ordered filter and pinned SDK.
// Separate isolated-parser runs verify extraction with the actual plugins.
func TestWindowsActionResultRaw(t *testing.T) {
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
	data, err := os.ReadFile("testdata/windows_action_result.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 20 {
		t.Fatal("connection and logon outcome classes are missing")
	}
	config := winConfig(t)
	cache := plugins.NewCELCache("windows-action-result-raw")
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			out := winParse(t, config, tc.Raw, tc.DataSource, cache)
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
				b, err := utils.ReadPbYaml(filepath.Join("../..", "rules", tc.Rule))
				if err != nil {
					t.Fatal(err)
				}
				got, err := cache.Eval(gjson.GetBytes(b, "where").String(), out)
				if err != nil || got != tc.Match {
					t.Errorf("rule %s = %v (%v); want %v", tc.Rule, got, err, tc.Match)
				}
			}
		})
	}
}
