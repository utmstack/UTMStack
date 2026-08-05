package main

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	tenantA = "8f1c1b8e-0000-4000-8000-00000000000a"
	tenantB = "8f1c1b8e-0000-4000-8000-00000000000b"
)

func writeFlow(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := "- name: x\n  conditions:\n    - operator: Is\n      field: name\n      value: y\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// A response flow decides what runs on a tenant's hosts. Evaluating one tenant's
// flow against another's alert would run their commands on the wrong machines.
// The flows the product ships are a copy inside each tenant, so even those are
// not shared: a tenant switches one off without touching anybody else's.
func TestAFlowIsOnlyEvaluatedForItsOwnTenant(t *testing.T) {
	root := t.TempDir()
	writeFlow(t, filepath.Join(root, "user", tenantA, "system", "shipped.yaml"))
	writeFlow(t, filepath.Join(root, "user", tenantA, "only-a.yaml"))
	writeFlow(t, filepath.Join(root, "user", tenantB, "system", "shipped.yaml"))
	writeFlow(t, filepath.Join(root, "user", tenantB, "only-b.yaml"))

	loadFlows(root)

	got := map[string]bool{}
	for _, f := range activeFlowsFor(tenantA) {
		got[f.RelPath] = true
	}

	if !got["shipped.yaml"] {
		t.Error("the tenant's copy of the shipped flow was not loaded")
	}
	if !got["only-a.yaml"] {
		t.Error("the tenant's own flow is missing")
	}
	if got["only-b.yaml"] {
		t.Error("another tenant's flow was evaluated")
	}
}

// A tenant that switched a shipped flow off must not have it run, while the
// tenant that left it on still does.
func TestADisabledShippedFlowIsOffForThatTenantOnly(t *testing.T) {
	root := t.TempDir()
	writeFlow(t, filepath.Join(root, "user", tenantA, "system", "shipped.yaml"))
	writeFlow(t, filepath.Join(root, "user", tenantB, "system", "shipped.yaml.disabled"))

	loadFlows(root)

	if len(activeFlowsFor(tenantA)) != 1 {
		t.Error("the tenant that left it on lost the flow")
	}
	if n := len(activeFlowsFor(tenantB)); n != 0 {
		t.Errorf("the tenant that switched it off still runs %d flows", n)
	}
}

// A file left directly in user/ names no tenant, so there is no one to run it
// for. Loading it as everyone's would be the bug this exists to prevent.
func TestAFlowWithNoTenantIsIgnored(t *testing.T) {
	root := t.TempDir()
	writeFlow(t, filepath.Join(root, "user", "orphan.yaml"))

	loadFlows(root)

	if len(activeFlowsFor(tenantA)) != 0 {
		t.Errorf("an untenanted user flow was loaded: %+v", activeFlowsFor(tenantA))
	}
}
