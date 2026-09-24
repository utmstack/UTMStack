package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
)

// Fabricated webhook payloads follow GitHub's documented workflow_run and push
// shapes and pass through the ordered filter model; the outcome is evaluated with
// the SDK's CEL. The EventProcessor json step cleans top-level keys with
// utils.SanitizeField, which since go-sdk v1.1.35 keeps letters, digits, dots and
// underscores, and the shared raw model does not clean them, so each payload's
// top-level keys are cleaned the same way first. The isolated parser is replayed
// separately; this model alone does not prove deployed behavior.
func TestGitHubActionResultRaw(t *testing.T) {
	var cases []struct {
		Name     string            `json:"name"`
		Raw      string            `json:"raw"`
		Result   string            `json:"result"`
		Expected map[string]string `json:"expected"`
	}
	data, err := os.ReadFile("testdata/github_action_result.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	cache := plugins.NewCELCache("github-action-result-raw")
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			raw := sanitizedTopLevelKeys(t, tc.Raw)
			fixture := Fixture{Filter: "github/github.yml", Raw: &raw, DataType: "github", DataSource: "fabricated-collector"}
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
			for _, value := range []string{"success", "failure", "denied", "failed"} {
				got, err := cache.Eval(`equals("actionResult","`+value+`")`, out)
				if err != nil || got != (tc.Result == value) {
					t.Errorf("outcome predicate %s = %v (%v)", value, got, err)
				}
			}
		})
	}
}

// sanitizedTopLevelKeys cleans a JSON object's top-level keys as the
// EventProcessor json step does. Nested keys are not cleaned.
func sanitizedTopLevelKeys(t *testing.T, raw string) string {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &object); err != nil {
		t.Fatal(err)
	}
	cleaned := make(map[string]json.RawMessage, len(object))
	for key, value := range object {
		utils.SanitizeField(&key)
		cleaned[key] = value
	}
	b, err := json.Marshal(cleaned)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
