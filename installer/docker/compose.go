package docker

import (
	"fmt"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
	"gopkg.in/yaml.v3"
)

type Logging struct {
	Driver  *string        `yaml:"driver,omitempty"`
	Options map[string]any `yaml:"options,omitempty"`
}

type Placement struct {
	Constraints []string          `yaml:"constraints,omitempty"`
	Preferences map[string]string `yaml:"preferences,omitempty"`
}

type Deploy struct {
	Mode           *string           `yaml:"mode,omitempty"`
	Replicas       *int              `yaml:"replicas,omitempty"`
	EndpointMode   *string           `yaml:"endpoint_mode,omitempty"`
	Labels         map[string]string `yaml:"labels,omitempty"`
	Placement      *Placement        `yaml:"placement,omitempty"`
	Resources      *Resources        `yaml:"resources,omitempty"`
	RestartPolicy  *RestartPolicy    `yaml:"restart_policy,omitempty"`
	RollbackConfig *RUConfig         `yaml:"rollback_config,omitempty"`
	UpdateConfig   *RUConfig         `yaml:"update_config,omitempty"`
}

type RUConfig struct {
	Parallelism     *int     `yaml:"parallelism,omitempty"`
	Delay           *string  `yaml:"delay,omitempty"`
	FailureAction   *string  `yaml:"failure_action,omitempty"`
	Monitor         *string  `yaml:"monitor,omitempty"`
	MaxFailureRatio *float64 `yaml:"max_failure_ratio,omitempty"`
	Order           *string  `yaml:"order,omitempty"`
}

type RestartPolicy struct {
	Condition   *string `yaml:"condition,omitempty"`
	Delay       *string `yaml:"delay,omitempty"`
	MaxAttempts *int    `yaml:"max_attempts,omitempty"`
	Window      *string `yaml:"window,omitempty"`
}

type Res struct {
	CPUs    *string          `yaml:"cpus,omitempty"`
	Memory  *string          `yaml:"memory,omitempty"`
	Devices []map[string]any `yaml:"devices,omitempty"`
}

type Resources struct {
	Limits       *Res `yaml:"limits,omitempty"`
	Reservations *Res `yaml:"reservations,omitempty"`
}

type Service struct {
	Image       *string  `yaml:"image,omitempty"`
	Volumes     []string `yaml:"volumes,omitempty"`
	Ports       []string `yaml:"ports,omitempty"`
	Environment []string `yaml:"environment,omitempty"`
	DependsOn   []string `yaml:"depends_on,omitempty"`
	Logging     *Logging `yaml:"logging,omitempty"`
	Deploy      *Deploy  `yaml:"deploy,omitempty"`
	Command     []string `yaml:"command,omitempty,flow"`
}

type Volume map[string]any

type Compose struct {
	Volumes  map[string]Volume  `yaml:"volumes,omitempty"`
	Services map[string]Service `yaml:"services,omitempty"`
}

func (c *Compose) Encode() ([]byte, error) {
	b, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (c *Compose) Populate(conf *config.Config, stack *StackConfig) error {
	err := SetPluginsConfigs(conf, stack)
	if err != nil {
		return err
	}

	// Before any service is declared: the store cannot create its own tables.
	if err := writeClickHouseSchema(stack.ClickHouseSchema); err != nil {
		return err
	}
	if err := writeClickHouseSettings(stack.ClickHouseConf); err != nil {
		return err
	}

	c.Services = make(map[string]Service)
	c.Volumes = make(map[string]Volume)

	pManager := Placement{
		Constraints: []string{"node.role == manager"},
	}

	dLogging := Logging{
		Driver: utils.PointerOf[string]("json-file"),
		Options: map[string]any{
			"max-size": "50m",
		},
	}

	agentManagerMem := stack.ServiceResources["agentmanager"].AssignedMemory
	c.Services["agentmanager"] = Service{
		Image: utils.PointerOf[string]("ghcr.io/utmstack/utmstack/agent-manager:${UTMSTACK_TAG}"),
		Volumes: []string{
			stack.Cert + ":/cert",
			conf.UpdatesFolder + ":/updates",
		},
		Ports: []string{
			"9000:9000",
			"9001:9001",
		},
		Environment: []string{
			"DB_PATH=/data/utmstack.db",
			"INTERNAL_KEY=" + conf.InternalKey,
			"ENCRYPTION_KEY=" + conf.InternalKey,
			"UTM_HOST=http://backend:8080",
			"DB_USER=postgres",
			"DB_PASSWORD=" + conf.Password,
			"DB_HOST=postgres",
			"DB_PORT=5432",
			"DB_NAME=agentmanager",
			"PANEL_SERV_NAME=http://backend:8080",
			"REDIS_ADDR=redis:6379",
			"REDIS_PASSWORD=" + conf.Password,
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", agentManagerMem)),
				},
			},
		},
		DependsOn: []string{
			"postgres",
			"redis",
		},
	}

	postgresMem := stack.ServiceResources["postgres"].AssignedMemory
	postgresMin := stack.ServiceResources["postgres"].MinMemory
	c.Services["postgres"] = Service{
		Image: utils.PointerOf[string]("ghcr.io/utmstack/utmstack/postgres:latest"),
		Environment: []string{
			"POSTGRES_PASSWORD=" + conf.Password,
			"PGDATA=/var/lib/postgresql/data/pgdata",
		},
		Volumes: []string{
			"postgres_data:/var/lib/postgresql/data",
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", postgresMem)),
				},
				Reservations: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", postgresMin)),
				},
			},
		},
		Command: []string{"postgres", "-c", "shared_buffers=256MB", "-c", "max_connections=1000"},
	}

	frontEndMem := stack.ServiceResources["frontend"].AssignedMemory
	c.Services["frontend"] = Service{
		Image: utils.PointerOf[string]("ghcr.io/utmstack/utmstack/frontend:${UTMSTACK_TAG}"),
		DependsOn: []string{
			"backend",
		},
		Ports: []string{
			"10001:80",
		},
		Volumes: []string{
			stack.Cert + ":/etc/nginx/cert",
			stack.FrontEndNginx + ":/etc/nginx/conf.d",
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", frontEndMem)),
				},
			},
		},
	}

	backendMem := stack.ServiceResources["backend"].AssignedMemory
	backendMin := stack.ServiceResources["backend"].MinMemory
	backendEnv := []string{
		"SERVER_NAME=" + conf.ServerName,
		"DB_USER=postgres",
		"DB_PASS=" + conf.Password,
		"DB_HOST=postgres",
		"DB_PORT=5432",
		"DB_NAME=utmstack",
		"CLICKHOUSE_HOST=clickhouse",
		"CLICKHOUSE_PORT=9000",
		"CLICKHOUSE_DB=utmstack",
		"CLICKHOUSE_USER=default",
		"CLICKHOUSE_PASSWORD=" + conf.Password,
		"INTERNAL_KEY=" + conf.InternalKey,
		"ENCRYPTION_KEY=" + conf.InternalKey,
		"GRPC_AGENT_MANAGER_HOST=agentmanager",
		"GRPC_AGENT_MANAGER_PORT=9000",
		"EVENT_PROCESSOR_HOST=event-processor-manager",
		"EVENT_PROCESSOR_PORT=9002",
		"SOC_AI_BASE_URL=http://event-processor-manager:8090",
		"PLAYGROUND_BASE_URL=http://event-processor-manager:8091",
		"UPLOAD_DIR=/uploads",
		"CLICKHOUSE_CONFIG_DIR=/clickhouse-conf",
		"UTMSTACK_ADMIN_PASSWORD=" + conf.Password,
	}

	// Disable TFA in dev and rc environments
	if conf.Branch == "dev" || conf.Branch == "rc" {
		backendEnv = append(backendEnv, "APP_TFA_ENABLED=false")
	}

	c.Services["backend"] = Service{
		Image: utils.PointerOf[string]("ghcr.io/utmstack/utmstack/backend:${UTMSTACK_TAG}"),
		DependsOn: []string{
			"postgres",
			"clickhouse",
			"agentmanager",
		},
		Environment: backendEnv,
		Volumes: []string{
			conf.UpdatesFolder + ":/updates",
			utils.MakeDir(0777, conf.DataDir, "uploads") + ":/uploads",
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "soar") + ":/workdir/soar",
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "compliance") + ":/workdir/compliance",
			// Shared with the event-processor: the backend authors the rules and
			// pipeline (tenants/patterns/filters); the EP reads the same host dirs.
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "rules") + ":/workdir/rules",
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "pipeline") + ":/workdir/pipeline",
			// Where cold storage is declared to ClickHouse, which reads the same
			// directory read-only.
			stack.ClickHouseConfigD + ":/clickhouse-conf",
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", backendMem)),
				},
				Reservations: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", backendMin)),
				},
			},
		},
	}

	epMem := stack.ServiceResources["event-processor"].AssignedMemory
	epMin := stack.ServiceResources["event-processor"].MinMemory
	c.Services["event-processor-worker"] = Service{
		Image: utils.PointerOf[string]("ghcr.io/utmstack/utmstack/eventprocessor:${UTMSTACK_TAG}"),
		DependsOn: utils.Mode(conf.ServerType, map[string]any{
			"aio": []string{
				"postgres",
				"clickhouse",
				"nats",
				"backend",
			},
			"cloud": []string{
				"postgres",
				"clickhouse",
				"nats",
				"backend",
			},
		}).([]string),
		Volumes: []string{
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "pipeline") + ":/workdir/pipeline",
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "logs") + ":/workdir/logs",
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "rules") + ":/workdir/rules/utmstack",
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "soar") + ":/workdir/soar",
			stack.Cert + ":/cert",
			conf.UpdatesFolder + ":/updates",
		},
		Environment: []string{
			"WORK_DIR=/workdir",
			"LOG_LEVEL=200",
			"GIN_MODE=release",
			"MODE=worker",
			"NATS_URL=nats://nats:4222",
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Mode: utils.PointerOf[string]("global"),
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", epMem)),
				},
				Reservations: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", epMin)),
				},
			},
		},
	}

	c.Services["event-processor-manager"] = Service{
		Image: utils.PointerOf[string]("ghcr.io/utmstack/utmstack/eventprocessor:${UTMSTACK_TAG}"),
		DependsOn: utils.Mode(conf.ServerType, map[string]any{
			"aio": []string{
				"postgres",
				"clickhouse",
				"nats",
				"backend",
			},
			"cloud": []string{
				"postgres",
				"clickhouse",
				"nats",
				"backend",
			},
		}).([]string),
		Ports: []string{
			"8000:8000",
		},
		Volumes: []string{
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "pipeline") + ":/workdir/pipeline",
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "rules") + ":/workdir/rules/utmstack",
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "logs") + ":/workdir/logs",
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "soar") + ":/workdir/soar",
			stack.Cert + ":/cert",
			conf.UpdatesFolder + ":/updates",
		},
		Environment: []string{
			"WORK_DIR=/workdir",
			"LOG_LEVEL=200",
			"GIN_MODE=release",
			"MODE=manager",
			"NODE_NAME=manager",
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &Placement{
				Constraints: []string{"node.role == manager"},
			},
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", epMem)),
				},
				Reservations: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", epMin)),
				},
			},
		},
	}

	clickHouseMem := stack.ServiceResources["clickhouse"].AssignedMemory
	c.Services["clickhouse"] = Service{
		Image: utils.PointerOf[string]("clickhouse/clickhouse-server:25.8"),
		Volumes: []string{
			stack.ClickHouseData + ":/var/lib/clickhouse",
			// Run on first boot only; this is what creates the tables.
			stack.ClickHouseSchema + ":/docker-entrypoint-initdb.d:ro",
			stack.ClickHouseConf + "/" + settingsFileName + ":/etc/clickhouse-server/users.d/" + settingsFileName + ":ro",
			// Read-only here and writable in the backend: one writer, one reader.
			stack.ClickHouseConfigD + ":/etc/clickhouse-server/config.d:ro",
		},
		Environment: []string{
			"CLICKHOUSE_DB=utmstack",
			"CLICKHOUSE_PASSWORD=" + conf.Password,
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", clickHouseMem)),
				},
				Reservations: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", clickHouseMem/2)),
				},
			},
		},
	}

	natsMem := stack.ServiceResources["nats"].AssignedMemory
	c.Services["nats"] = Service{
		Image: utils.PointerOf[string]("nats:2.10-alpine"),
		Volumes: []string{
			stack.NATSData + ":/data",
		},
		// JetStream on disk: the queue between ingest and processing has to
		// survive a restart of either side.
		Command: []string{"-js", "-sd", "/data"},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", natsMem)),
				},
			},
		},
	}

	redisMem := stack.ServiceResources["redis"].AssignedMemory
	c.Services["redis"] = Service{
		Image: utils.PointerOf[string]("redis:7-alpine"),
		Volumes: []string{
			stack.RedisData + ":/data",
		},
		// Holds the shared auth cache, which is rebuildable, so eviction under
		// pressure is preferable to refusing writes.
		Command: []string{
			"redis-server", "--requirepass", conf.Password,
			"--maxmemory-policy", "allkeys-lru",
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", redisMem)),
				},
			},
		},
	}

	logInputMem := stack.ServiceResources["log-input"].AssignedMemory
	c.Services["log-input"] = Service{
		Image: utils.PointerOf[string]("ghcr.io/utmstack/utmstack/log-input:${UTMSTACK_TAG}"),
		DependsOn: []string{
			"nats",
			"redis",
			"agentmanager",
		},
		Ports: []string{
			"50051:50051",
			"50052:50052",
		},
		Volumes: []string{
			stack.Cert + ":/cert",
		},
		Environment: []string{
			"NATS_URL=nats://nats:4222",
			"REDIS_ADDR=redis:6379",
			"REDIS_PASSWORD=" + conf.Password,
			"AGENT_MANAGER=agentmanager:9000",
			"BACKEND=http://backend:8080",
			"INTERNAL_KEY=" + conf.InternalKey,
			"CERTS_FOLDER=/cert",
		},
		Logging: &dLogging,
		// Global rather than replicated: this is what agents connect to, and it
		// holds no state of its own, so one per node is the cheapest way to
		// scale it with the cluster.
		Deploy: &Deploy{
			Mode: utils.PointerOf[string]("global"),
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", logInputMem)),
				},
			},
		},
	}

	c.Volumes["postgres_data"] = Volume{
		"external": false,
	}

	return nil
}
