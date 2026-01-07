package serv

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
)

func CleanOldServices(cnf *config.Config) {
	oldVersion := false

	if oldVersion {
		utils.Logger.Info("old version of agent found, downloading new version")

		if runtime.GOOS != "darwin" {
			if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, cnf.Server, config.DependenciesPort, fmt.Sprintf(config.UpdaterSelf, "")), map[string]string{}, fmt.Sprintf(config.UpdaterSelf, "_new"), utils.GetMyPath(), cnf.SkipCertValidation); err != nil {
				utils.Logger.LogF(100, "error downloading updater: %v", err)
				return
			}
		}

		oldFilePath := filepath.Join(utils.GetMyPath(), fmt.Sprintf(config.UpdaterSelf, ""))
		newFilePath := filepath.Join(utils.GetMyPath(), fmt.Sprintf(config.UpdaterSelf, "_new"))

		utils.Logger.LogF(100, "renaming %s to %s", newFilePath, oldFilePath)
		err := os.Remove(oldFilePath)
		if err != nil {
			utils.Logger.LogF(100, "error removing old updater: %v", err)
		}
		err = os.Rename(newFilePath, oldFilePath)
		if err != nil {
			utils.Logger.LogF(100, "error renaming updater: %v", err)
		}
	} else {
		utils.Logger.LogF(100, "no old version of agent found")
	}
}
