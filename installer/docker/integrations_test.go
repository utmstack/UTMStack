package docker

import (
	"os"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/system"
)

func composeFor(t *testing.T) *Compose {
	t.Helper()

	root := t.TempDir()
	stack := &StackConfig{
		EventsEngineWorkdir: root + "/engine",
		Cert:                root + "/cert",
		ClickHouseSchema:    root + "/ch-init",
		ClickHouseConf:      root + "/ch-conf",
		ClickHouseConfigD:   root + "/ch-config.d",
		ClickHouseData:      root + "/ch-data",
		ClickHouseCold:      root + "/ch-cold",
		NATSData:            root + "/nats",
		ValkeyData:          root + "/valkey",
		FrontEndNginx:       root + "/nginx",
		ServiceResources:    map[string]*system.ServiceConfig{},
	}

	for _, dir := range []string{stack.ClickHouseSchema, stack.ClickHouseConf, stack.ClickHouseConfigD} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	for _, name := range []string{"event-processor", "clickhouse", "log-input", "nats", "valkey", "backend", "postgres", "agentmanager", "frontend", "soc-ai"} {
		stack.ServiceResources[name] = &system.ServiceConfig{Name: name, AssignedMemory: 512}
	}

	for _, name := range Integrations {
		stack.ServiceResources[integrationServiceName(name)] = &system.ServiceConfig{Name: integrationServiceName(name), AssignedMemory: 300}
	}

	conf := &config.Config{Password: "pw", InternalKey: "key", DataDir: root, ServerType: "aio", UpdatesFolder: root + "/updates"}

	c := new(Compose)
	if err := c.Populate(conf, stack); err != nil {
		t.Fatal(err)
	}

	return c
}

func TestEveryPullerGetsItsOwnContainer(t *testing.T) {
	c := composeFor(t)

	if len(Integrations) != 8 {
		t.Fatalf("expected 8 integrations, got %d", len(Integrations))
	}

	for _, name := range Integrations {
		svc, ok := c.Services["integration-"+name]
		if !ok {
			t.Fatalf("no service for %s", name)
		}

		if got := *svc.Image; got != "ghcr.io/utmstack/utmstack/"+name+":${UTMSTACK_TAG}" {
			t.Errorf("%s image %q", name, got)
		}

		env := strings.Join(svc.Environment, " ")
		if !strings.Contains(env, "NATS_URL=nats://nats:4222") {
			t.Errorf("%s does not know where NATS is: %v", name, svc.Environment)
		}

		if !strings.Contains(env, "ENCRYPTION_KEY=key") {
			t.Errorf("%s cannot decrypt its configuration: %v", name, svc.Environment)
		}

		for _, v := range svc.Volumes {
			if strings.Contains(v, "/workdir") {
				t.Errorf("%s still mounts the pipeline folder: %v", name, svc.Volumes)
			}
		}

		if *svc.Deploy.Resources.Limits.Memory != "300M" {
			t.Errorf("%s memory %q", name, *svc.Deploy.Resources.Limits.Memory)
		}
	}
}

func TestTheBackendKeepsTheIntegrationConfigurationInNATS(t *testing.T) {
	c := composeFor(t)
	b := c.Services["backend"]

	if !strings.Contains(strings.Join(b.Environment, " "), "NATS_URL=nats://nats:4222") {
		t.Errorf("the backend does not know where NATS is: %v", b.Environment)
	}

	has := false

	for _, d := range b.DependsOn {
		has = has || d == "nats"
	}

	if !has {
		t.Errorf("the backend must wait for NATS: %v", b.DependsOn)
	}
}

func TestOnlyBitdefenderMountsTheCertificate(t *testing.T) {
	c := composeFor(t)

	for _, name := range Integrations {
		has := strings.Contains(strings.Join(c.Services["integration-"+name].Volumes, " "), ":/cert")
		if has != (name == "bitdefender") {
			t.Errorf("%s: certificate mounted=%v", name, has)
		}
	}
}

func TestNoPluginServiceOutsideTheEventProcessorRemains(t *testing.T) {
	c := composeFor(t)

	if _, ok := c.Services["event-processor-worker"]; !ok {
		t.Error("the worker is gone")
	}

	data, err := c.Encode()
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(data), "plugin-binaries") {
		t.Error("the compose still refers to plugin binaries")
	}
}


func TestTheSOCAgentIsItsOwnServiceTheBackendPointsAt(t *testing.T) {
	c := composeFor(t)

	svc, ok := c.Services["soc-ai"]
	if !ok {
		t.Fatal("there is no soc-ai service")
	}

	if got := *svc.Image; got != "ghcr.io/utmstack/utmstack/soc-ai:${UTMSTACK_TAG}" {
		t.Errorf("image %q", got)
	}

	env := strings.Join(svc.Environment, " ")
	for _, want := range []string{"BACKEND_URL=http://backend:8080", "INTERNAL_KEY=key", "ENCRYPTION_KEY=key", "NATS_URL=nats://nats:4222"} {
		if !strings.Contains(env, want) {
			t.Errorf("missing %s: %v", want, svc.Environment)
		}
	}

	if len(svc.Volumes) != 0 {
		t.Errorf("it needs nothing from the host: %v", svc.Volumes)
	}

	if !strings.Contains(strings.Join(c.Services["backend"].Environment, " "), "SOC_AI_BASE_URL=http://soc-ai:8090") {
		t.Errorf("the backend must reach the service by its own name: %v", c.Services["backend"].Environment)
	}

	if strings.Contains(strings.Join(c.Services["backend"].Environment, " "), "event-processor-manager:8090") {
		t.Error("the backend still points at the old plugin")
	}
}
