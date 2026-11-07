//go:build !windows

package utils

import "os/exec"

func setRawCmdLine(cmd *exec.Cmd, c string, arg []string) {}
