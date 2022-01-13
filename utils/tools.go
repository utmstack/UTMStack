package utils

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/quantfall/holmes"
	_ "github.com/lib/pq" //Import PostgreSQL driver
)

var ProbeImages = []string{
	"logstash",
	"datasources",
	"wazuh-odfe",
}

var MasterImages = []string{
	"opendistro",
	"postgres",
	"utmstack_frontend",
	"utmstack_backend",
	"correlation",
	"filebrowser",
}

var ProbeStandardImages = append(ProbeImages, "openvas")

var MasterLiteImages = append(ProbeImages, MasterImages...)

var MasterStandardImages = append(ProbeStandardImages, MasterImages...)

func RunEnvCmd(mode string, env []string, command string, arg ...string) error {
	cmd := exec.Command(command, arg...)
	cmd.Env = append(os.Environ(), env...)
	if mode == "cli" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		f, _ := os.OpenFile("/var/log/utm-setup.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		defer f.Close()
		cmd.Stdout = f
		cmd.Stderr = f
	}
	return cmd.Run()
}

func RunCmd(mode, command string, arg ...string) error {
	return RunEnvCmd(mode, []string{}, command, arg...)
}

func MakeDir(mode os.FileMode, arg ...string) string {
	path := ""
	for _, folder := range arg {
		path = filepath.Join(path, folder)
		os.MkdirAll(path, mode)
		os.Chmod(path, mode)
	}
	return path
}

// Check Check if error is not nil or exit with error code
func CheckErr(e error) {
	h := holmes.New("debug", "UTMStack")
	if e != nil {
		h.FatalError("%v", e)
	}
}
