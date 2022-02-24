package utils

import (
	"errors"
	"os"
	"time"
)

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

	time.Sleep(90 * time.Second)

	return nil
}

func InitDocker(mode string, env []string, master bool, tag string, lite bool, advAddr string) error {
	if err := InstallDocker(mode); err != nil {
		return err
	}

	// login to registry
	if err := RunCmd(mode, "docker", "login", "-u", "client", "-p", "4xYkVIAH8kdAH7mP/9BBhbb2ByzLGm4F", "utmstack.azurecr.io"); err != nil {
		return err
	}

	if err := RunCmd(mode, "docker", "pull", "containrrr/watchtower"); err != nil {
		return errors.New("failed to pull docker image: containrrr/watchtower")
	}

	var composerTemplate string

	// pull images from registry
	if master {
		if lite {
			for _, image := range MasterLiteImages {
				image = "utmstack.azurecr.io/" + image + ":" + tag
				if err := RunCmd(mode, "docker", "pull", image); err != nil {
					return errors.New("failed to pull docker image: " + image)
				}
			}
			composerTemplate = masterTemplateLite
		} else {
			for _, image := range MasterStandardImages {
				image = "utmstack.azurecr.io/" + image + ":" + tag
				if err := RunCmd(mode, "docker", "pull", image); err != nil {
					return errors.New("failed to pull docker image: " + image)
				}
			}
			composerTemplate = masterTemplateStandard
		}
	} else {
		if lite {
			for _, image := range ProbeImages {
				image = "utmstack.azurecr.io/" + image + ":" + tag
				if err := RunCmd(mode, "docker", "pull", image); err != nil {
					return errors.New("failed to pull docker image: " + image)
				}
			}
			composerTemplate = probeTemplateLite
		} else {
			for _, image := range ProbeStandardImages {
				image = "utmstack.azurecr.io/" + image + ":" + tag
				if err := RunCmd(mode, "docker", "pull", image); err != nil {
					return errors.New("failed to pull docker image: " + image)
				}
			}
			composerTemplate = probeTemplateStandard
		}
	}

	// generate composer file and deploy
	f, err := os.Create(composerFile)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(composerTemplate)

	for i := 1; i <= 3; i++ {
		err := RunEnvCmd(mode, env, "docker-compose", "up", "-d")
		if err == nil {
			break
		} else if i == 3 {
			return err
		} else {
			time.Sleep(5 * time.Second)
		}
	}

	time.Sleep(2 * time.Second)

	return nil
}
