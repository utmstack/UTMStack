package updates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/utils"
)

func ManageUpdates() {
	agentUpdater := NewAgentUpdater()
	as400Updater := NewAs400Updater()
	collectorInstallerUpdater := NewCollectorUpdater()

	go agentUpdater.UpdateDependencies()
	go as400Updater.UpdateDependencies()
	go collectorInstallerUpdater.UpdateDependencies()
}

type UpdaterInterf interface {
	GetLatestVersion(typ string) string
	GetFileName(osType, typ string) (string, string)
	GetFileNameWithVersion(osType, typ string) (string, string)
	GetFileContent(osType, typ string) ([]byte, error)
	UpdateDependencies()
}

type Updater struct {
	Name            string
	MasterVersion   string
	DownloadPath    string
	CurrentVersions models.Version
	LatestVersions  models.Version
	FileLookup      map[string]map[string]string
}

func (u *Updater) GetLatestVersion(typ string) string {
	switch typ {
	case config.DependInstallerLabel:
		return u.LatestVersions.InstallerVersion
	case config.DependServiceLabel:
		return u.LatestVersions.ServiceVersion
	case config.DependZipLabel:
		return u.LatestVersions.DependenciesVersion
	}
	return ""
}

func (u *Updater) GetFileNameWithVersion(osType, typ string) (string, string) {
	file, ok := u.FileLookup[osType][typ]
	if !ok {
		return "", ""
	}
	return fmt.Sprintf(file, fmt.Sprintf("_v%s", strings.ReplaceAll(u.GetLatestVersion(typ), ".", "_"))), u.DownloadPath
}

func (u *Updater) GetFileName(osType, typ string) (string, string) {
	file, ok := u.FileLookup[osType][typ]
	if !ok {
		return "", ""
	}
	return fmt.Sprintf(file, ""), u.DownloadPath
}

func (u *Updater) GetFileContent(osType, typ string) ([]byte, error) {
	fileName, _ := u.GetFileName(osType, typ)
	if fileName == "" {
		return nil, fmt.Errorf("file not found")
	}
	content, err := os.ReadFile(filepath.Join(u.DownloadPath, fileName))
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (u *Updater) UpdateDependencies() {
	for {
		env, err := readEnv()
		if err != nil {
			utils.ALogger.ErrorF("error reading environment: %v", err)
			time.Sleep(config.CHECK_EVERY)
			continue
		}

		u.MasterVersion, err = getMasterVersion()
		if err != nil {
			if !strings.Contains(err.Error(), "invalid character '<' looking for beginning of value") {
				utils.ALogger.ErrorF("error getting master version: %v", err)
			}
			time.Sleep(config.CHECK_EVERY)
			continue
		}

		err = updateCurrentVersions(u.DownloadPath, &u.CurrentVersions)
		if err != nil {
			utils.ALogger.ErrorF("error updating current versions: %v", err)
			time.Sleep(config.CHECK_EVERY)
			continue
		}

		err = updateLatestVersions(env, u.Name, u.DownloadPath, u.MasterVersion, &u.LatestVersions, &u.CurrentVersions)
		if err != nil {
			utils.ALogger.ErrorF("error updating latest versions: %v", err)
			time.Sleep(config.CHECK_EVERY)
			continue
		}

		err = u.CheckAvailablesUpdates(env)
		if err != nil {
			utils.ALogger.ErrorF("error checking available updates: %v", err)
			time.Sleep(config.CHECK_EVERY)
			continue
		}

		time.Sleep(config.CHECK_EVERY)
	}
}

func (u *Updater) CheckAvailablesUpdates(env string) error {
	updates := []struct {
		label          string
		currentVersion string
		latestVersion  string
	}{
		{config.DependInstallerLabel, u.CurrentVersions.InstallerVersion, u.LatestVersions.InstallerVersion},
		{config.DependServiceLabel, u.CurrentVersions.ServiceVersion, u.LatestVersions.ServiceVersion},
		{config.DependZipLabel, u.CurrentVersions.DependenciesVersion, u.LatestVersions.DependenciesVersion},
	}

	for _, update := range updates {
		if update.currentVersion != "" && update.latestVersion != "" && isVersionGreater(update.currentVersion, update.latestVersion) {
			utils.ALogger.Info("Updating new version for %s %s...", u.Name, update.label)
			for _, osType := range config.DependOsAllows {
				fileNameWithVersion, _ := u.GetFileNameWithVersion(osType, update.label)
				fileName, _ := u.GetFileName(osType, update.label)
				if fileNameWithVersion != "" && fileName != "" {
					if err := downloadAndUpdateFile(env, u.Name, fileNameWithVersion, filepath.Join(u.DownloadPath, fileName)); err != nil {
						return fmt.Errorf("error updating %s %s: %v", u.Name, update.label, err)
					}
				}
			}
		}
	}

	return nil

}

func updateCurrentVersions(downloadPath string, currentVersions *models.Version) error {
	versionsFileExists := utils.CheckIfPathExist(filepath.Join(downloadPath, config.VersionsFile))
	if !versionsFileExists {
		err := utils.WriteJSON(filepath.Join(downloadPath, config.VersionsFile), &currentVersions)
		if err != nil {
			return fmt.Errorf("error writing current versions file %s: %v", filepath.Join(downloadPath, config.VersionsFile), err)
		}
	}

	err := utils.ReadJson(filepath.Join(downloadPath, config.VersionsFile), &currentVersions)
	if err != nil {
		return fmt.Errorf("error reading current versions file %s: %v", filepath.Join(downloadPath, config.VersionsFile), err)
	}

	return nil
}

func updateLatestVersions(env string, dependType string, downloadPath string, masterVersion string, latestVersions, currentVersions *models.Version) error {
	err := utils.DownloadFile("http://192.168.1.223:8080/"+env+"/"+dependType+"/"+config.VersionsFile, filepath.Join(downloadPath, config.VersionsFile)) // DEBUG Change to config.Bucket
	if err != nil {
		return fmt.Errorf("error downloading latest versions file for %s: %v", dependType, err)
	}

	newData := models.DataVersions{}
	err = utils.ReadJson(filepath.Join(downloadPath, config.VersionsFile), &newData)
	if err != nil {
		return fmt.Errorf("error reading latest versions file for %s: %v", dependType, err)
	}
	copylatestVersions, masterExists := findLatestVersion(newData, masterVersion)
	if masterExists {
		*latestVersions = copylatestVersions
		err = utils.WriteJSON(filepath.Join(downloadPath, config.VersionsFile), latestVersions)
		if err != nil {
			return fmt.Errorf("error writing latest versions file for %s: %v", dependType, err)
		}
	} else {
		*latestVersions = *currentVersions
	}

	return nil
}

func downloadAndUpdateFile(env, name, filename, downloadPath string) error {
	url := config.Bucket + env + "/" + name + "/" + filename
	err := utils.DownloadFile(url, downloadPath)
	if err != nil {
		return fmt.Errorf("error downloading new %s: %v", filename, err)
	}
	return nil
}
