// +build linux

package utils

func installDocker(mode string) {
	aptGet(mode, "update")
	aptGet(mode, "install", "-y", "apt-transport-https", "ca-certificates", "curl", "gnupg-agent", "software-properties-common")

	env := []string{"DEBIAN_FRONTEND=noninteractive"}
	runEnvCmd(mode, env, "sh", "-c", "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -")
	runEnvCmd(mode, env, "sh", "-c", `add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"`)

	aptGet(mode, "update")
	aptGet(mode, "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io")
}

func aptGet(mode, arg ...string) error {
	env := []string{"DEBIAN_FRONTEND=noninteractive"}
	return runEnvCmd(mode, env, "apt", arg...)
}
