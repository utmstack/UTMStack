package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

// Fabricated webhook payloads follow GitHub's documented workflow_run shape and
// pass through the ordered filter model; the outcome is evaluated with the SDK's
// CEL. Alert payloads follow the documented secret scanning, code scanning and
// Dependabot alert objects; they check the secret scanning rule's condition and
// the flattened repository fields, whose names keep only letters and digits. The EventProcessor json step removes every character except letters,
// digits and dots from top-level keys with utils.SanitizeField, and the shared
// raw model does not, so each payload's top-level keys are cleaned the same way
// first. The isolated parser is replayed separately; this model alone does not
// prove deployed behavior.
func TestGitHubActionResultRaw(t *testing.T) {
	var cases []struct {
		Name   string `json:"name"`
		Raw    string `json:"raw"`
		Result string `json:"result"`
		// Optional: exact normalized values, and whether the secret scanning
		// rule's condition must match.
		Fields         map[string]string `json:"fields"`
		SecretScanning *bool             `json:"secretScanning"`
	}
	data, err := os.ReadFile("testdata/github_action_result.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	cache := plugins.NewCELCache("github-action-result-raw")
	secretScanning := githubRuleWhere(t, "../../rules/github/secret_scanning_alerts.yml")
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
			for _, value := range []string{"success", "failure", "denied", "failed"} {
				got, err := cache.Eval(`equals("actionResult","`+value+`")`, out)
				if err != nil || got != (tc.Result == value) {
					t.Errorf("outcome predicate %s = %v (%v)", value, got, err)
				}
			}
			for path, want := range tc.Fields {
				if got := gjson.Get(out, path).String(); got != want {
					t.Errorf("%s = %q; want %q", path, got, want)
				}
			}
			if tc.SecretScanning != nil {
				got, err := cache.Eval(secretScanning, out)
				if err != nil || got != *tc.SecretScanning {
					t.Errorf("secret scanning rule = %v (%v); want %v", got, err, *tc.SecretScanning)
				}
			}
		})
	}
}

// githubRuleWhere reads a rule the way the SDK does and returns its condition.
func githubRuleWhere(t *testing.T, path string) string {
	t.Helper()
	b, err := utils.ReadPbYaml(path)
	if err != nil {
		t.Fatal(err)
	}
	rule := new(plugins.Rule)
	if err := protojson.Unmarshal(b, rule); err != nil {
		t.Fatal(err)
	}
	return rule.Where
}

// sanitizedTopLevelKeys cleans a JSON object's top-level keys as the
// EventProcessor json step does. Nested keys keep their spelling.
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
