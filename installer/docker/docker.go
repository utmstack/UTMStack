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

	if err := utils.RunEnvCmd(env, "apt-get", "install", "-y", "ca-certificates", "curl"); err != nil {
		return err
	}

	if err := utils.RunEnvCmd(env, "install", "-m", "0755", "-d", "/etc/apt/keyrings"); err != nil {
		return err
	}

	if err := utils.RunEnvCmd(env, "curl", "-fsSL", "https://download.docker.com/linux/ubuntu/gpg", "-o", "/etc/apt/keyrings/docker.asc"); err != nil {
		return err
	}

	if err := utils.RunEnvCmd(env, "chmod", "a+r", "/etc/apt/keyrings/docker.asc"); err != nil {
		return err
	}

	repoCmd := `echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}") stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null`
	if err := utils.RunEnvCmd(env, "sh", "-c", repoCmd); err != nil {
		return err
	}

	if err := utils.RunEnvCmd(env, "apt-get", "update"); err != nil {
		return err
	}

	if err := utils.RunEnvCmd(env, "apt-get", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io", "docker-compose", "docker-buildx-plugin", "docker-compose-plugin"); err != nil {
		return err
	}

	return nil
}

func InitSwarm(mainIP string) error {
	if err := utils.RunCmd("docker", "swarm", "init", "--advertise-addr", mainIP); err != nil {
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
