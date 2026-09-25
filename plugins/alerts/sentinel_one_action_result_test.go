package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"google.golang.org/protobuf/encoding/protojson"
)

// The raw fixture is also run through the isolated EventProcessor playground.
// This test checks the ordered verdict steps with the branch's actual CEL SDK.
func TestSentinelOneActionResultContract(t *testing.T) {
	var cases []struct {
		Name     string            `json:"name"`
		Result   string            `json:"result"`
		Expected map[string]string `json:"expected"`
	}
	data, err := os.ReadFile("testdata/sentinel_one_action_result.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 8 {
		t.Fatal("missing outcome regression classes")
	}
	filterYAML, err := utils.ReadPbYaml("../../filters/antivirus/sentinel-one.yml")
	if err != nil {
		t.Fatal(err)
	}
	filter := new(plugins.Config)
	if err := protojson.Unmarshal(filterYAML, filter); err != nil {
		t.Fatal(err)
	}
	cache := plugins.NewCELCache("sentinel-one-final-outcome")
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			logFields := map[string]any{}
			event := map[string]any{"log": logFields}
			for path, value := range tc.Expected {
				if strings.HasPrefix(path, "log.") {
					logFields[strings.TrimPrefix(path, "log.")] = value
				}
			}
			seenAdd := false
			for _, stage := range filter.Pipeline {
				if len(stage.DataTypes) != 1 || stage.DataTypes[0] != "antivirus-sentinel-one" {
					continue
				}
				for _, step := range stage.Steps {
					if step.Add == nil || step.Add.Params["key"].GetStringValue() != "actionResult" {
						continue
					}
					seenAdd = true
					state, err := json.Marshal(event)
					if err != nil {
						t.Fatal(err)
					}
					matched, err := cache.Eval(step.Add.Where, string(state))
					if err != nil {
						t.Fatal(err)
					}
					if matched {
						event["actionResult"] = step.Add.Params["value"].GetStringValue()
					}
				}
			}
			if !seenAdd {
				t.Fatal("missing final result mapping")
			}
			got, exists := event["actionResult"]
			if tc.Result == "" {
				if exists {
					t.Fatalf("unexpected actionResult: %v", got)
				}
			} else if !exists || got != tc.Result {
				t.Fatalf("actionResult = %v, want %q", got, tc.Result)
			}
			state, err := json.Marshal(event)
			if err != nil {
				t.Fatal(err)
			}
			for _, result := range []string{"success", "failure", "denied"} {
				matched, err := cache.Eval(`equals("actionResult","`+result+`")`, string(state))
				if err != nil || matched != (tc.Result == result) {
					t.Errorf("%s predicate = %v (%v)", result, matched, err)
				}
			}
		})
	}
}
