package types

import (
	sigar "github.com/cloudfoundry/gosigar"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/utmstack/UTMStack/installer/utils"
)

type StackConfig struct {
	FrontEndNginx     string
	LogstashPipelines string
	LogstashConfig    string
	ServiceResources  map[string]*utils.ServiceConfig
	Threads           int
	ESData            string
	ESBackups         string
	Cert              string
	Datasources       string
	Rules             string
	LocksDir          string
	ShmFolder         string
}

var Services = []utils.ServiceConfig{
	{Name: "correlation", Priority: 1, MinMemory: 3 * 1024, MaxMemory: 60 * 1024},
	{Name: "logstash", Priority: 1, MinMemory: 2700, MaxMemory: 60 * 1024},
	{Name: "opensearch", Priority: 1, MinMemory: 4500, MaxMemory: 60 * 1024},
	{Name: "log-auth-proxy", Priority: 1, MinMemory: 1 * 1024, MaxMemory: 4 * 1024},
	{Name: "backend", Priority: 2, MinMemory: 700, MaxMemory: 8 * 1024},
	{Name: "web-pdf", Priority: 2, MinMemory: 1024, MaxMemory: 2 * 1024},
	{Name: "postgres", Priority: 2, MinMemory: 500, MaxMemory: 1 * 1024},
	{Name: "user-auditor", Priority: 3, MinMemory: 200, MaxMemory: 1024},
	{Name: "agentmanager", Priority: 3, MinMemory: 200, MaxMemory: 1024},
	{Name: "mutate", Priority: 3, MinMemory: 50, MaxMemory: 1024},
	{Name: "aws", Priority: 3, MinMemory: 50, MaxMemory: 1024},
	{Name: "filebrowser", Priority: 3, MinMemory: 50, MaxMemory: 512},
	{Name: "sophos", Priority: 3, MinMemory: 50, MaxMemory: 1024},
	{Name: "frontend", Priority: 3, MinMemory: 80, MaxMemory: 1024},
	{Name: "socai", Priority: 3, MinMemory: 30, MaxMemory: 512},
	{Name: "bitdefender", Priority: 3, MinMemory: 30, MaxMemory: 100},
	{Name: "office365", Priority: 3, MinMemory: 30, MaxMemory: 100},
}

func (s *StackConfig) Populate(c *Config) error {
	cores, err := cpu.Counts(false)
	if err != nil {
		return err
	}

	mem := sigar.Mem{}
	err = mem.Get()
	if err != nil {
		return err
	}

	s.Threads = cores
	s.Cert = utils.MakeDir(0777, c.DataDir, "cert")
	s.FrontEndNginx = utils.MakeDir(0777, c.DataDir, "front-end", "nginx")
	s.Datasources = utils.MakeDir(0777, c.DataDir, "datasources")
	s.Rules = utils.MakeDir(0777, c.DataDir, "rules")
	s.LogstashPipelines = utils.MakeDir(0777, c.DataDir, "logstash", "pipelines")
	s.LogstashConfig = utils.MakeDir(0777, c.DataDir, "logstash", "config")
	s.ESData = utils.MakeDir(0777, c.DataDir, "opensearch", "data")
	s.ESBackups = utils.MakeDir(0777, c.DataDir, "opensearch", "backups")
	s.LocksDir = utils.MakeDir(0777, c.DataDir, "locks")
	s.ShmFolder = utils.MakeDir(0777, c.DataDir, "tmpfs")

	total := int(mem.Total / 1024 / 1024)

	rsrcs, err := utils.BalanceMemory(Services, total)
	if err != nil {
		return err
	}

	s.ServiceResources = rsrcs

	return nil
}
