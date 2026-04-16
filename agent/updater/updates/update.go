package updates

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/utmstack/UTMStack/agent/updater/config"
	"github.com/utmstack/UTMStack/shared/exec"
	"github.com/utmstack/UTMStack/shared/fs"
	"github.com/utmstack/UTMStack/shared/http"
	"github.com/utmstack/UTMStack/shared/logger"
	"github.com/utmstack/UTMStack/shared/svc"
)

const (
	checkEvery = 5 * time.Minute
)

// Version represents the version info from version.json
type Version struct {
	Version string `json:"version"`
}

// legacyServiceFile returns the old naming convention for the agent binary.
// This is used for migration from old agents that don't have OS/arch suffix.
func legacyServiceFile() string {
	if runtime.GOOS == "windows" {
		return "utmstack_agent_service.exe"
	}
	return "utmstack_agent_service"
}

var currentVersion = Version{}

func UpdateDependencies(cnf *config.Config) {
	basePath := fs.GetExecutablePath()

	if fs.Exists(config.VersionPath) {
		if err := fs.ReadJSON(config.VersionPath, &currentVersion); err != nil {
			logger.Error("error reading version file: %v", err)
		}
	}

	for {
		time.Sleep(checkEvery)

		if err := http.DownloadFile(fmt.Sprintf(config.DependUrl, cnf.Server, config.DependenciesPort, "version.json"), nil, "version_new.json", basePath, cnf.SkipCertValidation); err != nil {
			logger.Error("error downloading version.json: %v", err)
			continue
		}

		newVersion := Version{}
		if err := fs.ReadJSON(filepath.Join(basePath, "version_new.json"), &newVersion); err != nil {
			logger.Error("error reading version file: %v", err)
			continue
		}

		if newVersion.Version != currentVersion.Version {
			logger.Info("New version of agent found: %s", newVersion.Version)

			agentBinary := config.ServiceFile("")
			if err := http.DownloadFile(fmt.Sprintf(config.DependUrl, cnf.Server, config.DependenciesPort, agentBinary), nil, config.ServiceFile("_new"), basePath, cnf.SkipCertValidation); err != nil {
				logger.Error("error downloading agent: %v", err)
				continue
			}

			if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
				if err := exec.Run("chmod", basePath, "-R", "755", filepath.Join(basePath, config.ServiceFile("_new"))); err != nil {
					logger.Error("error executing chmod: %v", err)
				}
			}

			logger.Info("Starting update process...")
			if err := runUpdateProcess(basePath); err != nil {
				logger.Error("error updating service: %v", err)
				os.Remove(filepath.Join(basePath, "version_new.json"))
				os.Remove(filepath.Join(basePath, config.ServiceFile("_new")))
			} else {
				logger.Info("Update completed successfully")
				if fs.Exists(config.VersionPath) {
					if err := fs.ReadJSON(config.VersionPath, &currentVersion); err != nil {
						logger.Error("error reading updated version file: %v", err)
					}
				}
			}
		} else {
			os.Remove(filepath.Join(basePath, "version_new.json"))
		}
	}
}

func runUpdateProcess(basePath string) error {
	newBin := config.ServiceFile("_new")
	oldBin := config.ServiceFile("")
	backupBin := config.ServiceFile(".old")

	agentNew := filepath.Join(basePath, newBin)
	if _, err := os.Stat(agentNew); err != nil {
		return fmt.Errorf("no _new binary found to update")
	}

	if err := svc.Stop(config.SERV_AGENT_NAME); err != nil {
		return fmt.Errorf("error stopping agent: %v", err)
	}

	// Migration: check if old naming convention exists and migrate to new naming
	oldBinPath := filepath.Join(basePath, oldBin)
	if !fs.Exists(oldBinPath) {
		legacyBin := legacyServiceFile()
		legacyBinPath := filepath.Join(basePath, legacyBin)
		if fs.Exists(legacyBinPath) {
			logger.Info("Migrating legacy binary from %s to %s", legacyBin, oldBin)
			if err := os.Rename(legacyBinPath, oldBinPath); err != nil {
				return fmt.Errorf("error migrating legacy binary: %v", err)
			}
		}
	}

	backupPath := filepath.Join(basePath, backupBin)
	if fs.Exists(backupPath) {
		logger.Info("Removing previous backup: %s", backupPath)
		if err := os.Remove(backupPath); err != nil {
			logger.Error("could not remove old backup: %v", err)
		}
	}

	if err := os.Rename(filepath.Join(basePath, oldBin), backupPath); err != nil {
		return fmt.Errorf("error backing up old binary: %v", err)
	}

	if err := os.Rename(filepath.Join(basePath, newBin), filepath.Join(basePath, oldBin)); err != nil {
		os.Rename(backupPath, filepath.Join(basePath, oldBin))
		return fmt.Errorf("error renaming new binary: %v", err)
	}

	if err := svc.Start(config.SERV_AGENT_NAME); err != nil {
		rollbackAgent(oldBin, backupBin, basePath)
		return fmt.Errorf("error starting agent: %v", err)
	}

	time.Sleep(30 * time.Second)

	isHealthy, err := svc.IsActive(config.SERV_AGENT_NAME)
	if err != nil || !isHealthy {
		logger.Info("New version failed health check, rolling back...")
		rollbackAgent(oldBin, backupBin, basePath)
		return fmt.Errorf("rollback completed: new version failed health check")
	}

	logger.Info("Health check passed for agent")

	versionNewPath := filepath.Join(basePath, "version_new.json")
	versionPath := filepath.Join(basePath, "version.json")
	if fs.Exists(versionNewPath) {
		if err := os.Rename(versionNewPath, versionPath); err != nil {
			logger.Error("error updating version file: %v", err)
		} else {
			logger.Info("Version file updated successfully")
		}
	}

	return nil
}

func rollbackAgent(currentBin, backupBin, basePath string) {
	logger.Info("Rolling back agent to previous version...")

	svc.Stop(config.SERV_AGENT_NAME)

	os.Remove(filepath.Join(basePath, currentBin))
	os.Rename(filepath.Join(basePath, backupBin), filepath.Join(basePath, currentBin))

	svc.Start(config.SERV_AGENT_NAME)
	os.Remove(filepath.Join(basePath, "version_new.json"))

	logger.Info("Rollback completed for agent")
}
