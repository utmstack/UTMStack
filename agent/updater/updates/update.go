package updates

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/utmstack/UTMStack/agent/updater/config"
	"github.com/utmstack/UTMStack/agent/updater/models"
	"github.com/utmstack/UTMStack/agent/updater/utils"
)

const (
	checkEvery = 5 * time.Minute
)

var currentVersion = models.Version{}

func UpdateDependencies(cnf *config.Config) {
	if utils.CheckIfPathExist(config.VersionPath) {
		err := utils.ReadJson(config.VersionPath, &currentVersion)
		if err != nil {
			utils.UpdaterLogger.ErrorF("error reading version file: %v", err)
		}
	}

	for {
		time.Sleep(checkEvery)

		if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, cnf.Server, config.DependenciesPort, "version.json"), map[string]string{}, "version_new.json", utils.GetMyPath(), cnf.SkipCertValidation); err != nil {
			utils.UpdaterLogger.ErrorF("error downloading version.json: %v", err)
			continue
		}
		newVersion := models.Version{}
		err := utils.ReadJson(filepath.Join(utils.GetMyPath(), "version_new.json"), &newVersion)
		if err != nil {
			utils.UpdaterLogger.ErrorF("error reading version file: %v", err)
			continue
		}

		if newVersion.Version != currentVersion.Version {
			utils.UpdaterLogger.Info("New version of agent found: %s", newVersion.Version)
			if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, cnf.Server, config.DependenciesPort, fmt.Sprintf(config.ServiceFile, "")), map[string]string{}, fmt.Sprintf(config.ServiceFile, "_new"), utils.GetMyPath(), cnf.SkipCertValidation); err != nil {
				utils.UpdaterLogger.ErrorF("error downloading agent: %v", err)
				continue
			}

			if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
				if err = utils.Execute("chmod", utils.GetMyPath(), "-R", "755", filepath.Join(utils.GetMyPath(), fmt.Sprintf(config.ServiceFile, "_new"))); err != nil {
					utils.UpdaterLogger.ErrorF("error executing chmod: %v", err)
				}
			}

			utils.UpdaterLogger.Info("Starting update process...")
			err = runUpdateProcess()
			if err != nil {
				utils.UpdaterLogger.ErrorF("error updating service: %v", err)
				os.Remove(filepath.Join(utils.GetMyPath(), "version_new.json"))
				os.Remove(filepath.Join(utils.GetMyPath(), fmt.Sprintf(config.ServiceFile, "_new")))
			} else {
				utils.UpdaterLogger.Info("Update completed successfully")
				if utils.CheckIfPathExist(config.VersionPath) {
					err := utils.ReadJson(config.VersionPath, &currentVersion)
					if err != nil {
						utils.UpdaterLogger.ErrorF("error reading updated version file: %v", err)
					}
				}
			}
		} else {
			os.Remove(filepath.Join(utils.GetMyPath(), "version_new.json"))
		}
	}
}

func runUpdateProcess() error {
	path := utils.GetMyPath()

	newBin := fmt.Sprintf(config.ServiceFile, "_new")
	oldBin := fmt.Sprintf(config.ServiceFile, "")
	backupBin := fmt.Sprintf(config.ServiceFile, ".old")

	agentNew := filepath.Join(path, newBin)
	if _, err := os.Stat(agentNew); err != nil {
		return fmt.Errorf("no _new binary found to update")
	}

	if err := utils.StopService(config.SERV_AGENT_NAME); err != nil {
		return fmt.Errorf("error stopping agent: %v", err)
	}

	time.Sleep(10 * time.Second)

	backupPath := filepath.Join(path, backupBin)
	if utils.CheckIfPathExist(backupPath) {
		utils.UpdaterLogger.Info("Removing previous backup: %s", backupPath)
		if err := os.Remove(backupPath); err != nil {
			utils.UpdaterLogger.ErrorF("could not remove old backup: %v", err)
		}
	}

	if err := os.Rename(filepath.Join(path, oldBin), backupPath); err != nil {
		return fmt.Errorf("error backing up old binary: %v", err)
	}

	if err := os.Rename(filepath.Join(path, newBin), filepath.Join(path, oldBin)); err != nil {
		os.Rename(backupPath, filepath.Join(path, oldBin))
		return fmt.Errorf("error renaming new binary: %v", err)
	}

	if err := utils.StartService(config.SERV_AGENT_NAME); err != nil {
		rollbackAgent(oldBin, backupBin, path)
		return fmt.Errorf("error starting agent: %v", err)
	}

	time.Sleep(30 * time.Second)

	isHealthy, err := utils.CheckIfServiceIsActive(config.SERV_AGENT_NAME)
	if err != nil || !isHealthy {
		utils.UpdaterLogger.Info("New version failed health check, rolling back...")
		rollbackAgent(oldBin, backupBin, path)
		return fmt.Errorf("rollback completed: new version failed health check")
	}

	utils.UpdaterLogger.Info("Health check passed for agent")

	versionNewPath := filepath.Join(path, "version_new.json")
	versionPath := filepath.Join(path, "version.json")
	if utils.CheckIfPathExist(versionNewPath) {
		if err := os.Rename(versionNewPath, versionPath); err != nil {
			utils.UpdaterLogger.ErrorF("error updating version file: %v", err)
		} else {
			utils.UpdaterLogger.Info("Version file updated successfully")
		}
	}

	return nil
}

func rollbackAgent(currentBin, backupBin, path string) {
	utils.UpdaterLogger.Info("Rolling back agent to previous version...")

	utils.StopService(config.SERV_AGENT_NAME)
	time.Sleep(5 * time.Second)

	os.Remove(filepath.Join(path, currentBin))
	os.Rename(filepath.Join(path, backupBin), filepath.Join(path, currentBin))

	utils.StartService(config.SERV_AGENT_NAME)
	os.Remove(filepath.Join(path, "version_new.json"))

	utils.UpdaterLogger.Info("Rollback completed for agent")
}
