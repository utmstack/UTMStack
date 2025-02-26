package docker

import (
	"fmt"
	"time"

	"github.com/utmstack/UTMStack/installer/services"
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

func PostInstallation() error {
	time.Sleep(3 * time.Minute)

	fmt.Print("  Securing ports 9200, 5432 and 10000")

	if err := utils.RunCmd("docker", "service", "update", "--publish-rm", "9200", "utmstack_node1"); err != nil {
		return err
	}

	if err := utils.RunCmd("docker", "service", "update", "--publish-rm", "5432", "utmstack_postgres"); err != nil {
		return err
	}

	fmt.Println(" [OK]")

	fmt.Print("  Restarting Stack")

	time.Sleep(60 * time.Second)

	if err := utils.RunCmd("systemctl", "restart", "docker"); err != nil {
		return err
	}

	if err := services.Backend(); err != nil {
		return err
	}

	fmt.Println(" [OK]")

	fmt.Print("  Cleaning up Docker system")

	if err := utils.RunCmd("docker", "system", "prune", "-f"); err != nil {
		return err
	}

	fmt.Println(" [OK]")

	return nil
}
