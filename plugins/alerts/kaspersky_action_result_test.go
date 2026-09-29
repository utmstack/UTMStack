package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/tidwall/gjson"
)

// Synthetic CEF and KSC records cover final decisions and IP roles. The
// isolated EventProcessor parser is replayed separately on the same inputs.
func TestKasperskyActionResultRaw(t *testing.T) {
	var cases []struct {
		Name       string            `json:"name"`
		Raw        string            `json:"raw"`
		DataSource string            `json:"dataSource"`
		Result     string            `json:"result"`
		Expected   map[string]string `json:"expected"`
		Absent     []string          `json:"absent"`
		Initial    map[string]any    `json:"initial"`
	}
	data, err := os.ReadFile("testdata/kaspersky_action_result.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 17 {
		t.Fatal("outcome and parser regression classes are missing")
	}
	config := kaspConfig(t)
	cache := plugins.NewCELCache("kaspersky-final-outcome")
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			out := kaspParse(t, config, tc.Raw, tc.DataSource, cache, tc.Initial)
			if got := gjson.Get(out, "actionResult").String(); got != tc.Result {
				t.Errorf("actionResult = %q; want %q", got, tc.Result)
			}
			for path, want := range tc.Expected {
				got := gjson.Get(out, path)
				if !got.Exists() || got.String() != want {
					t.Errorf("%s = %q; want %q", path, got.String(), want)
				}
			}
			for _, path := range tc.Absent {
				if gjson.Get(out, path).Exists() {
					t.Errorf("unexpected %s", path)
				}
			}
			for _, result := range []string{"success", "failure", "denied"} {
				matched, err := cache.Eval(`equals("actionResult","`+result+`")`, out)
				if err != nil || matched != (tc.Result == result) {
					t.Errorf("outcome predicate %s = %v (%v)", result, matched, err)
				}
			}
		})
	}
}
