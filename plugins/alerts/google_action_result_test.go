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

// Fabricated LogEntry payloads exercise the ordered filter model and the
// reviewed SDK's actual CEL predicates. The isolated parser is replayed
// separately; this model alone does not prove deployed behavior.
func TestGoogleActionResultRaw(t *testing.T) {
	var cases []struct {
		Name     string            `json:"name"`
		Raw      string            `json:"raw"`
		Result   string            `json:"result"`
		Expected map[string]string `json:"expected"`
		Absent   []string          `json:"absent"`
		Rule     string            `json:"rule"`
		Match    bool              `json:"match"`
	}
	data, err := os.ReadFile("testdata/google_action_result.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 40 {
		t.Fatal("outcome classes are missing")
	}
	cache := plugins.NewCELCache("google-action-result-raw")
	root := googleModelRoot(t)
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			fixture := Fixture{Filter: "google/gcp.yml", Raw: &tc.Raw, DataType: "google", DataSource: "fabricated-collector"}
			out, issues, err := normalize(root, fixture, cache)
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

// googleModelRoot returns a root holding the Google filter without its dynamic
// steps. Geolocation enrichment is external and the raw model does not run it,
// so records with a caller address would otherwise stop at that step.
func googleModelRoot(t *testing.T) string {
	t.Helper()
	b, err := utils.ReadPbYaml(filepath.Join("../..", "filters", "google", "gcp.yml"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := new(plugins.Config)
	if err := protojson.Unmarshal(b, cfg); err != nil {
		t.Fatal(err)
	}
	for _, stage := range cfg.Pipeline {
		steps := stage.Steps[:0]
		for _, step := range stage.Steps {
			if step.Dynamic == nil {
				steps = append(steps, step)
			}
		}
		stage.Steps = steps
	}
	out, err := protojson.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	dir := filepath.Join(root, "filters", "google")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "gcp.yml"), out, 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}
