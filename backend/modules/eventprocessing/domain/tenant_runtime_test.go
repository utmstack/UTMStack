package domain_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
)

// Use the same YAML-to-protobuf path as the runtime, rather than only
// round-tripping the backend's own YAML structs (which accepts the old bug).
func TestTenantConfigRuntimeContract(t *testing.T) {
	rules := make([]string, 501)
	for i := range rules {
		rules[i] = fmt.Sprintf("disabled_rule_%03d", i)
	}
	rules[0] = "apt_detection"
	rules[1] = "bootkit_detection"
	rules[2] = "antivirus_service_stopped"
	rules[3] = "18446744073709551615" // An identity, never a numeric conversion.
	first := domain.TenantEntry{
		ID: "tenant-a", Name: "A", DisabledRules: rules,
		DisabledPipelines: []string{"fortinet"}, PipelineOrder: []string{"fortinet", "windows"},
	}
	second := domain.TenantEntry{ID: "tenant-b", Name: "B", DisabledRules: []string{"brute_force"}}
	cases := []struct {
		name    string
		tenants []domain.TenantEntry
	}{
		{"existing selections", []domain.TenantEntry{first, second}},
		{"one rule re-enabled", []domain.TenantEntry{
			{ID: first.ID, Name: first.Name, DisabledRules: append([]string(nil), rules[1:]...)}, second,
		}},
		{"no disabled rules", []domain.TenantEntry{{ID: first.ID, Name: first.Name}, second}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := yaml.Marshal(domain.TenantsFile{Tenants: tc.tenants})
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "tenants.yaml")
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			jsonData, err := utils.ReadPbYaml(path)
			if err != nil {
				t.Fatal(err)
			}
			cfg := new(plugins.Config)
			if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(jsonData, cfg); err != nil {
				t.Fatalf("runtime rejected backend tenant config: %v", err)
			}
			if len(cfg.Tenants) != len(tc.tenants) {
				t.Fatalf("tenants = %d", len(cfg.Tenants))
			}
			for i, want := range tc.tenants {
				got := cfg.Tenants[i]
				if got.Id != want.ID || got.Name != want.Name || !reflect.DeepEqual(got.DisabledRules, want.DisabledRules) ||
					!reflect.DeepEqual(got.DisabledPipelines, want.DisabledPipelines) || !reflect.DeepEqual(got.PipelineOrder, want.PipelineOrder) {
					t.Fatalf("tenant %s preferences changed: %v", want.ID, got)
				}
			}
		})
	}
}

// Every independently built plugin must use the same protobuf schema. A
// regression in just one module otherwise returns the runtime uint64 errors.
func TestPluginSDKVersionsMatchBackend(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate repository")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../.."))
	version := func(path string) string {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			fields := strings.Fields(line)
			if len(fields) > 0 && fields[0] == "require" {
				fields = fields[1:]
			}
			if len(fields) >= 2 && fields[0] == "github.com/threatwinds/go-sdk" {
				return fields[1]
			}
		}
		return ""
	}
	want := version(filepath.Join(root, "backend/go.mod"))
	if want == "" {
		t.Fatal("backend SDK version missing")
	}
	paths, err := filepath.Glob(filepath.Join(root, "plugins/*/go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, path := range paths {
		if got := version(path); got != "" {
			checked++
			if got != want {
				t.Errorf("%s uses SDK %s; backend uses %s", path, got, want)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no plugin SDK dependencies checked")
	}
}
