package main

import (
	"fmt"
	"os"
	"time"

	sigar "github.com/cloudfoundry/gosigar"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/utmstack/UTMStack/installer/utils"
)

type StackConfig struct {
	FrontEndNginx     string
	LogstashPipelines string
	LogstashConfig    string
	ESMem             uint64
	LSMem             uint64
	Threads           int
	ESData            string
	ESBackups         string
	Cert              string
	Datasources       string
	Rules             string
	LocksDir          string
	ShmFolder         string
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
	s.ESMem = mem.Total / 1024 / 1024 / 1024 / 4
	s.LSMem = mem.Total / 1024 / 1024 / 1024 / 5
	s.LocksDir = utils.MakeDir(0777, c.DataDir, "locks")
	s.ShmFolder = utils.MakeDir(0777, c.DataDir, "tmpfs")

	return nil
}

func InstallDocker() error {
	env := []string{"DEBIAN_FRONTEND=noninteractive"}

	if err := utils.RunEnvCmd(env, "apt-get", "update"); err != nil {
		return err
	}

	if err := utils.RunEnvCmd(env, "apt-get", "install", "-y", "apt-transport-https", "ca-certificates", "curl", "gnupg-agent", "software-properties-common"); err != nil {
		return err
	}

	if err := utils.RunEnvCmd(env, "sh", "-c", "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -"); err != nil {
		return err
	}

	if err := utils.RunEnvCmd(env, "sh", "-c", `add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"`); err != nil {
		return err
	}

	if err := utils.RunEnvCmd(env, "apt-get", "update"); err != nil {
		return err
	}

	if err := utils.RunEnvCmd(env, "apt-get", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io", "docker-compose"); err != nil {
		return err
	}

	return nil
}

func InitSwarm(mainIP string) error {
	if err := utils.RunCmd("docker", "swarm", "init", "--advertise-addr", mainIP); err != nil {
		return err
	}

	if err := utils.RunCmd("docker", "login", "-u", "client", "-p", "4xYkVIAH8kdAH7mP/9BBhbb2ByzLGm4F", "utmstack.azurecr.io"); err != nil {
		return err
	}

	return nil
}

func StackUP(c *Config, stack *StackConfig) error {
	var compose = new(Compose)
	d, err := compose.Populate(c, stack).Encode()
	if err != nil {
		return err
	}

	err = os.WriteFile("compose.yml", d, 0644)
	if err != nil {
		return err
	}

	for _, service := range compose.Services {
		fmt.Println("Downloading ", *service.Image)
		if err := utils.RunCmd("docker", "pull", *service.Image); err != nil {
			return err
		}
	}

	if err := utils.RunCmd("docker", "stack", "deploy", "-c", "compose.yml", "utmstack"); err != nil {
		return err
	}

	return nil
}

func PostInstallation() error {
	time.Sleep(3 * time.Minute)

	fmt.Println("Securing ports 9200, 5432 and 10000")

	if err := utils.RunCmd("docker", "service", "update", "--publish-rm", "9200", "utmstack_node1"); err != nil {
		return err
	}

	if err := utils.RunCmd("docker", "service", "update", "--publish-rm", "5432", "utmstack_postgres"); err != nil {
		return err
	}

	if err := utils.RunCmd("docker", "service", "update", "--publish-rm", "10000", "utmstack_correlation"); err != nil {
		return err
	}

	fmt.Println("Securing ports 9200, 5432 and 10000 [OK]")

	fmt.Println("Restarting Stack")

	time.Sleep(60 * time.Second)

	if err := utils.RunCmd("systemctl", "restart", "docker"); err != nil {
		return err
	}

	if err := Backend(); err != nil {
		return err
	}

	fmt.Println("Restarting Stack [OK]")

	fmt.Println("Cleaning up Docker system")

	if err := utils.RunCmd("docker", "system", "prune", "-f"); err != nil {
		return err
	}

	fmt.Println("Cleaning up Docker system [OK]")

	return nil
}
