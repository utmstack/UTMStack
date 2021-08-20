package utils

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	_ "github.com/lib/pq" //Import PostgreSQL driver
)

var MasterImages = []string{
	"opendistro",
	"openvas",
	"postgres",
	"logstash",
	"nginx",
	"panel",
	"datasources",
	"zapier",
	"correlation",
	"filebrowser",
}

var ProbeImages = []string{
	"openvas",
	"logstash",
	"datasources",
}

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
	if e != nil {
		log.Fatal(e)
	}
}
