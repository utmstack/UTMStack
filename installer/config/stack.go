package config

import (
	"fmt"
	"os"
	"sync"

	sigar "github.com/cloudfoundry/gosigar"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/utmstack/UTMStack/installer/utils"
)

type StackConfig struct {
	FrontEndNginx       string
	ServiceResources    map[string]*utils.ServiceConfig
	Threads             int
	ESData              string
	ESBackups           string
	Cert                string
	Datasources         string
	EventsEngineWorkdir string
	LocksDir            string
	ShmFolder           string
}

var (
	stackConfig     *StackConfig
	stackConfigOnce sync.Once
)

func GetStackConfig() *StackConfig {
	stackConfigOnce.Do(func() {
		cnf := GetConfig()

		cores, err := cpu.Counts(false)
		if err != nil {
			fmt.Printf("error getting cpu cores: %v\n", err)
			os.Exit(1)
		}

		mem := sigar.Mem{}
		err = mem.Get()
		if err != nil {
			fmt.Printf("error getting memory: %v\n", err)
			os.Exit(1)
		}

		stackConfig = &StackConfig{}
		stackConfig.Threads = cores
		stackConfig.Cert = utils.MakeDir(0777, cnf.DataDir, "cert")
		stackConfig.FrontEndNginx = utils.MakeDir(0777, cnf.DataDir, "front-end", "nginx")
		stackConfig.Datasources = utils.MakeDir(0777, cnf.DataDir, "datasources")
		stackConfig.EventsEngineWorkdir = utils.MakeDir(0777, cnf.DataDir, "events-engine-workdir")
		stackConfig.ESData = utils.MakeDir(0777, cnf.DataDir, "opensearch", "data")
		stackConfig.ESBackups = utils.MakeDir(0777, cnf.DataDir, "opensearch", "backups")
		stackConfig.LocksDir = utils.MakeDir(0777, cnf.DataDir, "locks")
		stackConfig.ShmFolder = utils.MakeDir(0777, cnf.DataDir, "tmpfs")

		services := []utils.ServiceConfig{
			{Name: "event-processor", Priority: 1, MinMemory: 4 * 1024, MaxMemory: 60 * 1024},
			{Name: "opensearch", Priority: 1, MinMemory: 4350, MaxMemory: 60 * 1024},
			{Name: "backend", Priority: 2, MinMemory: 700, MaxMemory: 2 * 1024},
			{Name: "web-pdf", Priority: 2, MinMemory: 1024, MaxMemory: 2 * 1024},
			{Name: "postgres", Priority: 2, MinMemory: 500, MaxMemory: 2 * 1024},
			{Name: "user-auditor", Priority: 3, MinMemory: 200, MaxMemory: 1024},
			{Name: "agentmanager", Priority: 3, MinMemory: 200, MaxMemory: 1024},
			{Name: "frontend", Priority: 3, MinMemory: 80, MaxMemory: 1024},
		}

		total := int(mem.Total/1024/1024) - utils.SYSTEM_RESERVED_MEMORY

		rsrcs, err := utils.BalanceMemory(services, total)
		if err != nil {
			fmt.Printf("error balancing memory: %v\n", err)
			os.Exit(1)
		}

		stackConfig.ServiceResources = rsrcs
	})

	return stackConfig
}
