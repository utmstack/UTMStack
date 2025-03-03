package updates

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/models"
	"github.com/utmstack/UTMStack/agent/utils"
)

const (
	checkEvery = 5 * time.Minute
)

var currentVersion = models.Version{}

func UpdateDependencies(cnf *config.Config) {
	if utils.CheckIfPathExist(config.VersionPath) {
		err := utils.ReadJson(config.VersionPath, &currentVersion)
		if err != nil {
			utils.Logger.Fatal("error reading version file: %v", err)
		}
	}

	for {
		time.Sleep(checkEvery)

		headers := map[string]string{
			"key":  cnf.AgentKey,
			"id":   fmt.Sprintf("%v", cnf.AgentID),
			"type": "agent",
		}

		newVersion, _, err := utils.DoReq[models.Version](fmt.Sprintf(config.VersionUrl, cnf.Server), nil, "GET", headers, cnf.SkipCertValidation)
		if err != nil {
			utils.Logger.ErrorF("error getting agent version: %v", err)
			continue
		}

		if newVersion.Version != currentVersion.Version {
			utils.Logger.Info("New version of agent found: %s", newVersion.Version)
			if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, cnf.Server, config.GetAgentBin("")), headers, config.GetAgentBin("new"), utils.GetMyPath(), cnf.SkipCertValidation); err != nil {
				utils.Logger.ErrorF("error downloading agent: %v", err)
				continue
			}

			currentVersion = newVersion
			err = utils.WriteJSON(config.VersionPath, &currentVersion)
			if err != nil {
				utils.Logger.ErrorF("error writing version file: %v", err)
				continue
			}

			if runtime.GOOS == "linux" {
				if err = utils.Execute("chmod", utils.GetMyPath(), "-R", "777", filepath.Join(utils.GetMyPath(), config.GetAgentBin("_new"))); err != nil {
					utils.Logger.ErrorF("error executing chmod: %v", err)
				}
			}

			utils.Execute(config.GetSelfUpdaterPath(), utils.GetMyPath())
		}
	}
}
