package repository

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
)

const (
	tenantA = "8f1c1b8e-0000-4000-8000-00000000000a"
	tenantB = "8f1c1b8e-0000-4000-8000-00000000000b"
)

func readTenants(t *testing.T, dir string) domain.TenantsFile {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, TenantFileName))
	if err != nil {
		t.Fatalf("read tenants.yaml: %v", err)
	}
	var f domain.TenantsFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		t.Fatalf("parse tenants.yaml: %v", err)
	}
	return f
}

func entryOf(f domain.TenantsFile, id string) (domain.TenantEntry, bool) {
	for _, e := range f.Tenants {
		if e.ID == id {
			return e, true
		}
	}
	return domain.TenantEntry{}, false
}

// What a customer switches off is its own answer. Writing it across the file
// was stopping every other tenant's rule from firing.
func TestDisablingARuleTouchesOnlyThatTenant(t *testing.T) {
	dir := t.TempDir()
	w := NewEngineConfig(dir)

	if err := w.SetRuleDisabled(tenantA, "hidden_user_creation", true); err != nil {
		t.Fatalf("SetRuleDisabled: %v", err)
	}
	if err := w.SetRuleDisabled(tenantB, "brute_force", true); err != nil {
		t.Fatalf("SetRuleDisabled: %v", err)
	}

	a, ok := entryOf(readTenants(t, dir), tenantA)
	if !ok {
		t.Fatal("tenant A missing from the file")
	}
	if len(a.DisabledRules) != 1 || a.DisabledRules[0] != "hidden_user_creation" {
		t.Errorf("tenant A disabled rules = %v, want only its own", a.DisabledRules)
	}

	b, _ := entryOf(readTenants(t, dir), tenantB)
	if len(b.DisabledRules) != 1 || b.DisabledRules[0] != "brute_force" {
		t.Errorf("tenant B disabled rules = %v, want only its own", b.DisabledRules)
	}
}

func TestDisablingAPipelineTouchesOnlyThatTenant(t *testing.T) {
	dir := t.TempDir()
	w := NewEngineConfig(dir)

	if err := w.SetPipelineDisabled(tenantA, "kaspersky", true); err != nil {
		t.Fatal(err)
	}

	if got := w.DisabledPipelineSet(tenantA); !got["kaspersky"] {
		t.Error("tenant A's own disabled pipeline is missing")
	}
	if got := w.DisabledPipelineSet(tenantB); got["kaspersky"] {
		t.Error("tenant B inherited another tenant's disabled pipeline")
	}
}

// The order is the whole point of this change: each tenant keeps its own.
func TestEachTenantKeepsItsOwnPipelineOrder(t *testing.T) {
	dir := t.TempDir()
	w := NewEngineConfig(dir)

	if err := w.SetPipelineOrder(tenantA, []string{"azure-eventhub", "aws"}); err != nil {
		t.Fatal(err)
	}
	if err := w.SetPipelineOrder(tenantB, []string{"aws"}); err != nil {
		t.Fatal(err)
	}

	if got := w.PipelineOrder(tenantA); len(got) != 2 || got[0] != "azure-eventhub" {
		t.Errorf("tenant A order = %v", got)
	}
	if got := w.PipelineOrder(tenantB); len(got) != 1 || got[0] != "aws" {
		t.Errorf("tenant B order = %v", got)
	}
}

// Clearing the preference sends the tenant back to the order the files declare,
// which is what an empty list has to mean.
func TestAnEmptyOrderClearsThePreference(t *testing.T) {
	dir := t.TempDir()
	w := NewEngineConfig(dir)

	if err := w.SetPipelineOrder(tenantA, []string{"aws"}); err != nil {
		t.Fatal(err)
	}
	if err := w.SetPipelineOrder(tenantA, nil); err != nil {
		t.Fatal(err)
	}

	if got := w.PipelineOrder(tenantA); len(got) != 0 {
		t.Errorf("order = %v, want cleared", got)
	}
}

// The asset projection runs on its own schedule and must not wipe what the
// tenants chose while it refreshes their hosts.
func TestProjectingAssetsPreservesTenantChoices(t *testing.T) {
	dir := t.TempDir()
	w := NewEngineConfig(dir)

	if err := w.SetPipelineOrder(tenantA, []string{"aws"}); err != nil {
		t.Fatal(err)
	}
	if err := w.SetRuleDisabled(tenantA, "some_rule", true); err != nil {
		t.Fatal(err)
	}

	if err := w.WriteTenants([]domain.TenantAssets{
		{ID: tenantA, Name: "Acme", Assets: []domain.Asset{{Name: "web-01"}}},
	}); err != nil {
		t.Fatalf("WriteTenants: %v", err)
	}

	a, ok := entryOf(readTenants(t, dir), tenantA)
	if !ok {
		t.Fatal("tenant A vanished from the file")
	}
	if len(a.Assets) != 1 {
		t.Errorf("assets = %v, want the projection applied", a.Assets)
	}
	if len(a.PipelineOrder) != 1 || a.PipelineOrder[0] != "aws" {
		t.Errorf("pipeline order = %v, want it preserved", a.PipelineOrder)
	}
	if len(a.DisabledRules) != 1 {
		t.Errorf("disabled rules = %v, want them preserved", a.DisabledRules)
	}
}

// The engine parses this file with DiscardUnknown, so a key it does not expect
// is ignored in silence and the tenant quietly gets the shipped order.
func TestTheOrderIsWrittenUnderTheKeyTheEngineReads(t *testing.T) {
	dir := t.TempDir()
	w := NewEngineConfig(dir)

	if err := w.SetPipelineOrder(tenantA, []string{"aws"}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, TenantFileName))
	if err != nil {
		t.Fatal(err)
	}
	var raw struct {
		Tenants []map[string]any `yaml:"tenants"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, e := range raw.Tenants {
		if e["id"] != tenantA {
			continue
		}
		if _, ok := e["pipelineOrder"]; !ok {
			t.Fatalf("key is %v, want pipelineOrder — the engine will ignore anything else", keysOf(e))
		}
		return
	}
	t.Fatal("tenant A missing from the file")
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
