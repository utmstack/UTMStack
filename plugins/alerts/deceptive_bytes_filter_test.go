package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestDeceptiveBytesCommandAndKVContract(t *testing.T) {
	path := filepath.Join("..", "..", "filters", "antivirus", "deceptive-bytes.yml")
	encoded, err := utils.ReadPbYaml(path)
	if err != nil {
		t.Fatal(err)
	}
	config := new(plugins.Config)
	if err := protojson.Unmarshal(encoded, config); err != nil {
		t.Fatal(err)
	}
	if len(config.Pipeline) != 1 {
		t.Fatalf("pipeline stages: %d, want 1", len(config.Pipeline))
	}
	commandCaptures, commandTrims, deniedResults := 0, 0, 0
	kvSources := map[string]bool{
		"log.restMessageToKv": false,
		"log.restData":        false,
		"log.pidStatusToKv":   false,
	}
	for _, step := range config.Pipeline[0].Steps {
		if grok := step.Grok; grok != nil && grok.Source == "log.restMessage" {
			for _, pattern := range grok.Patterns {
				if pattern.FieldName == "command" {
					t.Fatal("root command is dropped during Event finalization")
				}
				if pattern.FieldName == "origin.command" {
					commandCaptures++
				}
			}
		}
		if trim := step.Trim; trim != nil {
			for _, field := range trim.Fields {
				if field == "command" {
					t.Fatal("quote trim still targets the discarded command field")
				}
				if field == "origin.command" && trim.Substring == `"` {
					commandTrims++
				}
			}
		}
		if kv := step.Kv; kv != nil {
			if _, ok := kvSources[kv.Source]; !ok {
				t.Fatalf("unexpected KV source %q", kv.Source)
			}
			if kv.Where != `exists("`+kv.Source+`")` {
				t.Errorf("optional KV source %q has guard %q", kv.Source, kv.Where)
			}
			kvSources[kv.Source] = true
		}
		if add := step.Add; add != nil && add.Params["key"].GetStringValue() == "actionResult" {
			if add.Params["value"].GetStringValue() != "denied" {
				t.Errorf("blocked/prevented must map to standard denied, got %q", add.Params["value"].GetStringValue())
			}
			deniedResults++
		}
	}
	if commandCaptures != 1 || commandTrims != 2 {
		t.Errorf("command captures=%d trims=%d, want 1 and 2", commandCaptures, commandTrims)
	}
	for source, seen := range kvSources {
		if !seen {
			t.Errorf("KV source %q missing", source)
		}
	}
	if deniedResults != 1 {
		t.Errorf("actionResult mappings: %d, want 1", deniedResults)
	}

	// The reviewed SDK silently drops unknown root fields in ordinary
	// finalization. The standard Side.command survives and can be read by CEL.
	predicate := `equals("dataType", "deceptive-bytes") && equals("origin.command", "synthetic --flag")`
	cache := plugins.NewCELCache("deceptive-bytes-command-contract")
	for _, tc := range []struct {
		name, input           string
		wantStored, wantMatch bool
	}{
		{"standard command", `{"dataType":"deceptive-bytes","origin":{"command":"synthetic --flag"}}`, true, true},
		{"old root command", `{"dataType":"deceptive-bytes","command":"synthetic --flag"}`, false, false},
		{"different command", `{"dataType":"deceptive-bytes","origin":{"command":"other"}}`, true, false},
		{"different source", `{"dataType":"other","origin":{"command":"synthetic --flag"}}`, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := tc.input
			event := new(plugins.Event)
			if err := utils.StringToProtoMessage(&input, event); err != nil {
				t.Fatal(err)
			}
			output, err := utils.ProtoMessageToString(event)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(*output, `"command":`) != tc.wantStored {
				t.Fatalf("command finalization: %s", *output)
			}
			match, err := cache.Eval(predicate, *output)
			if err != nil {
				t.Fatal(err)
			}
			if match != tc.wantMatch {
				t.Fatalf("diagnostic predicate match=%t want=%t", match, tc.wantMatch)
			}
		})
	}
}

// Deceptive Bytes parses its dynamic vendor keys through KV, which calls the
// pinned SDK sanitizer before storing them under log. A rule spelling that the
// sanitizer removes cannot read the value produced by this filter.
func TestDeceptiveBytesRuleFieldSanitization(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "..", "rules", "antivirus", "deceptive-bytes", "*.yml"))
	if err != nil || len(files) != 16 {
		t.Fatalf("source rules: %d files, error %v", len(files), err)
	}
	field := regexp.MustCompile(`(?:lastEvent\.)?log\.([A-Za-z][A-Za-z0-9_]*)`)
	for _, path := range files {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range field.FindAllStringSubmatch(string(contents), -1) {
			name := match[1]
			utils.SanitizeField(&name)
			if name != match[1] {
				t.Errorf("%s reads %q; the KV producer writes log.%s", path, match[0], name)
			}
		}
	}
}

func TestDeceptiveBytesKVBooleanRuleCompatibility(t *testing.T) {
	path := filepath.Join("..", "..", "rules", "antivirus", "deceptive-bytes", "nation_state_tactic_detection.yml")
	encoded, err := utils.ReadPbYaml(path)
	if err != nil {
		t.Fatal(err)
	}
	rule := new(plugins.Rule)
	if err := protojson.Unmarshal(encoded, rule); err != nil {
		t.Fatal(err)
	}
	cache := plugins.NewCELCache("deceptive-bytes-kv-boolean")
	// Build the vendor keys the way KV stores them, with the linked SDK sanitizer.
	keys := []string{"event_type", "threat_level", "attack_sophistication", "apt_indicators"}
	for i := range keys {
		utils.SanitizeField(&keys[i])
	}
	for _, tc := range []struct {
		name, value string
		want        bool
	}{
		{"KV string true", `"true"`, true},
		{"native boolean true", `true`, true},
		{"KV string false", `"false"`, false},
		{"native boolean false", `false`, false},
		{"numeric one", `1`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := `{"dataType":"deceptive-bytes","log":{"` + keys[0] + `":"decoy_interaction","` + keys[1] + `":"critical","` + keys[2] + `":"advanced","` + keys[3] + `":` + tc.value + `}}`
			event := new(plugins.Event)
			if err := utils.StringToProtoMessage(&input, event); err != nil {
				t.Fatal(err)
			}
			output, err := utils.ProtoMessageToString(event)
			if err != nil {
				t.Fatal(err)
			}
			match, err := cache.Eval(rule.Where, *output)
			if err != nil {
				t.Fatal(err)
			}
			if match != tc.want {
				t.Fatalf("shipped rule predicate match=%t want=%t", match, tc.want)
			}
		})
	}
}

func TestDeceptiveBytesExistingRulePredicate(t *testing.T) {
	path := filepath.Join("..", "..", "rules", "antivirus", "deceptive-bytes", "living_off_the_land_detection.yml")
	encoded, err := utils.ReadPbYaml(path)
	if err != nil {
		t.Fatal(err)
	}
	rule := new(plugins.Rule)
	if err := protojson.Unmarshal(encoded, rule); err != nil {
		t.Fatal(err)
	}
	cache := plugins.NewCELCache("deceptive-bytes-existing-rule")
	eventTypeField, processNameField, targetField := "event_type", "process_name", "deceptive_target"
	utils.SanitizeField(&eventTypeField)
	utils.SanitizeField(&processNameField)
	utils.SanitizeField(&targetField)
	for _, tc := range []struct {
		name, eventType string
		want            bool
	}{
		{"matching decoy process event", "lolbin_trap", true},
		{"ordinary process event", "ordinary", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := `{"dataType":"deceptive-bytes","log":{"` + eventTypeField + `":"` + tc.eventType + `","` + processNameField + `":"cmd.exe","` + targetField + `":"decoy"}}`
			event := new(plugins.Event)
			if err := utils.StringToProtoMessage(&input, event); err != nil {
				t.Fatal(err)
			}
			output, err := utils.ProtoMessageToString(event)
			if err != nil {
				t.Fatal(err)
			}
			match, err := cache.Eval(rule.Where, *output)
			if err != nil {
				t.Fatal(err)
			}
			if match != tc.want {
				t.Fatalf("shipped rule predicate match=%t want=%t", match, tc.want)
			}
		})
	}
}

// The outcome step and the CEF source-address step, evaluated with the linked
// SDK CEL. An address is mapped only on CEF records that also carry their
// action, so it never reaches threat intelligence without its outcome.
func TestDeceptiveBytesOutcomeAndSourceAddress(t *testing.T) {
	encoded, err := utils.ReadPbYaml(filepath.Join("..", "..", "filters", "antivirus", "deceptive-bytes.yml"))
	if err != nil {
		t.Fatal(err)
	}
	config := new(plugins.Config)
	if err := protojson.Unmarshal(encoded, config); err != nil {
		t.Fatal(err)
	}
	var outcome, source string
	for _, step := range config.Pipeline[0].Steps {
		if add := step.Add; add != nil && add.Params["key"].GetStringValue() == "actionResult" {
			outcome = add.Where
		}
		if r := step.Rename; r != nil && r.To == "origin.ip" && len(r.From) == 1 && r.From[0] == "log.src" {
			source = r.Where
		}
	}
	if outcome == "" || source == "" {
		t.Fatalf("steps not found: outcome=%t source=%t", outcome != "", source != "")
	}
	cache := plugins.NewCELCache("deceptive-bytes-outcome")
	const cef = `"raw":"<14>CEF:0|Deceptive Bytes|Deceptive Bytes|6.0|1001|Malware prevented|8|"`
	for _, tc := range []struct {
		name, input          string
		wantDenied, wantSide bool
	}{
		{"cef act blocked", `{` + cef + `,"log":{"act":"blocked","src":"203.0.113.20"}}`, true, true},
		{"cef act prevented", `{` + cef + `,"log":{"act":"prevented","src":"192.0.2.30"}}`, true, true},
		{"cef act capitalized", `{` + cef + `,"log":{"act":"Prevented","src":"192.0.2.30"}}`, true, true},
		{"cef act deny", `{` + cef + `,"log":{"act":"DENY","src":"2001:db8::20"}}`, true, true},
		{"cef act detected", `{` + cef + `,"log":{"act":"detected","src":"192.0.2.30"}}`, false, true},
		{"cef act phrase", `{` + cef + `,"log":{"act":"not blocked","src":"192.0.2.30"}}`, false, true},
		{"cef source without action", `{` + cef + `,"log":{"src":"192.0.2.31"}}`, false, false},
		{"cef unusable source", `{` + cef + `,"log":{"act":"blocked","src":"0.0.0.0"}}`, true, false},
		{"cef unspecified v6 source", `{` + cef + `,"log":{"act":"blocked","src":"::"}}`, true, false},
		{"cef non-address source", `{` + cef + `,"log":{"act":"blocked","src":"dbmgmt.example.com"}}`, true, false},
		{"key=value action blocked", `{"raw":"<14>1 2026-09-27T12:00:00Z host 2 foo:1 action=blocked","log":{"action":"blocked","src":"192.0.2.33"}}`, true, false},
		{"key=value action capitalized", `{"raw":"<14>1 2026-09-27T12:00:00Z host 2 foo:1 action=Blocked","log":{"action":"Blocked"}}`, true, false},
		{"key=value other action", `{"raw":"<14>1 2026-09-27T12:00:00Z host 2 foo:1 action=file_copy","log":{"action":"file_copy"}}`, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			denied, err := cache.Eval(outcome, tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if denied != tc.wantDenied {
				t.Errorf("denied = %t, want %t", denied, tc.wantDenied)
			}
			side, err := cache.Eval(source, tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if side != tc.wantSide {
				t.Errorf("origin.ip mapped = %t, want %t", side, tc.wantSide)
			}
		})
	}
}
