package ransomware

import "runtime"

func runningOnWindows() bool { return runtime.GOOS == "windows" }
