package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

// deceptiveBytesEvent builds an event from a log object plus dotted fields.
func deceptiveBytesEvent(t *testing.T, logFields map[string]any, extra map[string]string) string {
	t.Helper()
	logCopy := map[string]any{}
	for k, v := range logFields {
		logCopy[k] = v
	}
	event := map[string]any{"dataType": "deceptive-bytes", "log": logCopy}
	for path, value := range extra {
		parts := strings.Split(path, ".")
		node := event
		for _, part := range parts[:len(parts)-1] {
			next, ok := node[part].(map[string]any)
			if !ok {
				next = map[string]any{}
				node[part] = next
			}
			node = next
		}
		node[parts[len(parts)-1]] = value
	}
	b, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// A history search whose placeholder is missing returns an error, and five errors switch the
// rule off with a Circuit Breaker alert. Each rule must only match when its placeholders exist.
func TestDeceptiveBytesHistoryPlaceholdersGuarded(t *testing.T) {
	cache := plugins.NewCELCache("deceptive-bytes-history-guard")
	cases := []struct {
		rule   string
		log    map[string]any    // matches the condition apart from the placeholder fields
		needed map[string]string // the fields the history placeholders read
	}{
		{"data_theft_attempt_indicators",
			map[string]any{"event_type": "decoy_accessed", "action": "file_copy", "decoy_sensitivity": "high", "decoy_file": "f"},
			map[string]string{"origin.ip": "192.0.2.10"}},
		{"advanced_threat_tactic_identification",
			map[string]any{"eventType": "advanced_threat_detected", "threatLevel": "critical", "tacticName": "execution", "deceptionTriggered": "true", "behaviorScore": 95},
			map[string]string{"origin.ip": "192.0.2.10"}},
		{"zero_day_behavior_patterns",
			map[string]any{"eventType": "zero_day_suspect", "threatSignature": "unknown", "deceptionEnvironment": "true", "memoryAnomalyScore": 95, "knownMalwareFamily": "", "exploitTechnique": "t", "processName": "p.exe"},
			map[string]string{"origin.ip": "192.0.2.10"}},
		{"ransomware_behavior_patterns",
			map[string]any{"event_type": "ransomware_behavior", "behavior_pattern": "mass_encryption"},
			map[string]string{"log.process": "example.exe", "log.source_ip": "192.0.2.10"}},
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
		t.Run(tc.rule, func(t *testing.T) {
			// Without the placeholder fields, and without any one of them, the rule must not match.
			if got, err := cache.Eval(rule.Where, deceptiveBytesEvent(t, tc.log, nil)); err != nil || got {
				t.Errorf("matched without %v: %v (%v)", tc.needed, got, err)
			}
			for missing := range tc.needed {
				partial := map[string]string{}
				for path, value := range tc.needed {
					if path != missing {
						partial[path] = value
					}
				}
				if got, err := cache.Eval(rule.Where, deceptiveBytesEvent(t, tc.log, partial)); err != nil || got {
					t.Errorf("matched without %s: %v (%v)", missing, got, err)
				}
			}
			with := deceptiveBytesEvent(t, tc.log, tc.needed)
			got, err := cache.Eval(rule.Where, with)
			if err != nil || !got {
				t.Fatalf("did not match with %v: %v (%v)", tc.needed, got, err)
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
