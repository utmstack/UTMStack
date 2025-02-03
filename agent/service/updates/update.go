package updates

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/utmstack/UTMStack/agent/service/config"
	"github.com/utmstack/UTMStack/agent/service/utils"
)

const (
	checkEvery = 5 * time.Minute
)

var (
	versions map[string]string
)

type VersionResponse struct {
	Version string `json:"version"`
}

func UpdateDependencies(cnf *config.Config) {
	if utils.CheckIfPathExist(config.VersionPath) {
		err := utils.ReadJson(config.VersionPath, &versions)
		if err != nil {
			utils.Logger.Fatal("error reading version file: %v", err)
		}
	}

	for {
		time.Sleep(checkEvery)

		// Test update agent service
		utils.Logger.Info("Agent Service Updated Successfully")

		headers := map[string]string{
			"key":  cnf.AgentKey,
			"id":   fmt.Sprintf("%v", cnf.AgentID),
			"type": "agent",
		}

		agentVersionResp, _, err := utils.DoReq[VersionResponse](fmt.Sprintf(config.VersionUrl, cnf.Server, "agent-service"), nil, "GET", headers, cnf.SkipCertValidation)
		if err != nil {
			utils.Logger.ErrorF("error getting agent version: %v", err)
			continue
		}

		agentVersion := agentVersionResp.Version
		if agentVersion != versions["agent-service"] {
			utils.Logger.Info("New version of agent service found: %s", agentVersion)
			if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, cnf.Server, config.GetAgentBin("")), headers, config.GetAgentBin("new"), utils.GetMyPath(), cnf.SkipCertValidation); err != nil {
				utils.Logger.ErrorF("error downloading agent: %v", err)
				continue
			}

			newVerions := map[string]string{}
			err = utils.ReadJson(config.VersionPath, &newVerions)
			if err != nil {
				utils.Logger.ErrorF("error reading version file: %v", err)
				continue
			}

			newVerions["agent-service"] = agentVersion
			err = utils.WriteJSON(config.VersionPath, newVerions)
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
