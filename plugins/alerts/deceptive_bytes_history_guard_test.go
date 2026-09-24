package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

// A history search whose placeholder is missing returns an error, and five errors switch the
// rule off with a Circuit Breaker alert. Each rule must only match when its placeholders exist.
func TestDeceptiveBytesHistoryPlaceholdersGuarded(t *testing.T) {
	cache := plugins.NewCELCache("deceptive-bytes-history-guard")
	cases := []struct {
		rule  string
		event string
	}{
		{"data_theft_attempt_indicators", `{"eventtype":"decoy_accessed","action":"file_copy","decoysensitivity":"high","decoyfile":"f"}`},
		{"advanced_threat_tactic_identification", `{"eventType":"advanced_threat_detected","threatLevel":"critical","tacticName":"execution","deceptionTriggered":"true","behaviorScore":95}`},
		{"zero_day_behavior_patterns", `{"eventType":"zero_day_suspect","threatSignature":"unknown","deceptionEnvironment":"true","memoryAnomalyScore":95,"knownMalwareFamily":"","exploitTechnique":"t","processName":"p.exe"}`},
	}
	for _, tc := range cases {
		b, err := utils.ReadPbYaml(filepath.Join("../..", "rules/antivirus/deceptive-bytes", tc.rule+".yml"))
		if err != nil {
			t.Fatal(err)
		}
		rule := new(plugins.Rule)
		if err := protojson.Unmarshal(b, rule); err != nil {
			t.Fatal(err)
		}
		rule.Normalize()
		without := `{"dataType":"deceptive-bytes","log":` + tc.event + `}`
		with := `{"dataType":"deceptive-bytes","origin":{"ip":"192.0.2.10"},"log":` + tc.event + `}`
		t.Run(tc.rule, func(t *testing.T) {
			if got, err := cache.Eval(rule.Where, without); err != nil || got {
				t.Errorf("matched without origin.ip: %v (%v)", got, err)
			}
			got, err := cache.Eval(rule.Where, with)
			if err != nil || !got {
				t.Fatalf("did not match with origin.ip: %v (%v)", got, err)
			}
			for _, block := range rule.Correlation {
				for _, expr := range block.With {
					value := expr.Value.GetStringValue()
					if strings.HasPrefix(value, "{{.") && strings.HasSuffix(value, "}}") {
						field := strings.TrimSuffix(strings.TrimPrefix(value, "{{."), "}}")
						if !gjson.Get(with, field).Exists() {
							t.Errorf("placeholder %s unresolved on a matching event", field)
						}
					}
				}
			}
		})
	}
}
