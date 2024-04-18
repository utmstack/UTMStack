package main

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
	Constraints []string `yaml:"constraints,omitempty"`
}

type Deploy struct {
	Placement *Placement `yaml:"placement,omitempty"`
	Resources *Resources `yaml:"resources,omitempty"`
}

type Res struct {
	CPUs   *string `yaml:"cpus,omitempty"`
	Memory *string `yaml:"memory,omitempty"`
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
	Version  *string            `yaml:"version,omitempty"`
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

func (c *Compose) Populate(conf *Config, stack *StackConfig) *Compose {
	err := CreateLogstashConfig(stack.LogstashConfig)
	if err != nil {
		fmt.Println(err)
	}

	c.Version = utils.Str("3.8")
	c.Services = make(map[string]Service)
	c.Volumes = make(map[string]Volume)

	pManager := Placement{
		Constraints: []string{"node.role == manager"},
	}

	dLogging := Logging{
		Driver: utils.Str("json-file"),
		Options: map[string]interface{}{
			"max-size": "50m",
		},
	}

	c.Services["logstash"] = Service{
		Image: utils.Str("utmstack.azurecr.io/logstash:" + conf.Branch),
		Environment: []string{
			"CONFIG_RELOAD_AUTOMATIC=true",
			fmt.Sprintf("LS_JAVA_OPTS=-Xms%dg -Xmx%dg -Xss100m", stack.LSMem, stack.LSMem),
			fmt.Sprintf("PIPELINE_WORKERS=%d", stack.Threads),
		},
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.Str(fmt.Sprintf("%vG", stack.LSMem*2)),
				},
				Reservations: &Res{
					Memory: utils.Str(fmt.Sprintf("%vG", stack.LSMem)),
				},
			},
		},
		Volumes: []string{
			stack.Datasources + ":/etc/utmstack",
			stack.LogstashPipelines + ":/usr/share/logstash/pipelines",
			stack.LogstashConfig + "/pipelines.yml:/usr/share/logstash/config/pipelines.yml",
			stack.LogstashConfig + "/logstash.yml:/usr/share/logstash/config/logstash.yml",
			stack.Cert + ":/cert",
		},
		DependsOn: []string{
			"mutate",
		},
		Logging: &dLogging,
	}

	c.Services["mutate"] = Service{
		Image: utils.Str("ghcr.io/utmstack/utmstack/mutate:" + conf.Branch),
		Volumes: []string{
			stack.Datasources + ":/etc/utmstack",
			stack.LogstashPipelines + ":/usr/share/logstash/pipelines",
			stack.LogstashConfig + "/pipelines.yml:/usr/share/logstash/config/pipelines.yml",
		},
		Environment: []string{
			"SERVER_NAME=" + conf.ServerName,
			"SERVER_TYPE=" + conf.ServerType,
			"ENCRYPTION_KEY=" + conf.InternalKey,
			"DB_HOST=postgres",
			"DB_PASSWORD=" + conf.Password,
			"DB_USER=postgres",
			"DB_NAME=utmstack",
			"DB_PORT=5432",
			"CORRELATION_URL=http://correlation:8080/v1/newlog",
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.Str(fmt.Sprintf("%vG", 1)),
				},
			},
		},
	}

	c.Services["agentmanager"] = Service{
		Image: utils.Str("ghcr.io/utmstack/utmstack/agent-manager:" + conf.Branch),
		Volumes: []string{
			stack.Cert + ":/cert",
			//stack.Datasources + ":/etc/utmstack",
			"agent_manager:/data",
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
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.Str(fmt.Sprintf("%vG", 1)),
				},
			},
		},
		DependsOn: []string{
			"postgres",
			"node1",
		},
		Command: []string{"/app/server"},
	}

	c.Services["postgres"] = Service{
		Image: utils.Str("utmstack.azurecr.io/postgres:" + conf.Branch),
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
					Memory: utils.Str(fmt.Sprintf("%vG", 1)),
				},
				Reservations: &Res{
					Memory: utils.Str(fmt.Sprintf("%vM", 512)),
				},
			},
		},
		Command: []string{"postgres", "-c", "shared_buffers=256MB", "-c", "max_connections=1000"},
	}

	c.Services["frontend"] = Service{
		Image: utils.Str("ghcr.io/utmstack/utmstack/frontend:" + conf.Branch),
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
					Memory: utils.Str(fmt.Sprintf("%vG", 1)),
				},
			},
		},
	}

	c.Services["aws"] = Service{
		Image: utils.Str("utmstack.azurecr.io/datasources:" + conf.Branch),
		DependsOn: []string{
			"postgres",
			"node1",
			"backend",
		},
		Volumes: []string{
			stack.Datasources + ":/etc/utmstack",
		},
		Environment: []string{
			"SERVER_NAME=" + conf.ServerName,
			"ENCRYPTION_KEY=" + conf.InternalKey,
			"DB_PASS=" + conf.Password,
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.Str(fmt.Sprintf("%vG", 1)),
				},
			},
		},
		Command: []string{"python3", "-m", "utmstack.aws"},
	}

	c.Services["office365"] = Service{
		Image: utils.Str("ghcr.io/utmstack/utmstack/office365:" + conf.Branch),
		DependsOn: []string{
			"postgres",
			"node1",
			"backend",
			"correlation",
		},
		Volumes: []string{
			stack.Datasources + ":/etc/utmstack",
		},
		Environment: []string{
			"PANEL_SERV_NAME=backend:8080",
			"INTERNAL_KEY=" + conf.InternalKey,
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.Str(fmt.Sprintf("%vG", 1)),
				},
			},
		},
	}

	c.Services["sophos"] = Service{
		Image: utils.Str("utmstack.azurecr.io/datasources:" + conf.Branch),
		DependsOn: []string{
			"postgres",
			"node1",
			"backend",
		},
		Volumes: []string{
			stack.Datasources + ":/etc/utmstack",
		},
		Environment: []string{
			"SERVER_NAME=" + conf.ServerName,
			"ENCRYPTION_KEY=" + conf.InternalKey,
			"DB_PASS=" + conf.Password,
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.Str(fmt.Sprintf("%vG", 1)),
				},
			},
		},
		Command: []string{"python3", "-m", "utmstack.sophos"},
	}

	c.Services["bitdefender"] = Service{
		Image: utils.Str("ghcr.io/utmstack/utmstack/bitdefender:" + conf.Branch),
		DependsOn: []string{
			"backend",
			"logstash",
		},
		Ports: []string{
			"8000:8000",
		},
		Volumes: []string{
			stack.Datasources + ":/etc/utmstack",
		},
		Environment: []string{
			"PANEL_SERV_NAME=backend:8080",
			"INTERNAL_KEY=" + conf.InternalKey,
			"ENCRYPTION_KEY=" + conf.InternalKey,
			"SYSLOG_PROTOCOL=tcp",
			"SYSLOG_HOST=logstash",
			"SYSLOG_PORT=10019",
			"CONNECTOR_PORT=8000",
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.Str(fmt.Sprintf("%vG", 1)),
				},
			},
		},
	}

	c.Services["backend"] = Service{
		Image: utils.Str("ghcr.io/utmstack/utmstack/backend:" + conf.Branch),
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
			"LOGSTASH_URL=http://logstash:9600",
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
					Memory: utils.Str(fmt.Sprintf("%vG", 1)),
				},
				Reservations: &Res{
					Memory: utils.Str(fmt.Sprintf("%vM", 512)),
				},
			},
		},
	}

	c.Services["filebrowser"] = Service{
		Image: utils.Str("ghcr.io/utmstack/filebrowser/filebrowser:" + conf.Branch),
		Volumes: []string{
			stack.Rules + ":/srv",
		},
		Environment: []string{
			"PASSWORD=" + conf.Password,
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.Str(fmt.Sprintf("%vG", 1)),
				},
			},
		},
	}

	c.Services["correlation"] = Service{
		Image: utils.Str("ghcr.io/utmstack/utmstack/correlation:" + conf.Branch),
		DependsOn: utils.Mode(conf.ServerType, map[string]interface{}{
			"aio": []string{
				"postgres",
				"node1",
				"backend",
			},
			"cloud": []string{
				"postgres",
				"backend",
			},
		}).([]string),
		Ports: []string{
			"10000:8080",
		},
		Volumes: []string{
			stack.Rules + ":/app/rulesets",
			"geoip_data:/app/geosets",
		},
		Environment: []string{
			"SERVER_NAME=" + conf.ServerName,
			"POSTGRESQL_USER=postgres",
			"POSTGRESQL_PASSWORD=" + conf.Password,
			"POSTGRESQL_HOST=postgres",
			"POSTGRESQL_PORT=5432",
			"POSTGRESQL_DATABASE=utmstack",
			"ELASTICSEARCH_HOST=node1",
			"ELASTICSEARCH_PORT=9200",
			"ERROR_LEVEL=info",
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.Str(fmt.Sprintf("%vG", stack.ESMem*3)),
				},
				Reservations: &Res{
					Memory: utils.Str(fmt.Sprintf("%vG", stack.ESMem)),
				},
			},
		},
	}

	if conf.ServerType == "aio" {
		c.Services["node1"] = Service{
			Image: utils.Str("utmstack.azurecr.io/opensearch:" + conf.Branch),
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
				fmt.Sprintf("OPENSEARCH_JAVA_OPTS=-Xms%dg -Xmx%dg", stack.ESMem, stack.ESMem),
				"network.host:0.0.0.0",
			},
			Logging: &dLogging,
			Deploy: &Deploy{
				Placement: &pManager,
				Resources: &Resources{
					Limits: &Res{
						Memory: utils.Str(fmt.Sprintf("%vG", stack.ESMem*2)),
					},
					Reservations: &Res{
						Memory: utils.Str(fmt.Sprintf("%vG", stack.ESMem)),
					},
				},
			},
		}
	}

	c.Services["socai"] = Service{
		Image: utils.Str("ghcr.io/utmstack/soc-ai/soc-ai:" + conf.Branch),
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
					Memory: utils.Str(fmt.Sprintf("%vG", 1)),
				},
			},
		},
	}

	c.Services["log-auth-proxy"] = Service{
		Image: utils.Str("ghcr.io/utmstack/utmstack/log-auth-proxy:" + conf.Branch),
		DependsOn: []string{
			"logstash",
		},
		Ports: []string{
			"50051:50051",
			"8080:8080",
		},
		Volumes: []string{
			stack.Cert + ":/cert",
		},
		Environment: []string{
			"SERVER_NAME=" + conf.ServerName,
			"INTERNAL_KEY=" + conf.InternalKey,
			"UTM_AGENT_MANAGER_HOST=agentmanager:50051",
			"UTM_HOST=http://backend:8080",
			"UTM_LOGSTASH_HOST=logstash",
			"UTM_CERTS_LOCATION=/cert",
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.Str(fmt.Sprintf("%vG", stack.ESMem)),
				},
			},
		},
	}

	c.Services["user-auditor"] = Service{
		Image: utils.Str("ghcr.io/utmstack/utmstack/user-auditor:" + conf.Branch),
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
					Memory: utils.Str("1G"),
				},
			},
		},
	}

	c.Services["web-pdf"] = Service{
		Image: utils.Str("ghcr.io/utmstack/utmstack/web-pdf:" + conf.Branch),
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
					Memory: utils.Str("1G"),
				},
			},
		},
	}

	c.Volumes["postgres_data"] = Volume{
		"external": false,
	}

	c.Volumes["geoip_data"] = Volume{
		"external": false,
	}

	c.Volumes["agent_manager"] = Volume{
		"external": false,
	}

	return c
}
