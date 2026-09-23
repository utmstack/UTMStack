package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/tidwall/gjson"
)

// Fabricated raw records exercise overlapping final and intermediate vendor signals.
// The parser here is the existing offline model; EventProcessor replay is separate.
func TestAzureActionResultRaw(t *testing.T) {
	content, err := os.ReadFile("testdata/azure_action_result.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name   string `json:"name"`
		Raw    string `json:"raw"`
		Result string `json:"result"`
		Rule   string `json:"rule"`
		Match  bool   `json:"match"`
	}
	if err := json.Unmarshal(content, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 20 {
		t.Fatalf("unexpected outcome coverage: %d", len(cases))
	}
	config, cache, rules := azureConfig(t), plugins.NewCELCache("azure-final-outcome"), azureRules(t)
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := azureParse(t, config, c.Raw, "synthetic-collector", cache)
			if value := gjson.Get(got, "actionResult"); value.String() != c.Result {
				t.Errorf("actionResult = %q, want %q; kind=%q statusCode=%q vendor=%q api=%q kubeKind=%q stage=%q", value.String(), c.Result, gjson.Get(got, "log.azureKind").String(), gjson.Get(got, "statusCode").String(), gjson.Get(got, "log.azureProperties.ScStatus").String(), gjson.Get(got, "log.azureKubernetes.apiVersion").String(), gjson.Get(got, "log.azureKubernetes.kind").String(), gjson.Get(got, "log.azureKubernetes.stage").String())
			}
			if c.Rule != "" {
				key := strings.TrimSuffix(filepath.Base(c.Rule), filepath.Ext(c.Rule))
				rule := rules[key]
				if rule == nil {
					t.Fatalf("unknown rule %s", c.Rule)
				}
				match, err := cache.Eval(rule.Where, got)
				if err != nil {
					t.Fatal(err)
				}
				if match != c.Match {
					t.Errorf("%s matched %v, want %v", key, match, c.Match)
				}
			}
		})
	}
}
