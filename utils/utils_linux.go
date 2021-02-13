// +build linux

package utils

import "time"

func installDocker(mode string) {
	env := []string{"DEBIAN_FRONTEND=noninteractive"}
	runEnvCmd(mode, env, "apt","update")
	runEnvCmd(mode, env, "apt", "install", "-y", "apt-transport-https", "ca-certificates", "curl", "gnupg-agent", "software-properties-common")

	runEnvCmd(mode, env, "sh", "-c", "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -")
	runEnvCmd(mode, env, "sh", "-c", `add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"`)

	runEnvCmd(mode, env,"apt","update")
	runEnvCmd(mode, env,"apt","install", "-y", "docker-ce", "docker-ce-cli", "containerd.io")

	time.Sleep(20 * time.Second)
}
