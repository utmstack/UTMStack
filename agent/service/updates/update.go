package updates

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/utmstack/UTMStack/agent/service/beats"
	"github.com/utmstack/UTMStack/agent/service/config"
	"github.com/utmstack/UTMStack/agent/service/models"
	"github.com/utmstack/UTMStack/agent/service/utils"
)

const (
	checkEvery = 5 * time.Minute
)

func UpdateDependencies(cnf *config.Config) {
	RemoveOldServices()

	for {
		time.Sleep(checkEvery)
		versions := models.Version{}
		prepareForUpdate(&versions)

		headers := map[string]string{
			"key":  cnf.AgentKey,
			"id":   fmt.Sprintf("%v", cnf.AgentID),
			"type": "agent",
		}

		// Check for service update
		version, newUpdate, err := utils.DownloadFileByChunks(fmt.Sprintf(config.DEPEND_URL, cnf.Server, versions.ServiceVersion, runtime.GOOS, config.DEPEND_SERVICE_LABEL), headers, config.GetDownloadFilePath(config.DEPEND_SERVICE_LABEL, "_new"), cnf.SkipCertValidation)
		if err != nil {
			utils.Logger.ErrorF("error downloading service: %v", err)
			continue
		}
		if newUpdate {
			versions.ServiceVersion = version
			err = handlePostServDownload(&versions)
			if err != nil {
				utils.Logger.ErrorF("error handling post download service: %v", err)
				continue
			}
		}

		// Check for dependencies update
		version, newUpdate, err = utils.DownloadFileByChunks(fmt.Sprintf(config.DEPEND_URL, cnf.Server, versions.DependenciesVersion, runtime.GOOS, config.DEPEND_ZIP_LABEL), headers, config.GetDownloadFilePath(config.DEPEND_ZIP_LABEL, ""), cnf.SkipCertValidation)
		if err != nil {
			utils.Logger.ErrorF("error downloading dependencies: %v", err)
			continue
		}
		if newUpdate {
			versions.DependenciesVersion = version
			err = handlePostDependDownload(cnf, &versions)
			if err != nil {
				utils.Logger.ErrorF("error handling post download dependencies: %v", err)
				continue
			}
		}
	}
}

func prepareForUpdate(versions *models.Version) {
	isRecentAgentUpgradeDone := !utils.CheckIfPathExist(config.GetVersionPath())
	if isRecentAgentUpgradeDone {
		versions.ServiceVersion = "0.0.0"
		versions.DependenciesVersion = "0.0.0"
		os.RemoveAll(config.GetVersionOldPath())
	} else {
		err := utils.ReadJson(config.GetVersionPath(), versions)
		if err != nil {
			utils.Logger.ErrorF("error reading version file: %v", err)
		}
	}
}

func handlePostServDownload(versions *models.Version) error {
	utils.Logger.Info("New version of agent service found: %s", versions.ServiceVersion)
	err := utils.WriteJSON(config.GetVersionPath(), versions)
	if err != nil {
		return fmt.Errorf("error writing version file: %v", err)
	}
	if runtime.GOOS == "linux" {
		if err = utils.Execute("chmod", utils.GetMyPath(), "-R", "777", config.GetDownloadFilePath(config.DEPEND_SERVICE_LABEL, "_new")); err != nil {
			utils.Logger.ErrorF("error executing chmod: %v", err)
		}
	}
	utils.Execute(config.GetSelfUpdaterPath(), utils.GetMyPath())

	return nil
}

func handlePostDependDownload(cnf *config.Config, versions *models.Version) error {
	utils.Logger.Info("New version of dependencies found: %s", versions.DependenciesVersion)
	if err := beats.UninstallBeats(); err != nil {
		return fmt.Errorf("error uninstalling beats: %v", err)
	}
	err := removeDependencies()
	if err != nil {
		return fmt.Errorf("error removing dependencies: %v", err)
	}
	err = utils.Unzip(config.GetDownloadFilePath(config.DEPEND_ZIP_LABEL, ""), utils.GetMyPath())
	if err != nil {
		return fmt.Errorf("error unzipping dependencies: %v", err)
	}
	if runtime.GOOS == "linux" {
		if err = utils.Execute("chmod", utils.GetMyPath(), "-R", "777", "utmstack_updater_self"); err != nil {
			return fmt.Errorf("error executing chmod: %v", err)
		}
	}
	err = os.Remove(config.GetDownloadFilePath(config.DEPEND_ZIP_LABEL, ""))
	if err != nil {
		utils.Logger.ErrorF("error removing dependencies file: %v", err)
	}
	err = beats.InstallBeats(*cnf)
	if err != nil {
		return fmt.Errorf("error installing beats: %v", err)
	}

	err = utils.WriteJSON(config.GetVersionPath(), &versions)
	if err != nil {
		return fmt.Errorf("error writing version file: %v", err)
	}

	return nil
}

func removeDependencies() error {
	dependPaths := config.GetDependenPaths()
	for _, dep := range dependPaths {
		err := os.RemoveAll(dep)
		if err != nil {
			return fmt.Errorf("error removing file %s: %v", dep, err)
		}
	}
	return nil
}
