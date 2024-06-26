package updates

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/agent/agent/agent"
	"github.com/utmstack/UTMStack/agent/agent/beats"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/models"
	"github.com/utmstack/UTMStack/agent/agent/utils"
	"google.golang.org/grpc"
)

const (
	checkEvery = 5 * time.Minute
)

func UpdateDependencies(conn *grpc.ClientConn, cnf *configuration.Config, h *logger.Logger) {
	RemoveOldServices()

	for {
		time.Sleep(checkEvery)
		versions := models.Version{}
		prepareForUpdate(h, &versions)

		mapResponses, err := agent.UpdateDependencies(conn, cnf, versions)
		if err != nil {
			h.ErrorF("error checking dependencies: %v", err)
			continue
		}

		err = updateDependencies(mapResponses["dependencies"], cnf, &versions, h)
		if err != nil {
			h.ErrorF("error updating dependencies: %v", err)
			continue
		}

		err = updateService(mapResponses["service"], &versions, h)
		if err != nil {
			h.ErrorF("error updating service: %v", err)
			continue
		}
	}
}

func prepareForUpdate(h *logger.Logger, versions *models.Version) {
	isRecentAgentUpgradeDone := !utils.CheckIfPathExist(configuration.GetVersionPath())
	if isRecentAgentUpgradeDone {
		versions.ServiceVersion = "0.0.0"
		versions.DependenciesVersion = "0.0.0"
		os.RemoveAll(configuration.GetVersionOldPath())
	} else {
		err := utils.ReadJson(configuration.GetVersionPath(), versions)
		if err != nil {
			h.ErrorF("error reading version file: %v", err)
		}
	}
}

func updateService(response *agent.UpdateResponse, versions *models.Version, h *logger.Logger) error {
	newVersion, newUpdate, err := processUpdateResponse(response, configuration.GetDownloadFilePath("service", "_new"), true)
	if err != nil {
		return fmt.Errorf("error processing agent service response: %v", err)
	}

	if newUpdate {
		h.Info("New version of agent service found: %s", newVersion)
		versions.ServiceVersion = newVersion
		err = utils.WriteJSON(configuration.GetVersionPath(), versions)
		if err != nil {
			return fmt.Errorf("error writing version file: %v", err)
		}
		utils.Execute(configuration.GetSelfUpdaterPath(), utils.GetMyPath())
	}
	return nil
}

func updateDependencies(response *agent.UpdateResponse, cnf *configuration.Config, versions *models.Version, h *logger.Logger) error {
	newVersion, newUpdate, err := processUpdateResponse(response, configuration.GetDownloadFilePath("dependencies", ""), true)
	if err != nil {
		return fmt.Errorf("error processing dependencies response: %v", err)
	}

	if newUpdate {
		h.Info("New version of dependencies found: %s", newVersion)
		if err = beats.UninstallBeats(h); err != nil {
			return fmt.Errorf("error uninstalling beats: %v", err)
		}
		err = removeDependencies()
		if err != nil {
			return fmt.Errorf("error removing dependencies: %v", err)
		}
		err = utils.Unzip(configuration.GetDownloadFilePath("dependencies", ""), utils.GetMyPath())
		if err != nil {
			return fmt.Errorf("error unzipping dependencies: %v", err)
		}
		if runtime.GOOS == "linux" {
			if err = utils.Execute("chmod", utils.GetMyPath(), "-R", "777", "utmstack_updater_self"); err != nil {
				return fmt.Errorf("error executing chmod: %v", err)
			}
		}
		err = os.Remove(configuration.GetDownloadFilePath("dependencies", ""))
		if err != nil {
			h.ErrorF("error removing dependencies file: %v", err)
		}
		err = beats.InstallBeats(*cnf, h)
		if err != nil {
			return fmt.Errorf("error installing beats: %v", err)
		}

		versions.DependenciesVersion = newVersion
		err = utils.WriteJSON(configuration.GetVersionPath(), &versions)
		if err != nil {
			return fmt.Errorf("error writing version file: %v", err)
		}
	}
	return nil
}

func processUpdateResponse(updateResponse *agent.UpdateResponse, filepath string, download bool) (string, bool, error) {
	switch {
	case updateResponse.Message == "Dependency not found", strings.Contains(updateResponse.Message, "Error getting dependency file"):
		return updateResponse.Version, false, fmt.Errorf("error getting dependency file: %v", updateResponse.Message)
	case updateResponse.Message == "Dependency already up to date":
		return updateResponse.Version, false, nil
	case strings.Contains(updateResponse.Message, "Dependency update available: v"):
		version := strings.TrimPrefix(updateResponse.Message, "Dependency update available: v")
		if download {
			if err := utils.WriteBytesToFile(filepath, updateResponse.File); err != nil {
				return "", false, fmt.Errorf("error writing dependency file %s: %v", filepath, err)
			}
		}
		return version, true, nil
	default:
		return "", false, fmt.Errorf("error processing update response: %v", updateResponse.Message)
	}
}

func removeDependencies() error {
	dependPaths := configuration.GetDependenPaths()
	for _, dep := range dependPaths {
		err := os.RemoveAll(dep)
		if err != nil {
			return fmt.Errorf("error removing file %s: %v", dep, err)
		}
	}
	return nil
}
