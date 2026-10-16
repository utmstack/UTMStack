package runtimecompat_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/protobuf/encoding/protojson"
)

// Run this file from each plugin module so it uses that binary's resolved SDK,
// not a separate test module that could hide an incompatible runtime reader.
func TestRuntimeTenantConfig(t *testing.T) {
	rules := make([]string, 501)
	for i := range rules {
		rules[i] = fmt.Sprintf("disabled_rule_%03d", i)
	}
	rules[0], rules[1] = "apt_detection", "18446744073709551615"
	input := map[string]any{"tenants": []any{
		map[string]any{"id": "tenant-a", "name": "A", "disabledRules": rules,
			"disabledPipelines": []string{"fortinet"}, "pipelineOrder": []string{"fortinet", "windows"}},
		map[string]any{"id": "tenant-b", "name": "B", "disabledRules": []string{"brute_force"}},
	}}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	cfg := new(plugins.Config)
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, cfg); err != nil {
		t.Fatalf("runtime rejected backend tenant preferences: %v", err)
	}
	output, err := protojson.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var want, got any
	if err := json.Unmarshal(data, &want); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("runtime lost or changed tenant preferences: %s", output)
	}
}
