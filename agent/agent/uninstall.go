package agent

import (
	"fmt"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
)

func UninstallAll() error {
	err := utils.Execute(filepath.Join(utils.GetMyPath(), fmt.Sprintf(config.ServiceFile, "")), utils.GetMyPath(), "uninstall")
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}
