package utils

import "time"

func Restart(mode string) {
	if mode == "ui" || mode == "cli" {
		_ = RunCmd(mode, "init", "6")
	} else {
		time.Sleep(10 * time.Second)
		_ = RunCmd(mode, "init", "6")
	}
}
