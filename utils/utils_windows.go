// +build windows

package utils

func installDocker(mode string) {
	pshell(mode, "Install-PackageProvider -Name NuGet -MinimumVersion 2.8.5.201 -Force")
	pshell(mode, "Install-Module DockerProvider -Force")
	pshell(mode, "Install-Package Docker -ProviderName DockerProvider -Force")
}

func pshell(command string) error {
	args := []string{"-NoProfile", "-Confirm:$false", "-Command", command}
	return runCmd(mode, "powershell.exe", args...)
}
