// +build linux

package utils

import (
	"time"
	"os"
)

func installDocker(mode string) {
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

	for _, config := range sysctl{
		if err := runCmd(mode, "sysctl", "-w", config); err != nil {
			return errors.New("Failed to set sysctl config")
		}
		f.WriteString(config+"\n")
	}

	env := []string{"DEBIAN_FRONTEND=noninteractive"}
	runEnvCmd(mode, env, "apt","update")
	runEnvCmd(mode, env, "apt", "install", "-y", "apt-transport-https", "ca-certificates", "curl", "gnupg-agent", "software-properties-common")

	runEnvCmd(mode, env, "sh", "-c", "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -")
	runEnvCmd(mode, env, "sh", "-c", `add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"`)

	runEnvCmd(mode, env,"apt","update")
	runEnvCmd(mode, env,"apt","install", "-y", "docker-ce", "docker-ce-cli", "containerd.io")

	time.Sleep(20 * time.Second)
}
