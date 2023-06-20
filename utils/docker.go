package utils

import (
	"errors"
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

func InstallDocker() error {
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
		if err := RunCmd("sysctl", "-w", config); err != nil {
			return errors.New("failed to set sysctl config")
		}
		f.WriteString(config + "\n")
	}

	env := []string{"DEBIAN_FRONTEND=noninteractive"}

	if err := RunEnvCmd(env, "apt", "update"); err != nil {
		return err
	}

	if err := RunEnvCmd(env, "apt", "install", "-y", "apt-transport-https", "ca-certificates", "curl", "gnupg-agent", "software-properties-common"); err != nil {
		return err
	}

	if err := RunEnvCmd(env, "sh", "-c", "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -"); err != nil {
		return err
	}

	if err := RunEnvCmd(env, "sh", "-c", `add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"`); err != nil {
		return err
	}

	if err := RunEnvCmd(env, "apt", "update"); err != nil {
		return err
	}

	if err := RunEnvCmd(env, "apt", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io", "docker-compose"); err != nil {
		return err
	}

	return nil
}

func InitSwarm(mainIP string) error {
	if err := RunCmd("docker", "swarm", "init", "--advertise-addr", mainIP); err != nil {
		return err
	}

	if err := RunCmd("docker", "login", "-u", "client", "-p", "4xYkVIAH8kdAH7mP/9BBhbb2ByzLGm4F", "utmstack.azurecr.io"); err != nil {
		return err
	}

	return nil
}
