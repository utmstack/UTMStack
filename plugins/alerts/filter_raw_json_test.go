package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
)

// fixtureJSON models only documented JSON-to-log extraction. Key sanitization
// uses the pinned SDK. This is not the closed pipeline executor; ambiguous keys
// are rejected instead of guessing its overwrite/flattening behavior.
func fixtureJSON(raw string) (map[string]any, error) {
	var parsed map[string]any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}
	if parsed == nil {
		return nil, fmt.Errorf("raw JSON model requires an object")
	}
	value, err := sanitizeFixtureJSON(parsed)
	if err != nil {
		return nil, err
	}
	return value.(map[string]any), nil
}

func sanitizeFixtureJSON(value any) (any, error) {
	switch input := value.(type) {
	case map[string]any:
		output := make(map[string]any, len(input))
		for key, child := range input {
			utils.SanitizeField(&key)
			if key == "" || strings.Contains(key, ".") {
				return nil, fmt.Errorf("raw JSON model cannot establish empty/dotted key semantics: %q", key)
			}
			if _, exists := output[key]; exists {
				return nil, fmt.Errorf("raw JSON model has sanitization collision: %q", key)
			}
			clean, err := sanitizeFixtureJSON(child)
			if err != nil {
				return nil, err
			}
			output[key] = clean
		}
		return output, nil
	case []any:
		output := make([]any, len(input))
		for i, child := range input {
			clean, err := sanitizeFixtureJSON(child)
			if err != nil {
				return nil, err
			}
			output[i] = clean
		}
		return output, nil
	default:
		return value, nil
	}
}

func TestRawJSONFixtureNormalization(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "filters"), 0700); err != nil {
		t.Fatal(err)
	}
	filter := `pipeline:
  - dataTypes: [fixture-json]
    steps:
      - json: {source: raw}
      - rename:
          from: [log.SourceIP]
          to: origin.ip
          where: inCIDR("log.SourceIP", "0.0.0.0/0") || inCIDR("log.SourceIP", "::/0")
      - rename: {from: [log.EventID], to: action}
      - cast: {fields: [action], to: string}
`
	if err := os.WriteFile(filepath.Join(root, "filters", "fixture.yml"), []byte(filter), 0600); err != nil {
		t.Fatal(err)
	}
	cache := plugins.NewCELCache("raw-json-model-test")
	for _, ip := range []string{"203.0.113.7", "-"} {
		raw := fmt.Sprintf(`{"Source_IP":%q,"Event-ID":4769,"Nested_Items":[{"Child-Name":"preserved"}]}`, ip)
		fixture := Fixture{Filter: "fixture.yml", Raw: &raw, DataType: "fixture-json", DataSource: "synthetic-host"}
		out, issues, err := normalize(root, fixture, cache)
		if err != nil || len(issues) != 0 {
			t.Fatalf("normalize: %v / %v", err, issues)
		}
		if gjson.Get(out, "action").String() != "4769" || gjson.Get(out, "log.NestedItems.0.ChildName").String() != "preserved" {
			t.Fatalf("JSON sanitization/normalization failed: %s", out)
		}
		if gjson.Get(out, "origin.ip").Exists() != (ip != "-") {
			t.Fatalf("IP promotion guard failed: %s", out)
		}
		if ip == "-" && gjson.Get(out, "log.SourceIP").String() != "-" {
			t.Fatalf("invalid vendor IP should remain at its source: %s", out)
		}
		fixture.Input = map[string]any{"log": map[string]any{}}
		if _, _, err := normalize(root, fixture, cache); err == nil {
			t.Fatal("raw and synthetic input must be mutually exclusive")
		}
	}
	unsupported := strings.Replace(filter, "- json: {source: raw}", "- kv: {source: raw}", 1)
	if err := os.WriteFile(filepath.Join(root, "filters", "fixture.yml"), []byte(unsupported), 0600); err != nil {
		t.Fatal(err)
	}
	raw := `{"SourceIP":"203.0.113.7"}`
	fixture := Fixture{Filter: "fixture.yml", Raw: &raw, DataType: "fixture-json", DataSource: "synthetic-host"}
	if _, _, err := normalize(root, fixture, cache); err == nil || !strings.Contains(err.Error(), "does not support step kv") {
		t.Fatalf("raw fixture must reject unsupported extraction: %v", err)
	}
}

func TestRawJSONFixtureRejectsAmbiguousInput(t *testing.T) {
	for _, raw := range []string{`null`, `[]`, `{`, `{"key-name":1,"key_name":2}`, `{"source.ip":"192.0.2.1"}`, `{"!!!":1}`} {
		if _, err := fixtureJSON(raw); err == nil {
			t.Errorf("expected model boundary error for %s", raw)
		}
	}
}

func TestRawFixtureRejectsUnsupportedAddAndMultilineGrok(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "filters"), 0700); err != nil {
		t.Fatal(err)
	}
	cache := plugins.NewCELCache("raw-json-boundaries-test")
	for _, function := range []string{"string", "uuid", "time", "join", ""} {
		filter := fmt.Sprintf(`pipeline:
  - dataTypes: [fixture-json]
    steps:
      - add:
          function: %q
          params: {key: log.literal, value: 17}
`, function)
		if err := os.WriteFile(filepath.Join(root, "filters", "fixture.yml"), []byte(filter), 0600); err != nil {
			t.Fatal(err)
		}
		raw := `{}`
		fixture := Fixture{Filter: "fixture.yml", Raw: &raw, DataType: "fixture-json", DataSource: "synthetic-host"}
		out, _, err := normalize(root, fixture, cache)
		if function == "string" {
			if err != nil || gjson.Get(out, "log.literal").Type != gjson.String || gjson.Get(out, "log.literal").String() != "17" {
				t.Fatalf("add string must use SDK string conversion: %s / %v", out, err)
			}
		} else if err == nil || !strings.Contains(err.Error(), "does not support add function") {
			t.Fatalf("unsupported add %q was silently accepted: %v", function, err)
		}
	}
	for _, pattern := range []string{"{{.greedy}}", "(.*)", "(?s:.*)"} {
		filter := fmt.Sprintf(`pipeline:
  - dataTypes: [fixture-json]
    steps:
      - grok:
          source: raw
          patterns: [{fieldName: log.message, pattern: %q}]
`, pattern)
		if err := os.WriteFile(filepath.Join(root, "filters", "fixture.yml"), []byte(filter), 0600); err != nil {
			t.Fatal(err)
		}
		for _, raw := range []string{"first line\nsecond line", "first line\r\nsecond line"} {
			fixture := Fixture{Filter: "fixture.yml", Raw: &raw, DataType: "fixture-json", DataSource: "synthetic-host"}
			out, _, err := normalize(root, fixture, cache)
			if pattern == "(?s:.*)" {
				if err != nil || gjson.Get(out, "log.message").String() != raw {
					t.Fatalf("explicit dot-all copy must preserve the complete message: %s / %v", out, err)
				}
			} else if err == nil || !strings.Contains(err.Error(), "multiline copy grok") {
				t.Fatalf("multiline %s must not be treated as whole-field copy: %v", pattern, err)
			}
		}
	}
}
