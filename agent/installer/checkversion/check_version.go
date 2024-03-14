package checkversion

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/agent/runner/agent"
	"github.com/utmstack/UTMStack/agent/runner/utils"
)

func CleanOldVersions(h *logger.Logger) error {
	// Get current path
	path, err := utils.GetMyPath()
	if err != nil {
		return fmt.Errorf("failed to get current path: %v", err)
	}

	// Clean UTMStackAgent v9 version
	agent.CleanOldAgent(h)

	var exeName string
	switch runtime.GOOS {
	case "windows":
		exeName = filepath.Join(path, "utmstackagent-windows.exe")
	case "linux":
		exeName = filepath.Join(path, "utmstackagent-linux")
	}

	// Check if UTMStackAgent is installed
	if isInstalled, err := utils.CheckIfServiceIsInstalled("UTMStackAgent"); err != nil {
		return fmt.Errorf("error checking UTMStackAgent service: %v", err)
	} else if isInstalled {
		if utils.CheckIfPathExist(filepath.Join(path, "version.json")) {
			result, errB := utils.ExecuteWithResult(exeName, path, "uninstall")
			if errB {
				return fmt.Errorf("%s", result)
			}
		} else {
			return fmt.Errorf("UTMStackAgent is already installed")
		}
	}

	return nil
}
