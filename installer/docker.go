package main

import (
	"fmt"
	"os"
	"time"

	"github.com/utmstack/UTMStack/installer/types"
	"github.com/utmstack/UTMStack/installer/utils"
)

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

func StackUP(c *types.Config, stack *types.StackConfig) error {
	var compose = new(types.Compose)
	err := compose.Populate(c, stack)
	if err != nil {
		return err
	}

	d, err := compose.Encode()
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
