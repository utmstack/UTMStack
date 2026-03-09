package updates

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/utmstack/UTMStack/as400/updater/config"
	"github.com/utmstack/UTMStack/as400/updater/models"
	"github.com/utmstack/UTMStack/as400/updater/utils"
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

		binaryNeedsUpdate := newVersion.Version != currentVersion.Version
		jarNeedsUpdate := newVersion.JarVersion != currentVersion.JarVersion

		if binaryNeedsUpdate || jarNeedsUpdate {
			if binaryNeedsUpdate {
				utils.UpdaterLogger.Info("New version of agent found: %s -> %s", currentVersion.Version, newVersion.Version)
			}
			if jarNeedsUpdate {
				utils.UpdaterLogger.Info("New version of JAR found: %s -> %s", currentVersion.JarVersion, newVersion.JarVersion)
			}

			if binaryNeedsUpdate {
				if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, cnf.Server, config.DependenciesPort, fmt.Sprintf(config.ServiceFile, "")), map[string]string{}, fmt.Sprintf(config.ServiceFile, "_new"), utils.GetMyPath(), cnf.SkipCertValidation); err != nil {
					utils.UpdaterLogger.ErrorF("error downloading agent: %v", err)
					os.Remove(filepath.Join(utils.GetMyPath(), "version_new.json"))
					continue
				}

				if err = utils.Execute("chmod", utils.GetMyPath(), "-R", "755", filepath.Join(utils.GetMyPath(), fmt.Sprintf(config.ServiceFile, "_new"))); err != nil {
					utils.UpdaterLogger.ErrorF("error executing chmod: %v", err)
				}
			}

			if jarNeedsUpdate {
				if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, cnf.Server, config.DependenciesPort, config.JAR_FILE), map[string]string{}, config.JAR_FILE+"_new", utils.GetMyPath(), cnf.SkipCertValidation); err != nil {
					utils.UpdaterLogger.ErrorF("error downloading JAR: %v", err)
					if binaryNeedsUpdate {
						os.Remove(filepath.Join(utils.GetMyPath(), fmt.Sprintf(config.ServiceFile, "_new")))
					}
					os.Remove(filepath.Join(utils.GetMyPath(), "version_new.json"))
					continue
				}
			}

			utils.UpdaterLogger.Info("Starting update process...")
			err = runUpdateProcess(binaryNeedsUpdate, jarNeedsUpdate)
			if err != nil {
				utils.UpdaterLogger.ErrorF("error updating service: %v", err)
				os.Remove(filepath.Join(utils.GetMyPath(), "version_new.json"))
				if binaryNeedsUpdate {
					os.Remove(filepath.Join(utils.GetMyPath(), fmt.Sprintf(config.ServiceFile, "_new")))
				}
				if jarNeedsUpdate {
					os.Remove(filepath.Join(utils.GetMyPath(), config.JAR_FILE+"_new"))
				}
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

func runUpdateProcess(updateBinary, updateJar bool) error {
	path := utils.GetMyPath()

	newBin := fmt.Sprintf(config.ServiceFile, "_new")
	oldBin := fmt.Sprintf(config.ServiceFile, "")
	backupBin := fmt.Sprintf(config.ServiceFile, ".old")

	if updateBinary {
		agentNew := filepath.Join(path, newBin)
		if _, err := os.Stat(agentNew); err != nil {
			return fmt.Errorf("no _new binary found to update")
		}
	}

	if updateJar {
		jarNew := filepath.Join(path, config.JAR_FILE+"_new")
		if _, err := os.Stat(jarNew); err != nil {
			return fmt.Errorf("no _new JAR found to update")
		}
		utils.UpdaterLogger.Info("New JAR file found, will be updated")
	}

	if err := utils.StopService(config.SERV_COLLECTOR_NAME); err != nil {
		return fmt.Errorf("error stopping agent: %v", err)
	}

	time.Sleep(10 * time.Second)

	if updateBinary {
		utils.UpdaterLogger.Info("Updating binary...")
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
		utils.UpdaterLogger.Info("Binary updated successfully")
	}

	if updateJar {
		utils.UpdaterLogger.Info("Updating JAR file...")
		jarBackup := filepath.Join(path, config.JAR_FILE+".old")
		jarCurrent := filepath.Join(path, config.JAR_FILE)

		if utils.CheckIfPathExist(jarBackup) {
			if err := os.Remove(jarBackup); err != nil {
				utils.UpdaterLogger.ErrorF("could not remove old JAR backup: %v", err)
			}
		}

		if utils.CheckIfPathExist(jarCurrent) {
			if err := os.Rename(jarCurrent, jarBackup); err != nil {
				utils.UpdaterLogger.ErrorF("error backing up JAR: %v", err)
			}
		}

		jarNew := filepath.Join(path, config.JAR_FILE+"_new")
		if err := os.Rename(jarNew, jarCurrent); err != nil {
			utils.UpdaterLogger.ErrorF("error installing new JAR: %v", err)
			if utils.CheckIfPathExist(jarBackup) {
				os.Rename(jarBackup, jarCurrent)
			}
		} else {
			utils.UpdaterLogger.Info("JAR file updated successfully")
		}
	}

	if err := utils.StartService(config.SERV_COLLECTOR_NAME); err != nil {
		rollbackAgent(oldBin, backupBin, path, updateBinary, updateJar)
		return fmt.Errorf("error starting agent: %v", err)
	}

	time.Sleep(30 * time.Second)

	isHealthy, err := utils.CheckIfServiceIsActive(config.SERV_COLLECTOR_NAME)
	if err != nil || !isHealthy {
		utils.UpdaterLogger.Info("New version failed health check, rolling back...")
		rollbackAgent(oldBin, backupBin, path, updateBinary, updateJar)
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

func rollbackAgent(currentBin, backupBin, path string, binaryWasUpdated, jarWasUpdated bool) {
	utils.UpdaterLogger.Info("Rolling back agent to previous version...")

	utils.StopService(config.SERV_COLLECTOR_NAME)
	time.Sleep(5 * time.Second)

	if binaryWasUpdated {
		utils.UpdaterLogger.Info("Rolling back binary...")
		os.Remove(filepath.Join(path, currentBin))
		os.Rename(filepath.Join(path, backupBin), filepath.Join(path, currentBin))
		utils.UpdaterLogger.Info("Binary rolled back successfully")
	}

	if jarWasUpdated {
		utils.UpdaterLogger.Info("Rolling back JAR file...")
		jarCurrent := filepath.Join(path, config.JAR_FILE)
		jarBackup := filepath.Join(path, config.JAR_FILE+".old")

		if utils.CheckIfPathExist(jarBackup) {
			os.Remove(jarCurrent)
			os.Rename(jarBackup, jarCurrent)
			utils.UpdaterLogger.Info("JAR file rolled back successfully")
		}
	}

	utils.StartService(config.SERV_COLLECTOR_NAME)
	os.Remove(filepath.Join(path, "version_new.json"))
	if jarWasUpdated {
		os.Remove(filepath.Join(path, config.JAR_FILE+"_new"))
	}
	if binaryWasUpdated {
		os.Remove(filepath.Join(path, fmt.Sprintf(config.ServiceFile, "_new")))
	}

	utils.UpdaterLogger.Info("Rollback completed for agent")
}
