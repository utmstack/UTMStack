package utils

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func setRawCmdLine(cmd *exec.Cmd, c string, arg []string) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: strconv.Quote(c) + " " + strings.Join(arg, " ")}
}
