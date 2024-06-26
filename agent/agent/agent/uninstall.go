package agent

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/utmstack/UTMStack/agent/agent/utils"
)

func UninstallAll() error {
	path := utils.GetMyPath()
	var exeName string
	switch runtime.GOOS {
	case "windows":
		exeName = filepath.Join(path, "utmstack-runner-windows.exe")
	case "linux":
		exeName = filepath.Join(path, "utmstack-runner-linux")
	}

	err := utils.Execute(exeName, path, "uninstall")
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}
