package usecase

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

// Existing on-disk playbooks use the old `commands:` chain shape. The reader
// has to keep parsing them, upgrading to a linear DAG at load time — otherwise
// every deployment loses its shipped flows the day this ships.
func TestReadFlowFile_LegacyChainUpgrades(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "legacy.yaml")
	body := []byte(`
- name: Restart wazuh
  agentPlatform: linux
  commands:
  - systemctl restart wazuh
  - echo done
  shell: bash
`)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
	flow, err := readFlowFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(flow.Roots) != 1 || flow.Roots[0] != "step_0" {
		t.Fatalf("roots = %v, want [step_0]", flow.Roots)
	}
	if flow.Nodes["step_0"].Kind != domain.NodeKindExecutor {
		t.Errorf("step_0 kind = %q", flow.Nodes["step_0"].Kind)
	}
	if got := flow.Nodes["step_0"].Command; got != "systemctl restart wazuh" {
		t.Errorf("step_0 command = %q", got)
	}
	if got := flow.Nodes["step_0"].Shell; got != "bash" {
		t.Errorf("step_0 shell = %q, want bash", got)
	}
	// A legacy step with no `condition:` links via both edges — the old
	// `;` operator ran the next command regardless of outcome.
	step0 := flow.Nodes["step_0"]
	if len(step0.OnSuccess) != 1 || step0.OnSuccess[0] != "step_1" {
		t.Errorf("step_0.OnSuccess = %v", step0.OnSuccess)
	}
	if len(step0.OnError) != 1 || step0.OnError[0] != "step_1" {
		t.Errorf("step_0.OnError = %v", step0.OnError)
	}
}

func TestReadFlowFile_DAGShapePreserved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dag.yaml")
	body := []byte(`
- name: Fan-in
  agentPlatform: linux
  roots: [geoip, extract]
  nodes:
    geoip:
      kind: enrichment
      executor: http
      params: {"url":"https://geo"}
      onSuccess: [notify]
    extract:
      kind: enrichment
      executor: select
      params: {"fields":{"user":"alert.user.name"}}
      onSuccess: [notify]
    notify:
      kind: executor
      executor: http
      params: {"url":"https://slack"}
`)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
	flow, err := readFlowFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(flow.Roots) != 2 {
		t.Errorf("roots = %v", flow.Roots)
	}
	if flow.Nodes["geoip"].Kind != domain.NodeKindEnrichment {
		t.Errorf("geoip kind = %q", flow.Nodes["geoip"].Kind)
	}
	// notify must be the AND-join sink — both roots reference it.
	if got := flow.IncomingCounts()["notify"]; got != 2 {
		t.Errorf("incoming[notify] = %d, want 2", got)
	}
}

func TestFlow_ResolvedMaxDepthDefaults(t *testing.T) {
	if got := (domain.Flow{}).ResolvedMaxDepth(); got != domain.DefaultMaxDepth {
		t.Errorf("default = %d, want %d", got, domain.DefaultMaxDepth)
	}
	if got := (domain.Flow{MaxDepth: 7}).ResolvedMaxDepth(); got != 7 {
		t.Errorf("explicit = %d, want 7", got)
	}
}
