package updates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/util"
)

type UpdaterInterf interface {
	GetLatestVersion(typ string) string
	GetFileName(osType, typ string) (string, string)
	GetFile(osType, typ string) ([]byte, error)
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
	case "DEPEND_TYPE_INSTALLER":
		if u.LatestVersions.InstallerVersion != "" {
			return u.LatestVersions.InstallerVersion
		}
	case "DEPEND_TYPE_SERVICE":
		if u.LatestVersions.ServiceVersion != "" {
			return u.LatestVersions.ServiceVersion
		}
	case "DEPEND_TYPE_DEPEND_ZIP":
		if u.LatestVersions.DependenciesVersion != "" {
			return u.LatestVersions.DependenciesVersion
		}
	}
	return ""
}

func (u *Updater) GetFileName(osType, typ string) (string, string) {
	file, ok := u.FileLookup[osType][typ]
	if !ok {
		return "", ""
	}
	return fmt.Sprintf(file, strings.ReplaceAll(u.GetLatestVersion(typ), ".", "_")), u.DownloadPath
}

func (u *Updater) GetFile(osType, typ string) ([]byte, error) {
	fileName, _ := u.GetFileName(osType, typ)
	content, err := os.ReadFile(filepath.Join(u.DownloadPath, fileName))
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (u *Updater) UpdateDependencies() {
	h := util.GetLogger()

	for {
		//h.Info("Checking for updates for %s...", u.Name)
		u.updateCycle(h)
		time.Sleep(config.CHECK_EVERY)
	}
}

func (u *Updater) updateCycle(h *logger.Logger) {
	env, err := ReadEnv()
	if err != nil {
		h.ErrorF("Error reading environment: %v", err)
		return
	}

	u.MasterVersion, err = GetMasterVersion()
	if err != nil {
		h.ErrorF("Error getting master version: %v", err)
		return
	}

	err = updateCurrentVersions(u.DownloadPath, &u.CurrentVersions)
	if err != nil {
		h.ErrorF("Error updating current versions: %v", err)
		return
	}

	err = updateLatestVersions(env, u.Name, u.DownloadPath, u.MasterVersion, &u.LatestVersions, h)
	if err != nil {
		h.ErrorF("Error updating latest versions: %v", err)
		return
	}

	err = u.CheckAvailablesUpdates(env, h)
	if err != nil {
		h.ErrorF("Error checking available updates: %v", err)
		return
	}
}

func (u *Updater) CheckAvailablesUpdates(env string, h *logger.Logger) error {
	if u.LatestVersions.ServiceVersion != "" && IsVersionGreater(u.CurrentVersions.ServiceVersion, u.LatestVersions.ServiceVersion) {
		h.Info("Updating new version for %s service...", u.Name)
		windowsFileName, _ := u.GetFileName("OS_WINDOWS", "DEPEND_TYPE_SERVICE")
		linuxFileName, _ := u.GetFileName("OS_LINUX", "DEPEND_TYPE_SERVICE")
		if windowsFileName != "" {
			if err := downloadAndUpdateFile(env, u.Name, windowsFileName, u.DownloadPath, h); err != nil {
				return fmt.Errorf("error updating %s windows service: %v", u.Name, err)
			}
		}
		if linuxFileName != "" {
			if err := downloadAndUpdateFile(env, u.Name, linuxFileName, u.DownloadPath, h); err != nil {
				return fmt.Errorf("error updating %s linux service: %v", u.Name, err)
			}
		}
		h.Info("New version for %s service downloaded.", u.Name)
	}
	if u.LatestVersions.DependenciesVersion != "" && IsVersionGreater(u.CurrentVersions.DependenciesVersion, u.LatestVersions.DependenciesVersion) {
		h.Info("Updating new version for %s dependencies...", u.Name)
		windowsFileName, _ := u.GetFileName("OS_WINDOWS", "DEPEND_TYPE_DEPEND_ZIP")
		linuxFileName, _ := u.GetFileName("OS_LINUX", "DEPEND_TYPE_DEPEND_ZIP")
		if windowsFileName != "" {
			if err := downloadAndUpdateFile(env, u.Name, windowsFileName, u.DownloadPath, h); err != nil {
				return fmt.Errorf("error updating %s windows dependencies: %v", u.Name, err)
			}
		}
		if linuxFileName != "" {
			if err := downloadAndUpdateFile(env, u.Name, linuxFileName, u.DownloadPath, h); err != nil {
				return fmt.Errorf("error updating %s linux dependencies: %v", u.Name, err)
			}
		}
		h.Info("New version for %s dependencies downloaded.", u.Name)
	}
	if u.LatestVersions.InstallerVersion != "" && IsVersionGreater(u.CurrentVersions.InstallerVersion, u.LatestVersions.InstallerVersion) {
		h.Info("Updating new version for %s installer...", u.Name)
		windowsFileName, _ := u.GetFileName("OS_WINDOWS", "DEPEND_TYPE_INSTALLER")
		linuxFileName, _ := u.GetFileName("OS_LINUX", "DEPEND_TYPE_INSTALLER")
		if windowsFileName != "" {
			if err := downloadAndUpdateFile(env, u.Name, windowsFileName, u.DownloadPath, h); err != nil {
				return fmt.Errorf("error updating %s windows installer: %v", u.Name, err)
			}
		}
		if linuxFileName != "" {
			if err := downloadAndUpdateFile(env, u.Name, linuxFileName, u.DownloadPath, h); err != nil {
				return fmt.Errorf("error updating %s linux installer: %v", u.Name, err)
			}
		}
		h.Info("New version for %s installer downloaded.", u.Name)
	}
	return nil

}

func updateCurrentVersions(downloadPath string, currentVersions *models.Version) error {
	versionsFileExists := util.CheckIfPathExist(filepath.Join(downloadPath, config.VersionsFile))
	if !versionsFileExists {
		err := util.WriteJSON(filepath.Join(downloadPath, config.VersionsFile), &currentVersions)
		if err != nil {
			return fmt.Errorf("error writing current versions file %s: %v", filepath.Join(downloadPath, config.VersionsFile), err)
		}
	}

	err := util.ReadJson(filepath.Join(downloadPath, config.VersionsFile), &currentVersions)
	if err != nil {
		return fmt.Errorf("error reading current versions file %s: %v", filepath.Join(downloadPath, config.VersionsFile), err)
	}

	return nil
}

func updateLatestVersions(env string, dependType string, downloadPath string, masterVersion string, latestVersions *models.Version, h *logger.Logger) error {
	err := util.DownloadFile(config.Bucket+env+"/"+dependType+"/"+config.VersionsFile, filepath.Join(downloadPath, config.VersionsFile), h)
	if err != nil {
		return fmt.Errorf("error downloading latest versions file for %s: %v", dependType, err)
	}

	newData := models.DataVersions{}
	err = util.ReadJson(filepath.Join(downloadPath, config.VersionsFile), &newData)
	if err != nil {
		return fmt.Errorf("error reading latest versions file for %s: %v", dependType, err)
	}

	*latestVersions = FindLatestVersion(newData, masterVersion)
	err = util.WriteJSON(filepath.Join(downloadPath, config.VersionsFile), latestVersions)
	if err != nil {
		return fmt.Errorf("error writing latest versions file for %s: %v", dependType, err)
	}

	return nil
}

func downloadAndUpdateFile(env, name, filename, downloadPath string, h *logger.Logger) error {
	url := config.Bucket + env + "/" + name + "/" + filename
	err := util.DownloadFile(url, filepath.Join(downloadPath, filename), h)
	if err != nil {
		return fmt.Errorf("error downloading new %s: %v", filename, err)
	}
	return nil
}
