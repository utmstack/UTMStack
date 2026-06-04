package modulekinds_test

import (
	"testing"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/usecase"
)

// expectedKinds is the snapshot of the legacy ModuleName enum that the Go
// factory must expose post-RegisterAll. Order doesn't matter — Has() is the
// check. Keep this list in sync with the legacy enum file:
// backend-legacy/src/main/java/com/park/utmstack/domain/application_modules/enums/ModuleName.java
var expectedKinds = []string{
	"AD_AUDIT", "AIX", "APACHE", "APACHE2", "AS_400", "AUDITD",
	"AWS_BEANSTALK", "AWS_CLOUDTRAIL", "AWS_FARGATE", "AWS_IAM_USER",
	"AWS_LAMBDA", "AWS_POSTGRESQL", "AWS_SQL_SERVER", "AWS_TRAFFIC_MIRROR",
	"AZURE", "BITDEFENDER", "CISCO", "CISCO_SWITCH", "CROWDSTRIKE",
	"DECEPTIVE_BYTES", "ELASTICSEARCH", "ESET", "FILE_INTEGRITY", "FIRE_POWER",
	"FORTIGATE", "FORTIWEB", "GCP", "GITHUB", "HAPROXY", "IIS", "JSON",
	"KAFKA", "KASPERSKY", "KIBANA", "LINUX_AGENT", "LINUX_LOGS", "LOGSTASH",
	"MACOS", "MACOS_AGENT", "MERAKI", "MIKROTIK", "MONGODB", "MYSQL", "NATS",
	"NETFLOW", "NGINX", "O365", "ORACLE", "OSQUERY", "PALO_ALTO", "PFSENSE",
	"POSTGRESQL", "REDIS", "SALESFORCE", "SENTINEL_ONE", "SOC_AI",
	"SONIC_WALL", "SOPHOS", "SOPHOS_XG", "SURICATA", "SYSLOG",
	"SYSLOG_GENERIC", "TRAEFIK", "UFW", "UTMSTACK", "VMWARE", "WINDOWS_AGENT",
	"WINDOWS_EVENTS",
}

// TestRegisterAll_RegistersEveryEnumValue is the regression test that keeps
// the factory in lockstep with the legacy enum. If a new ModuleName ships in
// the panel without a Go kind, this test goes red.
func TestRegisterAll_RegistersEveryEnumValue(t *testing.T) {
	f := usecase.NewModuleFactory()
	modulekinds.RegisterAll(f)
	for _, name := range expectedKinds {
		if !f.Has(name) {
			t.Errorf("factory missing kind %q after RegisterAll", name)
		}
	}
}

// TestRegisterAll_KindsReturnTheirName guards against copy-paste errors where
// a kind's Name() returns the wrong enum value (because the const Name in
// some package was edited but the embedded Defaults wasn't).
func TestRegisterAll_KindsReturnTheirName(t *testing.T) {
	f := usecase.NewModuleFactory()
	modulekinds.RegisterAll(f)
	for _, name := range expectedKinds {
		k, ok := f.Get(name)
		if !ok {
			t.Errorf("missing kind %q", name)
			continue
		}
		if k.Name() != name {
			t.Errorf("kind registered as %q reports Name()=%q", name, k.Name())
		}
	}
}
