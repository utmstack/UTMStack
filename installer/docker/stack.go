package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	sigar "github.com/cloudfoundry/gosigar"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/system"
	"github.com/utmstack/UTMStack/installer/utils"
)

type StackConfig struct {
	FrontEndNginx       string
	ServiceResources    map[string]*system.ServiceConfig
	Threads             int
	ESData              string
	ESBackups           string
	Cert                string
	EventsEngineWorkdir string
	LocksDir            string
	ShmFolder           string
}

var (
	stackConfig     *StackConfig
	stackConfigOnce sync.Once
	Services        = []system.ServiceConfig{}
)

func GetStackConfig() *StackConfig {
	stackConfigOnce.Do(func() {
		cnf := config.GetConfig()

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
		stackConfig.EventsEngineWorkdir = utils.MakeDir(0777, cnf.DataDir, "events-engine-workdir")
		stackConfig.ESData = utils.MakeDir(0777, cnf.DataDir, "opensearch", "data")
		stackConfig.ESBackups = utils.MakeDir(0777, cnf.DataDir, "opensearch", "backups")
		stackConfig.LocksDir = utils.MakeDir(0777, cnf.DataDir, "locks")
		stackConfig.ShmFolder = utils.MakeDir(0777, cnf.DataDir, "tmpfs")

		Services = []system.ServiceConfig{
			{Name: "event-processor", Priority: 1, MinMemory: 4 * 1024, MaxMemory: 60 * 1024},
			{Name: "opensearch", Priority: 1, MinMemory: 4350, MaxMemory: 60 * 1024},
			{Name: "backend", Priority: 2, MinMemory: 700, MaxMemory: 2 * 1024},
			{Name: "postgres", Priority: 2, MinMemory: 500, MaxMemory: 2 * 1024},
			{Name: "agentmanager", Priority: 3, MinMemory: 200, MaxMemory: 1024},
			{Name: "frontend", Priority: 3, MinMemory: 80, MaxMemory: 1024},
		}

		total := int(mem.Total/1024/1024) - system.SYSTEM_RESERVED_MEMORY

		rsrcs, err := system.BalanceMemory(Services, total)
		if err != nil {
			fmt.Printf("error balancing memory: %v\n", err)
			os.Exit(1)
		}

		stackConfig.ServiceResources = rsrcs
	})

	return stackConfig
}

func StackUP(tag string) error {
	var compose = new(Compose)
	err := compose.Populate(config.GetConfig(), GetStackConfig())
	if err != nil {
		return err
	}

	d, err := compose.Encode()
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(utils.GetMyPath(), "compose.yml"), d, 0644)
	if err != nil {
		return err
	}

	if config.ConnectedToInternet {
		fmt.Println("  Downloading images:")
		for _, service := range compose.Services {
			image := strings.ReplaceAll(*service.Image, "${UTMSTACK_TAG}", tag)
			fmt.Printf("    Downloading %s...", image)
			if err := utils.RunCmd("docker", "pull", image); err != nil {
				return err
			}
			fmt.Println(" [OK]")
		}
	} else {
		fmt.Println("  Loading images from local folder:")

		files, err := filepath.Glob(filepath.Join(config.ImagesPath, "*.tar"))
		if err != nil {
			return fmt.Errorf("failed to list tar files: %w", err)
		}

		if len(files) == 0 {
			return fmt.Errorf("no .tar files found in %s", config.ImagesPath)
		}

		for _, tarPath := range files {
			fileName := filepath.Base(tarPath)
			fmt.Printf("    Importing %s...", fileName)

			if err := utils.RunCmd("docker", "load", "-i", tarPath); err != nil {
				return fmt.Errorf("failed to load image from %s: %w", tarPath, err)
			}
			fmt.Println(" [OK]")
		}

	}

	env := []string{"UTMSTACK_TAG=" + tag}
	if err := utils.RunEnvCmd(env, "docker", "stack", "deploy", "--prune", "-c", filepath.Join(utils.GetMyPath(), "compose.yml"), "utmstack"); err != nil {
		return err
	}

	return nil
}
