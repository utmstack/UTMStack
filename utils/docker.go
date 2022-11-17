package utils

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	ServerType       string
	Lite             bool
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
	ScannerIface     string
	ScannerIP        string
	Correlation      string
	Rules            string
	Tag              string
	Kind             string
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

func InitDocker(mode string, c Config, master bool, tag string, lite bool) error {
	if err := InstallDocker(mode); err != nil {
		return err
	}

	// login to registry
	if err := RunCmd(mode, "docker", "login", "-u", "client", "-p", "4xYkVIAH8kdAH7mP/9BBhbb2ByzLGm4F", "utmstack.azurecr.io"); err != nil {
		return err
	}

	var composerTemplate string

	// pull images from registry
	if master {
		if lite {
			composerTemplate = masterTemplateLite
		} else {
			composerTemplate = masterTemplateStandard
		}
	} else {
		if lite {
			composerTemplate = probeTemplateLite
		} else {
			composerTemplate = probeTemplateStandard
		}
	}

	// generate composer file and deploy
	if err := GenerateFromTemplate(c, composerTemplate, composerFile); err != nil {
		return errors.New("failed to generate compose file: " + err.Error())
	}

	for intent := 0; intent <= 10; intent++ {
		time.Sleep(1 * time.Minute)

		err := RunCmd(mode, "docker-compose", "up", "-d")

		if err != nil && intent <= 9 {
			continue
		} else if err == nil {
			break
		}

		return err
	}

	return nil
}
