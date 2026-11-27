package collector

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

func ensureTestLogger() {
	testLoggerOnce.Do(func() {
		utils.InitLogger(os.DevNull)
	})
}

// sandboxPointers redirects the package-level config file variables to tmp
// so tests never touch real files on disk.
func sandboxPointers(t *testing.T, tmp string) {
	t.Helper()
	orig := config.CollectorFileName
	config.CollectorFileName = filepath.Join(tmp, "log-collector-config.json")
	t.Cleanup(func() { config.CollectorFileName = orig })

	origCat := config.CatalogFileName
	config.CatalogFileName = filepath.Join(tmp, "log-collector-catalog.json")
	t.Cleanup(func() { config.CatalogFileName = origCat })

	origLock := config.CatalogLockFile
	config.CatalogLockFile = filepath.Join(tmp, "log-collector-catalog.lock")
	t.Cleanup(func() { config.CatalogLockFile = origLock })
}

func writeCatalog(t *testing.T, contents string) {
	t.Helper()
	if err := os.WriteFile(config.CatalogFileName, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
}

func readStored(t *testing.T) schema.CollectorConfig {
	t.Helper()
	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		t.Fatal(err)
	}
	return cnf
}

// ---- T1: SyncCollectorConfig must keep user catalog types (S1, S4) ----

func TestSyncCollectorConfigKeepsUserCatalogType(t *testing.T) {
	ensureTestLogger()
	tmp := t.TempDir()
	sandboxPointers(t, tmp)

	writeCatalog(t, `{"user_datatypes":{"firewall-barracuda":{"kind":"syslog","udp":"7100","tcp":"7101"}}}`)
	if err := os.WriteFile(config.CollectorFileName, []byte(`{
		"integrations": {
			"firewall-barracuda": {
				"udp_port": {"enabled": true, "value": "7100"},
				"tcp_port": {"enabled": false, "value": "7101"}
			},
			"cisco-switch": {
				"udp_port": {"enabled": false, "value": "514"}
			}
		}
	}`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := SyncCollectorConfig(); err != nil {
		t.Fatal(err)
	}

	cnf := readStored(t)
	barracuda, ok := cnf.Integrations["firewall-barracuda"]
	if !ok {
		t.Fatal("SyncCollectorConfig removed user catalog integration firewall-barracuda")
	}
	if barracuda.UDP.Port != "7100" || barracuda.UDP.IsListen != true {
		t.Errorf("firewall-barracuda udp = %+v, want port 7100 listening", barracuda.UDP)
	}
	if barracuda.TCP.Port != "7101" {
		t.Errorf("firewall-barracuda tcp = %+v, want port 7101 untouched", barracuda.TCP)
	}
	// regression pin: a real built-in must still survive
	if _, ok := cnf.Integrations["cisco-switch"]; !ok {
		t.Error("SyncCollectorConfig removed built-in cisco-switch")
	}
}

func TestSyncCollectorConfigStillRemovesTrueOrphans(t *testing.T) {
	ensureTestLogger()
	tmp := t.TempDir()
	sandboxPointers(t, tmp)

	writeCatalog(t, `{"user_datatypes":{"firewall-barracuda":{"kind":"syslog","udp":"7100"}}}`)
	if err := os.WriteFile(config.CollectorFileName, []byte(`{
		"integrations": {
			"firewall-barracuda": {"udp_port": {"enabled": false, "value": "7100"}},
			"totally-unknown": {"udp_port": {"enabled": false, "value": "9999"}}
		}
	}`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := SyncCollectorConfig(); err != nil {
		t.Fatal(err)
	}

	cnf := readStored(t)
	if _, ok := cnf.Integrations["totally-unknown"]; ok {
		t.Error("SyncCollectorConfig kept true orphan totally-unknown (not in ProtoPorts, not in catalog)")
	}
	if _, ok := cnf.Integrations["firewall-barracuda"]; !ok {
		t.Error("SyncCollectorConfig removed catalog user type firewall-barracuda")
	}
}

func TestSyncCollectorConfigAddsMissingBuiltIns(t *testing.T) {
	ensureTestLogger()
	tmp := t.TempDir()
	sandboxPointers(t, tmp)

	// no catalog: only user types are expected to be kept; everything else is built-in
	if err := os.WriteFile(config.CollectorFileName, []byte(`{"integrations": {}}`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := SyncCollectorConfig(); err != nil {
		t.Fatal(err)
	}

	cnf := readStored(t)
	syslog, ok := cnf.Integrations["syslog"]
	if !ok {
		t.Fatal("SyncCollectorConfig did not add missing built-in syslog")
	}
	if syslog.UDP.Port != "7014" || syslog.TCP.Port != "7014" {
		t.Errorf("syslog added with ports %+v, want 7014/7014", syslog)
	}
	if syslog.UDP.IsListen || syslog.TCP.IsListen {
		t.Errorf("missing built-in syslog added as listening: %+v", syslog)
	}
}

func TestSyncCollectorConfigCorruptCatalogSkipsPruning(t *testing.T) {
	ensureTestLogger()
	tmp := t.TempDir()
	sandboxPointers(t, tmp)

	// The catalog EXISTS but is unreadable/corrupt: S4 fail-safe -> no pruning
	// of either Integrations or FileIntegrations.
	writeCatalog(t, `{this is not json`)
	if err := os.WriteFile(config.CollectorFileName, []byte(`{
		"integrations": {
			"firewall-barracuda": {"udp_port": {"enabled": false, "value": "7100"}},
			"totally-unknown": {"udp_port": {"enabled": false, "value": "9999"}},
			"netflow": {"udp_port": {"enabled": false, "value": "2055"}}
		},
		"file_integrations": {
			"my-user-file": {"enabled": true, "paths": ["/var/log/myapp"]},
			"nginx": {"enabled": false, "paths": ["/var/log/nginx/"]}
		}
	}`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := SyncCollectorConfig(); err != nil {
		t.Fatal(err)
	}

	cnf := readStored(t)
	for _, name := range []string{"firewall-barracuda", "totally-unknown", "netflow"} {
		if _, ok := cnf.Integrations[name]; !ok {
			t.Errorf("corrupt catalog: integration %q was removed, want no pruning", name)
		}
	}
	for _, name := range []string{"my-user-file", "nginx"} {
		if _, ok := cnf.FileIntegrations[name]; !ok {
			t.Errorf("corrupt catalog: file integration %q was removed, want no pruning", name)
		}
	}
}

func TestSyncCollectorConfigNoCatalogFilePrunesAsBefore(t *testing.T) {
	ensureTestLogger()
	tmp := t.TempDir()
	sandboxPointers(t, tmp)
	// no catalog file written at all

	if err := os.WriteFile(config.CollectorFileName, []byte(`{
		"integrations": {
			"firewall-barracuda": {"udp_port": {"enabled": false, "value": "7100"}},
			"syslog": {"udp_port": {"enabled": false, "value": "7014"}}
		}
	}`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := SyncCollectorConfig(); err != nil {
		t.Fatal(err)
	}

	cnf := readStored(t)
	if _, ok := cnf.Integrations["firewall-barracuda"]; ok {
		t.Error("no catalog: firewall-barracuda should be pruned as today for unknown names")
	}
	if _, ok := cnf.Integrations["syslog"]; !ok {
		t.Error("no catalog: built-in syslog should survive")
	}
}

func TestSyncCollectorConfigUserFileTypesKept(t *testing.T) {
	ensureTestLogger()
	tmp := t.TempDir()
	sandboxPointers(t, tmp)

	writeCatalog(t, `{"user_datatypes":{"my-user-file":{"kind":"file"}}}`)
	if err := os.WriteFile(config.CollectorFileName, []byte(`{
		"file_integrations": {
			"my-user-file": {"enabled": true, "paths": ["/var/log/myapp"]},
			"totally-unknown-file": {"enabled": false, "paths": ["/var/log/x"]}
		}
	}`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := SyncCollectorConfig(); err != nil {
		t.Fatal(err)
	}

	cnf := readStored(t)
	fi, ok := cnf.FileIntegrations["my-user-file"]
	if !ok {
		t.Fatal("SyncCollectorConfig removed user file type my-user-file")
	}
	if !fi.Enabled || len(fi.Paths) != 1 || fi.Paths[0] != "/var/log/myapp" {
		t.Errorf("my-user-file = %+v, want enabled with path /var/log/myapp", fi)
	}
	if _, ok := cnf.FileIntegrations["totally-unknown-file"]; ok {
		t.Error("true orphan file integration totally-unknown-file should still be removed")
	}
}

func TestSyncCollectorConfigDoesNotModifyKeptPorts(t *testing.T) {
	ensureTestLogger()
	tmp := t.TempDir()
	sandboxPointers(t, tmp)

	writeCatalog(t, `{"user_datatypes":{"firewall-barracuda":{"kind":"syslog","udp":"7100","tcp":"7101"}}}`)
	if err := os.WriteFile(config.CollectorFileName, []byte(`{
		"integrations": {
			"firewall-barracuda": {"udp_port": {"enabled": false, "value": "7999"}, "tcp_port": {"enabled": false, "value": "8888"}},
			"cisco-switch": {"udp_port": {"enabled": false, "value": "514"}}
		}
	}`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := SyncCollectorConfig(); err != nil {
		t.Fatal(err)
	}

	cnf := readStored(t)
	barracuda := cnf.Integrations["firewall-barracuda"]
	if barracuda.UDP.Port != "7999" || barracuda.TCP.Port != "8888" {
		t.Errorf("SyncCollectorConfig modified kept user ports: %+v, want 7999/8888", barracuda)
	}
	sw := cnf.Integrations["cisco-switch"]
	if sw.UDP.Port != "514" {
		t.Errorf("SyncCollectorConfig modified kept built-in port: %+v, want 514", sw)
	}
}
