package utils

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	ServerType       string
	ServerName       string
	DBHost           string
	DBPass           string
	LogstashPipeline string
	ESMem            uint64
	LSMem            uint64
	Updates          uint32
	ESData           string
	ESBackups        string
	Cert             string
	Datasources      string
	Correlation      string
	Rules            string
	Tag              string
}

func InstallDocker(mode string) error {
	// set map_max_count size to 262144 and disable IPv6
	sysctl := []string{
		"vm.max_map_count=262144",
		"net.ipv6.conf.all.disable_ipv6=1",
		"net.ipv6.conf.default.disable_ipv6=1",
	}
	f, err := os.OpenFile("/etc/sysctl.conf", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, config := range sysctl {
		if err := RunCmd(mode, "sysctl", "-w", config); err != nil {
			return errors.New("failed to set sysctl config")
		}
		f.WriteString(config + "\n")
	}

	env := []string{"DEBIAN_FRONTEND=noninteractive"}

	if err := RunEnvCmd(mode, env, "apt", "update"); err != nil {
		return err
	}

	if err := RunEnvCmd(mode, env, "apt", "install", "-y", "apt-transport-https", "ca-certificates", "curl", "gnupg-agent", "software-properties-common"); err != nil {
		return err
	}

	if err := RunEnvCmd(mode, env, "sh", "-c", "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -"); err != nil {
		return err
	}

	if err := RunEnvCmd(mode, env, "sh", "-c", `add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"`); err != nil {
		return err
	}

	if err := RunEnvCmd(mode, env, "apt", "update"); err != nil {
		return err
	}

	if err := RunEnvCmd(mode, env, "apt", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io", "docker-compose"); err != nil {
		return err
	}

	return nil
}

func InitDocker(mode string, c Config, master bool, tag, mainIP string) error {
	if err := InstallDocker(mode); err != nil {
		return err
	}

	// init swarm
	if err := RunCmd(mode, "docker", "swarm", "init", "--advertise-addr", mainIP); err != nil {
		return err
	}

	// login to registry
	if err := RunCmd(mode, "docker", "login", "-u", "client", "-p", "4xYkVIAH8kdAH7mP/9BBhbb2ByzLGm4F", "utmstack.azurecr.io"); err != nil {
		return err
	}

	var composerTemplate string

	if master {
		composerTemplate = masterTemplateStandard
	} else {
		composerTemplate = probeTemplateStandard
	}

	// generate composer file and deploy
	if err := GenerateFromTemplate(c, composerTemplate, composerFile); err != nil {
		return errors.New("failed to generate compose file: " + err.Error())
	}

	// deploy stack
	if err := RunCmd(mode, "docker", "stack", "deploy", "-c", composerFile, tag); err != nil {
		return fmt.Errorf("failed to deploy stack: %s", err.Error())
	}

	return nil
}
