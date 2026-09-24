package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"google.golang.org/protobuf/encoding/protojson"
)

// Fabricated Workspace Login Audit records, shaped like Google's documented Cloud
// Logging example, pass through the ordered filter model before the rule is
// evaluated with the SDK's CEL. The isolated parser is replayed separately; this
// model alone does not prove deployed behavior.
func TestGoogleWorkspaceSuspiciousSignInRule(t *testing.T) {
	var cases []struct {
		Name  string `json:"name"`
		Raw   string `json:"raw"`
		Match bool   `json:"match"`
	}
	data, err := os.ReadFile("testdata/google_workspace_signin_rule.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	b, err := utils.ReadPbYaml(filepath.Join("../..", "rules/cloud/google/cloud_identity_suspicious_signins.yml"))
	if err != nil {
		t.Fatal(err)
	}
	rule := new(plugins.Rule)
	if err := protojson.Unmarshal(b, rule); err != nil {
		t.Fatal(err)
	}
	root := modelRootWithoutDynamic(t, "google/gcp.yml")
	cache := plugins.NewCELCache("google-workspace-signin-rule")
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			fixture := Fixture{Filter: "google/gcp.yml", Raw: &tc.Raw, DataType: "google", DataSource: "fabricated-collector"}
			out, issues, err := normalize(root, fixture, cache)
			if err != nil || len(issues) != 0 {
				t.Fatalf("normalize: %v / %v", err, issues)
			}
			got, err := cache.Eval(rule.Where, out)
			if err != nil || got != tc.Match {
				t.Errorf("%s = %v (%v); want %v", rule.Name, got, err, tc.Match)
			}
		})
	}
}

// modelRootWithoutDynamic returns a root holding one filter without its dynamic
// steps. The raw model does not run external enrichment such as geolocation, so
// records with a caller address would otherwise stop at that step.
func modelRootWithoutDynamic(t *testing.T, filter string) string {
	t.Helper()
	b, err := utils.ReadPbYaml(filepath.Join("../..", "filters", filter))
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
	path := filepath.Join(root, "filters", filter)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}
