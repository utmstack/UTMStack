package collector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/collectors/forwarder/config"
)

// ---- T2: ChangeIntegrationStatus seeds empty ports and fails without one (S1, S2) ----
//
// User types must be resolvable via config.ResolveDataType, which requires a
// catalog entry; the built-in netflow type (ProtoPorts: UDP 2055, no TCP)
// supplies a known type that has no default for one proto, which makes the
// "no port anywhere" error reachable.

const testCatalog = `{"user_datatypes":{"firewall-barracuda":{"kind":"syslog","udp":"7100","tcp":"7101"}}}`

func TestChangeIntegrationStatusSeedsPortWhenStoredEmpty(t *testing.T) {
	ensureTestLogger()
	tmp := t.TempDir()
	sandboxPointers(t, tmp)
	writeCatalog(t, testCatalog)
	wipeConfig(t)

	port, err := ChangeIntegrationStatus("firewall-barracuda", "udp", true, "7100")
	if err != nil {
		t.Fatalf("ChangeIntegrationStatus enable with seed: %v", err)
	}
	if port != "7100" {
		t.Errorf("returned port = %q, want 7100", port)
	}
	cnf := readStored(t)
	udp := cnf.Integrations["firewall-barracuda"].UDP
	if !udp.IsListen {
		t.Error("udp not marked listening after enable")
	}
	if udp.Port != "7100" {
		t.Errorf("persisted udp port = %q, want seeded 7100", udp.Port)
	}
}

func TestChangeIntegrationStatusFallsBackToCatalogDefault(t *testing.T) {
	ensureTestLogger()
	tmp := t.TempDir()
	sandboxPointers(t, tmp)
	writeCatalog(t, testCatalog)
	wipeConfig(t)

	port, err := ChangeIntegrationStatus("firewall-barracuda", "udp", true, "")
	if err != nil {
		t.Fatalf("ChangeIntegrationStatus enable with catalog default: %v", err)
	}
	if port != "7100" {
		t.Errorf("returned port = %q, want catalog default 7100", port)
	}
	cnf := readStored(t)
	udp := cnf.Integrations["firewall-barracuda"].UDP
	if !udp.IsListen {
		t.Error("udp not marked listening after enable")
	}
	if udp.Port != "7100" {
		t.Errorf("persisted udp port = %q, want catalog default 7100", udp.Port)
	}
}

func TestChangeIntegrationStatusErrorsWithoutAnyPort(t *testing.T) {
	ensureTestLogger()
	tmp := t.TempDir()
	sandboxPointers(t, tmp)
	// netflow is a built-in with UDP 2055 but NO TCP default; no catalog, no
	// seed, stored tcp port empty -> the only way to reach "no port anywhere".
	if err := os.WriteFile(config.CollectorFileName, []byte(`{
		"integrations": {
			"netflow": {"tcp_port": {"enabled": false, "value": ""}}
		}
	}`), 0644); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(config.CollectorFileName)
	if err != nil {
		t.Fatal(err)
	}

	port, err := ChangeIntegrationStatus("netflow", "tcp", true, "")
	if err == nil {
		t.Fatalf("ChangeIntegrationStatus enable with no port anywhere: want error, got port %q", port)
	}
	after, err := os.ReadFile(config.CollectorFileName)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("config file was modified even though enable failed with no port")
	}
	cnf := readStored(t)
	if cnf.Integrations["netflow"].TCP.IsListen {
		t.Error("IsListen must stay false when enable fails with no port")
	}
	if cnf.Integrations["netflow"].TCP.Port != "" {
		t.Error("no port must be written when enable fails with no port")
	}
}

func TestChangeIntegrationStatusStoredPortWinsOverSeed(t *testing.T) {
	ensureTestLogger()
	tmp := t.TempDir()
	sandboxPointers(t, tmp)
	writeCatalog(t, testCatalog)
	if err := os.WriteFile(config.CollectorFileName, []byte(`{
		"integrations": {
			"firewall-barracuda": {"udp_port": {"enabled": false, "value": "7200"}}
		}
	}`), 0644); err != nil {
		t.Fatal(err)
	}

	port, err := ChangeIntegrationStatus("firewall-barracuda", "udp", true, "7100")
	if err != nil {
		t.Fatalf("ChangeIntegrationStatus: %v", err)
	}
	if port != "7200" {
		t.Errorf("returned port = %q, want stored 7200 (seed must not override a non-empty port)", port)
	}
	cnf := readStored(t)
	udp := cnf.Integrations["firewall-barracuda"].UDP
	if !udp.IsListen {
		t.Error("udp not marked listening after enable")
	}
	if udp.Port != "7200" {
		t.Errorf("persisted udp port = %q, want unchanged 7200", udp.Port)
	}
}

func TestChangeIntegrationStatusPortConflictStillErrors(t *testing.T) {
	ensureTestLogger()
	tmp := t.TempDir()
	sandboxPointers(t, tmp)
	writeCatalog(t, testCatalog)
	// netflow listens on 2055 (its built-in default); seeding 2055 must conflict
	if err := os.WriteFile(config.CollectorFileName, []byte(`{
		"integrations": {
			"netflow": {"udp_port": {"enabled": true, "value": "2055"}},
			"firewall-barracuda": {"udp_port": {"enabled": false, "value": ""}}
		}
	}`), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ChangeIntegrationStatus("firewall-barracuda", "udp", true, "2055")
	if err == nil {
		t.Fatal("want port conflict error, got success")
	}
	cnf := readStored(t)
	if cnf.Integrations["firewall-barracuda"].UDP.IsListen {
		t.Error("conflicting enable must not flip IsListen")
	}
}

func TestChangeIntegrationStatusRejectsInvalidPortSeed(t *testing.T) {
	ensureTestLogger()
	tmp := t.TempDir()
	sandboxPointers(t, tmp)
	// netflow is a built-in with no TCP default, so only the seed can supply a port.
	if err := os.WriteFile(config.CollectorFileName, []byte(`{
		"integrations": {
			"netflow": {"tcp_port": {"enabled": false, "value": ""}}
		}
	}`), 0644); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(config.CollectorFileName)
	if err != nil {
		t.Fatal(err)
	}

	for _, bad := range []string{"0", "99999", "abc"} {
		if _, err := ChangeIntegrationStatus("netflow", "tcp", true, bad); err == nil {
			t.Errorf("seed %q accepted, want range error", bad)
		}
	}

	after, err := os.ReadFile(config.CollectorFileName)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("config file was modified even though invalid seeds were rejected")
	}
}

func TestChangeIntegrationStatusDisableDoesNotSeed(t *testing.T) {
	ensureTestLogger()
	tmp := t.TempDir()
	sandboxPointers(t, tmp)
	writeCatalog(t, testCatalog)
	if err := os.WriteFile(config.CollectorFileName, []byte(`{
		"integrations": {
			"firewall-barracuda": {"udp_port": {"enabled": true, "value": "7100"}}
		}
	}`), 0644); err != nil {
		t.Fatal(err)
	}

	port, err := ChangeIntegrationStatus("firewall-barracuda", "udp", false, "7300")
	if err != nil {
		t.Fatalf("ChangeIntegrationStatus disable: %v", err)
	}
	if port != "7100" {
		t.Errorf("disable returned port %q, want stored 7100 (no seed attempted)", port)
	}
	cnf := readStored(t)
	udp := cnf.Integrations["firewall-barracuda"].UDP
	if udp.IsListen {
		t.Error("udp must be not listening after disable")
	}
	if udp.Port != "7100" {
		t.Errorf("stored port changed to %q on disable, want untouched 7100", udp.Port)
	}
}

func wipeConfig(t *testing.T) {
	t.Helper()
	// a "wiped" entry mirrors the post-restart ticket state: enabled, empty port
	if err := os.WriteFile(config.CollectorFileName, []byte(`{
		"integrations": {
			"firewall-barracuda": {
				"udp_port": {"enabled": true, "value": ""},
				"tcp_port": {"enabled": false, "value": ""}
			}
		}
	}`), 0644); err != nil {
		t.Fatal(err)
	}
}

var _ = filepath.Join
