package upstream

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/utmstack/UTMStack/collectors/forwarder/collector/schema"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
)

var testLoggerOnce sync.Once

func TestMain(m *testing.M) {
	testLoggerOnce.Do(func() {
		utils.InitLogger(os.DevNull)
	})
	os.Exit(m.Run())
}

// applyGroup builds a gRPC-style CollectorConfigGroup (the shape the server
// pushes) and applies it through the same entry point the stream uses.
func applyGroup(t *testing.T, dataType, proto, port string) *ConfigKnowledge {
	t.Helper()
	group := &CollectorConfigGroup{
		GroupName: dataType,
		Configurations: []*CollectorGroupConfigurations{
			{ConfKey: "enabled", ConfValue: "true"},
			{ConfKey: "proto", ConfValue: proto},
			{ConfKey: "port", ConfValue: port},
		},
	}
	return applyCollectorConfigGroup(group)
}

// The gRPC apply path must persist the port the server pushes (S1): before the
// fix, applyNetworkIntegration dropped conf.port and a wiped entry re-enabled
// with an empty port and no listener.
func TestApplyCollectorConfigGroupPersistsPushedPort(t *testing.T) {
	tmp := t.TempDir()
	origCfg := config.CollectorFileName
	config.CollectorFileName = filepath.Join(tmp, "log-collector-config.json")
	t.Cleanup(func() { config.CollectorFileName = origCfg })
	origCat := config.CatalogFileName
	config.CatalogFileName = filepath.Join(tmp, "log-collector-catalog.json")
	t.Cleanup(func() { config.CatalogFileName = origCat })
	origLock := config.CatalogLockFile
	config.CatalogLockFile = filepath.Join(tmp, "log-collector-catalog.lock")
	t.Cleanup(func() { config.CatalogLockFile = origLock })

	// A wiped config: the ticket state, enabled with an empty port.
	if err := os.WriteFile(config.CollectorFileName, []byte(`{
		"integrations": {
			"firewall-barracuda": {"udp_port": {"enabled": true, "value": ""}}
		}
	}`), 0644); err != nil {
		t.Fatal(err)
	}

	result := applyGroup(t, "firewall-barracuda", "udp", "7100")
	if result.GetAccepted() != "true" {
		t.Fatalf("apply rejected: %s", result.GetErrorMessage())
	}

	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		t.Fatal(err)
	}
	udp := cnf.Integrations["firewall-barracuda"].UDP
	if !udp.IsListen {
		t.Error("udp not listening after apply with pushed port")
	}
	if udp.Port != "7100" {
		t.Errorf("persisted udp port = %q, want pushed 7100", udp.Port)
	}
}

// Disabling through the apply path must still succeed for known types.
func TestApplyCollectorConfigGroupDisable(t *testing.T) {
	tmp := t.TempDir()
	origCfg := config.CollectorFileName
	config.CollectorFileName = filepath.Join(tmp, "log-collector-config.json")
	t.Cleanup(func() { config.CollectorFileName = origCfg })
	origCat := config.CatalogFileName
	config.CatalogFileName = filepath.Join(tmp, "log-collector-catalog.json")
	t.Cleanup(func() { config.CatalogFileName = origCat })
	origLock := config.CatalogLockFile
	config.CatalogLockFile = filepath.Join(tmp, "log-collector-catalog.lock")
	t.Cleanup(func() { config.CatalogLockFile = origLock })

	// firewall-barracuda is a user type, so the catalog must list it for the
	// disable path to resolve it and reach the stored entry.
	if err := os.WriteFile(config.CatalogFileName, []byte(`{"user_datatypes":{"firewall-barracuda":{"kind":"syslog","udp":"7100"}}}`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(config.CollectorFileName, []byte(`{
		"integrations": {
			"firewall-barracuda": {"udp_port": {"enabled": true, "value": "7100"}}
		}
	}`), 0644); err != nil {
		t.Fatal(err)
	}

	group := &CollectorConfigGroup{
		GroupName: "firewall-barracuda",
		Configurations: []*CollectorGroupConfigurations{
			{ConfKey: "enabled", ConfValue: "false"},
			{ConfKey: "proto", ConfValue: "udp"},
		},
	}
	result := applyCollectorConfigGroup(group)
	if result.GetAccepted() != "true" {
		t.Fatalf("disable apply rejected: %s", result.GetErrorMessage())
	}

	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		t.Fatal(err)
	}
	udp := cnf.Integrations["firewall-barracuda"].UDP
	if udp.IsListen {
		t.Error("udp still listening after disable apply")
	}
	if udp.Port != "7100" {
		t.Errorf("stored port changed to %q on disable, want untouched 7100", udp.Port)
	}
}
