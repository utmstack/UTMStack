// +build linux

package main

import "os"

func installDocker() {
	aptGet("update")
	aptGet("install", "apt-transport-https", "ca-certificates", "curl", "gnupg-agent", "software-properties-common")

	env := []string{"DEBIAN_FRONTEND=noninteractive"}
	runEnvCmd(env, "sh", "-c", "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -")
	runEnvCmd(env, "sh", "-c", `add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"`)

	aptGet("update")
	aptGet("install", "docker-ce", "docker-ce-cli", "containerd.io")

	// set map_max_count size to 262144
	check(runCmd("sysctl", "-w", "vm.max_map_count=262144"))
	f, err := os.OpenFile("/etc/sysctl.conf", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	check(err)
	defer f.Close()
	f.WriteString("vm.max_map_count=262144")
}

func aptGet(command string, arg ...string) error {
	env := []string{"DEBIAN_FRONTEND=noninteractive"}
	return runEnvCmd(env, "apt-get", append([]string{"--yes"}, arg...)...)
}
