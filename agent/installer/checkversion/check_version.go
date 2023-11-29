package checkversion

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/runner/agent"
	"github.com/utmstack/UTMStack/agent/runner/utils"
)

func CleanOldVersions(h *holmes.Logger) error {
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
			err = utils.Execute(exeName, path, "uninstall")
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("UTMStackAgent is already installed")
		}
	}

	return nil
}
