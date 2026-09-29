package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

// o365ActionResultNormalize runs fabricated raw JSON through the committed
// extraction/normalization model. Only external geolocation is omitted; it does
// not produce actionResult. This does not execute the closed EventProcessor,
// threat-intelligence membership, or alert creation.
func o365ActionResultNormalize(t *testing.T, raw string) string {
	t.Helper()
	b, err := utils.ReadPbYaml("../../filters/office365/o365.yml")
	if err != nil {
		t.Fatal(err)
	}
	cfg := new(plugins.Config)
	if err = protojson.Unmarshal(b, cfg); err != nil {
		t.Fatal(err)
	}
	for _, stage := range cfg.Pipeline {
		steps := stage.Steps[:0]
		for _, step := range stage.Steps {
			if step.Dynamic != nil {
				if step.Dynamic.Plugin != "com.utmstack.geolocation" {
					t.Fatalf("unmodeled dynamic producer: %s", step.Dynamic.Plugin)
				}
				continue
			}
			steps = append(steps, step)
		}
		stage.Steps = steps
	}
	root := t.TempDir()
	if err = os.Mkdir(filepath.Join(root, "filters"), 0700); err != nil {
		t.Fatal(err)
	}
	b, err = protojson.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "filters", "o365.yml"), b, 0600); err != nil {
		t.Fatal(err)
	}
	out, issues, err := normalize(root, Fixture{Filter: "o365.yml", Raw: &raw, DataType: "o365", DataSource: "synthetic-collector"}, plugins.NewCELCache("o365-raw-outcome"))
	if err != nil || len(issues) != 0 {
		t.Fatalf("raw extraction model: %v; CEL issues: %v", err, issues)
	}
	return out
}

func TestO365ActionResult(t *testing.T) {
	b, err := os.ReadFile("testdata/o365_action_result.json")
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
	cache := plugins.NewCELCache("o365-result-predicates")
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			out := o365ActionResultNormalize(t, tc.Raw)
			got := gjson.Get(out, "actionResult")
			if got.String() != tc.Result || (tc.Result == "" && got.Exists()) {
				t.Fatalf("actionResult = %s, want %q", got.Raw, tc.Result)
			}
			for _, result := range []string{"success", "failure", "denied"} {
				match, err := cache.Eval(`equals("actionResult","`+result+`") && inCIDR("origin.ip","0.0.0.0/0")`, out)
				want := tc.Result == result && (tc.IP == nil || *tc.IP)
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
