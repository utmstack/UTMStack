package main

import (
	"os"
	"path/filepath"
	"testing"
)

// Byte-for-byte what the backend's ConfigStore writes. If the two sides drift,
// this plugin silently reads zero groups and stops collecting without an error,
// so the fixture is kept literal rather than built from a helper here.
const backendWritten = `plugins:
    azure:
        tenants:
            - id: 8f1c1b8e-0000-4000-8000-00000000000a
              groups:
                - name: production
                  description: east us
                  config:
                    consumerGroup: $Default
                    eventHubConnection: conn-A
                - name: dev
                  config:
                    eventHubConnection: conn-A-dev
            - id: 8f1c1b8e-0000-4000-8000-00000000000b
              groups:
                - name: production
                  config:
                    eventHubConnection: conn-B
`

func writeFixture(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "system_plugins_azure.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestThePluginReadsWhatTheBackendWrites(t *testing.T) {
	sec := readConfig(writeFixture(t, backendWritten), "")
	if sec == nil {
		t.Fatal("readConfig returned nil for a well-formed file")
	}
	if !sec.ModuleActive {
		t.Error("module reads as inactive with three groups configured")
	}
	if n := len(sec.ModuleGroups); n != 3 {
		t.Fatalf("loaded %d groups, want 3 — the plugin cannot see the backend's format", n)
	}
}

// Two tenants may each call their connector "production". Collapsing them would
// run one customer's credentials under the other's name, or stop their collector.
func TestSameGroupNameInTwoTenantsStaysApart(t *testing.T) {
	sec := readConfig(writeFixture(t, backendWritten), "")

	keys := map[string]string{}
	for _, g := range sec.ModuleGroups {
		if prev, dup := keys[g.Key()]; dup {
			t.Fatalf("two groups share the key %q (%s and %s)", g.Key(), prev, g.GroupName)
		}
		keys[g.Key()] = g.GroupName
	}
	if len(keys) != 3 {
		t.Fatalf("%d distinct keys for 3 groups", len(keys))
	}

	var a, b string
	for _, g := range sec.ModuleGroups {
		if g.GroupName != "production" {
			continue
		}
		for _, c := range g.ModuleGroupConfigurations {
			if c.ConfKey != "eventHubConnection" {
				continue
			}
			if g.TenantId == "8f1c1b8e-0000-4000-8000-00000000000a" {
				a = c.ConfValue
			} else {
				b = c.ConfValue
			}
		}
	}
	if a != "conn-A" || b != "conn-B" {
		t.Errorf("credentials crossed tenants: A=%q B=%q", a, b)
	}
}

// Every group carries the tenant that owns it, because the events it produces
// are filed under that tenant and nothing downstream can recover it otherwise.
func TestEveryGroupCarriesItsTenant(t *testing.T) {
	sec := readConfig(writeFixture(t, backendWritten), "")
	for _, g := range sec.ModuleGroups {
		if g.TenantId == "" {
			t.Errorf("group %q has no tenant", g.GroupName)
		}
	}
}

// An empty tenant section must not make the module read as configured — that is
// what decides whether any collector starts at all.
func TestATenantWithNoGroupsIsNotActive(t *testing.T) {
	sec := readConfig(writeFixture(t, `plugins:
    azure:
        tenants:
            - id: 8f1c1b8e-0000-4000-8000-00000000000a
              groups: []
`), "")
	if sec.ModuleActive {
		t.Error("module reads as active with no groups configured")
	}
	if len(sec.ModuleGroups) != 0 {
		t.Errorf("loaded %d groups from an empty section", len(sec.ModuleGroups))
	}
}

// The file is removed when nothing is configured; that has to stop the work,
// not crash the plugin.
func TestAMissingFileReportsInactive(t *testing.T) {
	sec := readConfig(filepath.Join(t.TempDir(), "gone.yaml"), "")
	if sec == nil {
		t.Fatal("readConfig returned nil instead of an inactive section")
	}
	if sec.ModuleActive || len(sec.ModuleGroups) != 0 {
		t.Error("a missing file did not read as inactive")
	}
}
