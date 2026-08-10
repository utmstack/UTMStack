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

func writeEnabled(t *testing.T, dir string, rel ...string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := "enabled:\n"
	for _, r := range rel {
		body += "    - " + r + "\n"
	}
	if err := os.WriteFile(filepath.Join(dir, enabledFileName), []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func dirs(t *testing.T) (systemDir, userRoot string) {
	t.Helper()
	root := t.TempDir()
	return filepath.Join(root, "shipped"), filepath.Join(root, "user")
}

// Nothing runs until a tenant asks for it. These flows shut machines down and
// cut them off the network; installing the product must not be how that starts.
func TestNothingRunsUntilATenantEnablesIt(t *testing.T) {
	systemDir, userRoot := dirs(t)
	writeFlow(t, filepath.Join(systemDir, "shipped.yaml"))
	writeFlow(t, filepath.Join(userRoot, tenantA, "mine.yaml"))

	loadFlows(systemDir, userRoot)

	if n := len(activeFlowsFor(tenantA)); n != 0 {
		t.Errorf("%d flows are running with nothing switched on", n)
	}
}

// A response flow decides what runs on a tenant's hosts. Evaluating one
// tenant's flow against another's alert would run commands on the wrong
// machines, so the enabled list is read per tenant and never shared.
func TestAFlowIsOnlyEvaluatedForItsOwnTenant(t *testing.T) {
	systemDir, userRoot := dirs(t)
	writeFlow(t, filepath.Join(systemDir, "shipped.yaml"))
	writeFlow(t, filepath.Join(userRoot, tenantA, "only-a.yaml"))
	writeFlow(t, filepath.Join(userRoot, tenantB, "only-b.yaml"))
	writeEnabled(t, filepath.Join(userRoot, tenantA), "shipped.yaml", "only-a.yaml")
	writeEnabled(t, filepath.Join(userRoot, tenantB), "shipped.yaml", "only-b.yaml")

	loadFlows(systemDir, userRoot)

	got := map[string]bool{}
	for _, f := range activeFlowsFor(tenantA) {
		got[f.RelPath] = true
	}
	if !got["shipped.yaml"] {
		t.Error("the shipped flow did not reach the tenant that switched it on")
	}
	if !got["only-a.yaml"] {
		t.Error("the tenant's own flow is missing")
	}
	if got["only-b.yaml"] {
		t.Error("another tenant's flow was evaluated")
	}
}

// The shipped file is one copy shared by every tenant, so switching it on is
// recorded per tenant and must not reach anyone else.
func TestEnablingAShippedFlowIsPerTenant(t *testing.T) {
	systemDir, userRoot := dirs(t)
	writeFlow(t, filepath.Join(systemDir, "shipped.yaml"))
	writeEnabled(t, filepath.Join(userRoot, tenantA), "shipped.yaml")

	loadFlows(systemDir, userRoot)

	if n := len(activeFlowsFor(tenantA)); n != 1 {
		t.Errorf("the tenant that switched it on runs %d flows, want 1", n)
	}
	if n := len(activeFlowsFor(tenantB)); n != 0 {
		t.Errorf("a tenant who never asked runs %d flows", n)
	}
}

// A tenant writing their own file at the shipped path replaces it for them —
// running both would execute the commands twice on the same alert.
func TestATenantsOwnFileReplacesTheShippedOne(t *testing.T) {
	systemDir, userRoot := dirs(t)
	writeFlow(t, filepath.Join(systemDir, "shipped.yaml"))
	writeFlow(t, filepath.Join(userRoot, tenantA, "shipped.yaml"))
	writeEnabled(t, filepath.Join(userRoot, tenantA), "shipped.yaml")

	loadFlows(systemDir, userRoot)

	n := 0
	for _, f := range activeFlowsFor(tenantA) {
		if f.RelPath == "shipped.yaml" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("shipped.yaml resolved to %d flows, want the tenant's one", n)
	}
}

// The enabled list lives in the same directory as the tenant's flows and must
// not be mistaken for one.
func TestTheEnabledFileIsNotLoadedAsAFlow(t *testing.T) {
	systemDir, userRoot := dirs(t)
	writeEnabled(t, filepath.Join(userRoot, tenantA), "nothing.yaml")

	loadFlows(systemDir, userRoot)

	if n := len(activeFlowsFor(tenantA)); n != 0 {
		t.Errorf("loaded %d flows, want none — the enabled file was read as a flow", n)
	}
}

// matches() answers true for an empty condition set, so a flow with none would
// run its commands against every alert. It must not be loaded at all.
func TestAFlowWithNoConditionsIsNotLoaded(t *testing.T) {
	systemDir, userRoot := dirs(t)
	path := filepath.Join(systemDir, "no-conditions.yaml")
	if err := os.MkdirAll(systemDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("- name: x\n  conditions: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeEnabled(t, filepath.Join(userRoot, tenantA), "no-conditions.yaml")

	loadFlows(systemDir, userRoot)

	if n := len(activeFlowsFor(tenantA)); n != 0 {
		t.Errorf("a flow with no conditions is live: %d active — it matches every alert", n)
	}
}
