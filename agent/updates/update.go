package updates

import (
	"fmt"
	"os"
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

		if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, cnf.Server, config.DependenciesPort, "version.json"), map[string]string{}, "version_new.json", utils.GetMyPath(), cnf.SkipCertValidation); err != nil {
			utils.Logger.ErrorF("error downloading version.json: %v", err)
			continue
		}
		newVersion := models.Version{}
		err := utils.ReadJson(filepath.Join(utils.GetMyPath(), "version_new.json"), &newVersion)
		if err != nil {
			utils.Logger.ErrorF("error reading version file: %v", err)
			continue
		}

		if newVersion.UpdaterVersion != currentVersion.UpdaterVersion {
			utils.Logger.Info("New version of updater found: %s", newVersion.UpdaterVersion)
			if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, cnf.Server, config.DependenciesPort, fmt.Sprintf(config.UpdaterFile, "")), map[string]string{}, fmt.Sprintf(config.UpdaterFile, "_new"), utils.GetMyPath(), cnf.SkipCertValidation); err != nil {
				utils.Logger.ErrorF("error downloading updater: %v", err)
				continue
			}

			if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
				if err = utils.Execute("chmod", utils.GetMyPath(), "-R", "755", filepath.Join(utils.GetMyPath(), fmt.Sprintf(config.UpdaterFile, "_new"))); err != nil {
					utils.Logger.ErrorF("error executing chmod: %v", err)
				}
			}

			utils.Logger.Info("Starting updater update process...")
			err = runUpdateProcess()
			if err != nil {
				utils.Logger.ErrorF("error updating updater: %v", err)
				os.Remove(filepath.Join(utils.GetMyPath(), "version_new.json"))
				os.Remove(filepath.Join(utils.GetMyPath(), fmt.Sprintf(config.UpdaterFile, "_new")))
			} else {
				utils.Logger.Info("Updater update completed successfully")
				if utils.CheckIfPathExist(config.VersionPath) {
					err := utils.ReadJson(config.VersionPath, &currentVersion)
					if err != nil {
						utils.Logger.ErrorF("error reading updated version file: %v", err)
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

	newBin := fmt.Sprintf(config.UpdaterFile, "_new")
	oldBin := fmt.Sprintf(config.UpdaterFile, "")
	backupBin := fmt.Sprintf(config.UpdaterFile, ".old")

	updaterNew := filepath.Join(path, newBin)
	if _, err := os.Stat(updaterNew); err != nil {
		return fmt.Errorf("no _new binary found to update")
	}

	if err := utils.StopService(config.SERVICE_UPDATER_NAME); err != nil {
		return fmt.Errorf("error stopping updater: %v", err)
	}

	time.Sleep(10 * time.Second)

	backupPath := filepath.Join(path, backupBin)
	if utils.CheckIfPathExist(backupPath) {
		utils.Logger.Info("Removing previous backup: %s", backupPath)
		if err := os.Remove(backupPath); err != nil {
			utils.Logger.ErrorF("could not remove old backup: %v", err)
		}
	}

	if err := os.Rename(filepath.Join(path, oldBin), backupPath); err != nil {
		return fmt.Errorf("error backing up old binary: %v", err)
	}

	if err := os.Rename(filepath.Join(path, newBin), filepath.Join(path, oldBin)); err != nil {
		os.Rename(backupPath, filepath.Join(path, oldBin))
		return fmt.Errorf("error renaming new binary: %v", err)
	}

	if err := utils.StartService(config.SERVICE_UPDATER_NAME); err != nil {
		rollbackUpdater(oldBin, backupBin, path)
		return fmt.Errorf("error starting updater: %v", err)
	}

	time.Sleep(30 * time.Second)

	isHealthy, err := utils.CheckIfServiceIsActive(config.SERVICE_UPDATER_NAME)
	if err != nil || !isHealthy {
		utils.Logger.Info("New version failed health check, rolling back...")
		rollbackUpdater(oldBin, backupBin, path)
		return fmt.Errorf("rollback completed: new version failed health check")
	}

	utils.Logger.Info("Health check passed for updater")

	versionNewPath := filepath.Join(path, "version_new.json")
	versionPath := filepath.Join(path, "version.json")
	if utils.CheckIfPathExist(versionNewPath) {
		if err := os.Rename(versionNewPath, versionPath); err != nil {
			utils.Logger.ErrorF("error updating version file: %v", err)
		} else {
			utils.Logger.Info("Version file updated successfully")
		}
	}

	return nil
}

func rollbackUpdater(currentBin, backupBin, path string) {
	utils.Logger.Info("Rolling back updater to previous version...")

	utils.StopService(config.SERVICE_UPDATER_NAME)
	time.Sleep(5 * time.Second)

	os.Remove(filepath.Join(path, currentBin))
	os.Rename(filepath.Join(path, backupBin), filepath.Join(path, currentBin))

	utils.StartService(config.SERVICE_UPDATER_NAME)
	os.Remove(filepath.Join(path, "version_new.json"))

	utils.Logger.Info("Rollback completed for updater")
}
