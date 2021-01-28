// +build windows
package main

func installDocker() {
	pshell("Install-PackageProvider -Name NuGet -MinimumVersion 2.8.5.201 -Force")
	pshell("Install-Module DockerProvider -Force")
	pshell("Install-Package Docker -ProviderName DockerProvider -Force")
}

func pshell(command string) error {
	args := []string{"-NoProfile", "-Confirm:$false", "-Command", command}
	return runCmd("powershell.exe", args...)
}
