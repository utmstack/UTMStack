package types

import (
	"fmt"

	"github.com/utmstack/UTMStack/installer/utils"
	"gopkg.in/yaml.v3"
)

type Logging struct {
	Driver  *string                `yaml:"driver,omitempty"`
	Options map[string]interface{} `yaml:"options,omitempty"`
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
	CPUs    *string                  `yaml:"cpus,omitempty"`
	Memory  *string                  `yaml:"memory,omitempty"`
	Devices []map[string]interface{} `yaml:"devices,omitempty"`
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

type Volume map[string]interface{}

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

func (c *Compose) Populate(conf *Config, stack *StackConfig) error {
	var pluginsConfig PluginsConfig
	err := pluginsConfig.Set(conf, stack)
	if err != nil {
		return err
	}

	c.Services = make(map[string]Service)
	c.Volumes = make(map[string]Volume)

	pManager := Placement{
		Constraints: []string{"node.role == manager"},
	}

	dLogging := Logging{
		Driver: utils.PointerOf[string]("json-file"),
		Options: map[string]interface{}{
			"max-size": "50m",
		},
	}

	agentManagerMem := stack.ServiceResources["agentmanager"].AssignedMemory
	c.Services["agentmanager"] = Service{
		Image: utils.PointerOf[string]("ghcr.io/utmstack/utmstack/agent-manager:" + conf.Branch),
		Volumes: []string{
			stack.Cert + ":/cert",
			stack.AgentManager + ":/data",
		},
		Ports: []string{
			"9000:50051",
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
			"node1",
		},
	}

	postgresMem := stack.ServiceResources["postgres"].AssignedMemory
	postgresMin := stack.ServiceResources["postgres"].MinMemory
	c.Services["postgres"] = Service{
		Image: utils.PointerOf[string]("postgres:13"),
		Environment: []string{
			"POSTGRES_PASSWORD=" + conf.Password,
			"PGDATA=/var/lib/postgresql/data/pgdata",
		},
		Volumes: []string{
			"postgres_data:/var/lib/postgresql/data",
		},
		Ports: []string{
			"5432:5432",
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
		Image: utils.PointerOf[string]("ghcr.io/utmstack/utmstack/frontend:" + conf.Branch),
		DependsOn: []string{
			"backend",
			"filebrowser",
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
	c.Services["backend"] = Service{
		Image: utils.PointerOf[string]("ghcr.io/utmstack/utmstack/backend:" + conf.Branch),
		DependsOn: []string{
			"postgres",
			"node1",
			"agentmanager",
		},
		Environment: []string{
			"SERVER_NAME=" + conf.ServerName,
			"DB_USER=postgres",
			"DB_PASS=" + conf.Password,
			"DB_HOST=postgres",
			"DB_PORT=5432",
			"DB_NAME=utmstack",
			"ELASTICSEARCH_HOST=node1",
			"ELASTICSEARCH_PORT=9200",
			"INTERNAL_KEY=" + conf.InternalKey,
			"ENCRYPTION_KEY=" + conf.InternalKey,
			"SOC_AI_BASE_URL=http://socai:8080/process",
			"GRPC_AGENT_MANAGER_HOST=agentmanager",
			"GRPC_AGENT_MANAGER_PORT=50051",
		},
		Volumes: []string{
			stack.Datasources + ":/etc/utmstack",
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
		Image: utils.PointerOf[string]("ghcr.io/threatwinds/eventprocessor/base:1.0.0-beta"),
		DependsOn: utils.Mode(conf.ServerType, map[string]interface{}{
			"aio": []string{
				"postgres",
				"node1",
				"backend",
			},
			"cloud": []string{
				"postgres",
				"node1",
				"backend",
			},
		}).([]string),
		Ports: []string{
			"50051:50051",
			"8080:8080",
		},
		Volumes: []string{
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "pipeline") + ":/workdir/pipeline",
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "geolocation") + ":/workdir/geolocation",
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "rules") + ":/workdir/rules/utmstack",
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "plugins") + ":/workdir/plugins/utmstack",
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "logs") + ":/workdir/logs",
			stack.Cert + ":/cert",
		},
		Environment: []string{
			"WORK_DIR=/workdir",
			"LOG_LEVEL=100",
			"GIN_MODE=release",
			"MODE=worker",
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
		Image: utils.PointerOf[string]("ghcr.io/threatwinds/eventprocessor/base:1.0.0-beta"),
		DependsOn: utils.Mode(conf.ServerType, map[string]interface{}{
			"aio": []string{
				"postgres",
				"node1",
				"backend",
			},
			"cloud": []string{
				"postgres",
				"node1",
				"backend",
			},
		}).([]string),
		Ports: []string{
			"8000:8000",
		},
		Volumes: []string{
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "pipeline") + ":/workdir/pipeline",
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "geolocation") + ":/workdir/geolocation",
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "rules") + ":/workdir/rules/utmstack",
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "plugins") + ":/workdir/plugins/utmstack",
			utils.MakeDir(0777, stack.EventsEngineWorkdir, "logs") + ":/workdir/logs",
			stack.Cert + ":/cert",
		},
		Environment: []string{
			"WORK_DIR=/workdir",
			"LOG_LEVEL=100",
			"GIN_MODE=release",
			"MODE=manager",
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

	// Calculating the less closed odd value to assign the final memory, because if an even value is returned
	// opensearch raises an error
	opensearchMem := utils.GetOddValue(stack.ServiceResources["opensearch"].AssignedMemory)
	// temporary create node1 always
	if true {
		c.Services["node1"] = Service{
			Image: utils.PointerOf[string]("opensearchproject/opensearch:2"),
			Ports: []string{
				"9200:9200",
			},
			Volumes: []string{
				stack.ESData + ":/usr/share/opensearch/data",
				stack.ESBackups + ":/usr/share/opensearch/backups",
				stack.Cert + ":/usr/share/opensearch/config/certificates:ro",
			},
			Environment: []string{
				"cluster.name=utmstack",
				"node.name=node1",
				"discovery.seed_hosts=node1",
				"cluster.initial_master_nodes=node1",
				"bootstrap.memory_lock=false",
				"DISABLE_SECURITY_PLUGIN=true",
				"DISABLE_INSTALL_DEMO_CONFIG:true",
				"JAVA_HOME:/usr/share/opensearch/jdk",
				"action.auto_create_index:true",
				"compatibility.override_main_response_version:true",
				"opensearch_security.disabled: true",
				"path.repo=/usr/share/opensearch",
				fmt.Sprintf("OPENSEARCH_JAVA_OPTS=-Xms%dm -Xmx%dm", opensearchMem/2, opensearchMem/2),
				"network.host:0.0.0.0",
			},
			Logging: &dLogging,
			Deploy: &Deploy{
				Placement: &pManager,
				Resources: &Resources{
					Limits: &Res{
						Memory: utils.PointerOf[string](fmt.Sprintf("%vM", opensearchMem)),
					},
					Reservations: &Res{
						Memory: utils.PointerOf[string](fmt.Sprintf("%vM", opensearchMem/2)),
					},
				},
			},
		}
	}

	socAIMem := stack.ServiceResources["socai"].AssignedMemory
	c.Services["socai"] = Service{
		Image: utils.PointerOf[string]("ghcr.io/utmstack/soc-ai/soc-ai:" + conf.Branch),
		DependsOn: []string{
			"node1",
			"backend",
		},
		Environment: []string{
			"GPT_MODEL=gpt-3.5-turbo-1106",
			"PANEL_SERV_NAME=http://backend:8080",
			"OPENSEARCH_HOST=node1",
			"OPENSEARCH_PORT=9200",
			"INTERNAL_KEY=" + conf.InternalKey,
			"ENCRYPTION_KEY=" + conf.InternalKey,
		},
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", socAIMem)),
				},
			},
		},
	}

	userAuditorMem := stack.ServiceResources["user-auditor"].AssignedMemory
	c.Services["user-auditor"] = Service{
		Image: utils.PointerOf[string]("ghcr.io/utmstack/utmstack/user-auditor:" + conf.Branch),
		DependsOn: []string{
			"postgres",
			"node1",
		},
		Environment: []string{
			"SERVER_NAME=" + conf.ServerName,
			"INTERNAL_KEY=" + conf.InternalKey,
			"DB_USER=postgres",
			"DB_HOST=postgres",
			"DB_PORT=5432",
			"DB_NAME=userauditor",
			"DB_PASS=" + conf.Password,
			"ELASTICSEARCH_HOST=node1",
			"ELASTICSEARCH_PORT=9200",
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", userAuditorMem)),
				},
			},
		},
	}

	webPDFMem := stack.ServiceResources["web-pdf"].AssignedMemory
	c.Services["web-pdf"] = Service{
		Image: utils.PointerOf[string]("ghcr.io/utmstack/utmstack/web-pdf:" + conf.Branch),
		Volumes: []string{
			stack.ShmFolder + ":/dev/shm",
		},
		DependsOn: []string{
			"backend",
			"frontend",
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", webPDFMem)),
				},
			},
		},
	}

	c.Volumes["postgres_data"] = Volume{
		"external": false,
	}

	c.Volumes["agent_manager"] = Volume{
		"external": false,
	}

	return nil
}
