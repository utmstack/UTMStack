package utils

import "time"

func Restart(mode string) {
	select {
	case <-time.After(10 * time.Second):
		_ = RunCmd(mode, "init", "6")
	}
}
