package updates

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/agent/updater/configuration"
	"github.com/utmstack/UTMStack/agent/updater/utils"
)

type UTMServices struct {
	MasterVersion   string
	CurrentVersions Version
	LatestVersions  Version
	ServAttr        map[string]configuration.ServiceAttribt
}

func (u *UTMServices) UpdateCurrentMasterVersion(cnf configuration.Config) error {
	masterVersion, err := getMasterVersion(cnf.Server, cnf.SkipCertValidation)
	if err != nil {
		return fmt.Errorf("error getting master version: %v", err)
	}
	u.MasterVersion = masterVersion
	return nil
}

func (u *UTMServices) UpdateCurrentVersions() error {
	path, err := utils.GetMyPath()
	if err != nil {
		return fmt.Errorf("failed to get current path: %v", err)
	}

	err = utils.ReadJson(filepath.Join(path, "versions.json"), &u.CurrentVersions)
	if err != nil {
		return fmt.Errorf("error reading current versions.json: %v", err)
	}

	return nil
}

func (u *UTMServices) UpdateLatestVersions(env string, utmLogger *logger.Logger) error {
	path, err := utils.GetMyPath()
	if err != nil {
		return fmt.Errorf("failed to get current path: %v", err)
	}

	err = utils.DownloadFile(configuration.Bucket+env+"/versions.json", filepath.Join(path, "versions.json"), utmLogger)
	if err != nil {
		return fmt.Errorf("error downloading latest versions.json: %v", err)
	}

	newData := DataVersions{}
	err = utils.ReadJson(filepath.Join(path, "versions.json"), &newData)
	if err != nil {
		return fmt.Errorf("error reading latest versions.json: %v", err)
	}

	u.LatestVersions = getLatestVersion(newData, u.MasterVersion)
	err = utils.WriteJSON(filepath.Join(path, "versions.json"), &u.LatestVersions)
	if err != nil {
		return fmt.Errorf("error writing versions.json: %v", err)
	}

	return nil
}

func (u *UTMServices) CheckUpdates(env string, utmLogger *logger.Logger) error {
	path, err := utils.GetMyPath()
	if err != nil {
		return fmt.Errorf("failed to get current path: %v", err)
	}

	if isVersionGreater(u.CurrentVersions.RedlineVersion, u.LatestVersions.RedlineVersion) {
		err := u.UpdateService(path, env, "redline", u.LatestVersions.RedlineVersion, utmLogger)
		if err != nil {
			return fmt.Errorf("error updating UTMStackRedline service: %v", err)
		}
		utmLogger.Info("UTMStackRedline service updated correctly")
	}

	if isVersionGreater(u.CurrentVersions.AgentVersion, u.LatestVersions.AgentVersion) {
		err := u.UpdateService(path, env, "agent", u.LatestVersions.AgentVersion, utmLogger)
		if err != nil {
			return fmt.Errorf("error updating UTMStackAgent service: %v", err)
		}
		utmLogger.Info("UTMStackAgent service updated correctly")
	}

	if isVersionGreater(u.CurrentVersions.UpdaterVersion, u.LatestVersions.UpdaterVersion) {
		err := u.UpdateService(path, env, "updater", u.LatestVersions.UpdaterVersion, utmLogger)
		if err != nil {
			return fmt.Errorf("error updating UTMStackUpdater service: %v", err)
		}
		utmLogger.Info("UTMStackUpdater service updated correctly")
	}

	return nil
}

func (u *UTMServices) UpdateService(path string, env string, servCode string, newVers string, utmLogger *logger.Logger) error {
	attr := u.ServAttr[servCode]
	err := utils.CreatePathIfNotExist(filepath.Join(path, "locks"))
	if err != nil {
		return fmt.Errorf("error creating locks path: %v", err)
	}

	err = utils.SetLock(filepath.Join(path, "locks", attr.ServLock))
	if err != nil {
		return fmt.Errorf("error setting lock %s: %v", attr.ServLock, err)
	}

	if servCode == "updater" {
		utils.Execute(filepath.Join(path, attr.ServBin), path)
	} else {
		err = utils.StopService(attr.ServName)
		if err != nil {
			return fmt.Errorf("error stoping %s service: %v", attr.ServName, err)
		}

		url := configuration.Bucket + env + "/" + servCode + "_service/v" + newVers + "/" + attr.ServBin + "?time=" + utils.GetCurrentTime()
		err = utils.DownloadFile(url, filepath.Join(path, attr.ServBin), utmLogger)
		if err != nil {
			return fmt.Errorf("error downloading new %s: %v", attr.ServBin, err)
		}

		if runtime.GOOS == "linux" {
			if err = utils.Execute("chmod", path, "-R", "777", attr.ServBin); err != nil {
				return fmt.Errorf("error executing chmod: %v", err)
			}
		}

		err = utils.RestartService(attr.ServName)
		if err != nil {
			return fmt.Errorf("error restarting %s service: %v", attr.ServName, err)
		}
	}

	err = utils.RemoveLock(filepath.Join(path, "locks", attr.ServLock))
	if err != nil {
		return fmt.Errorf("error removing lock %s: %v", attr.ServLock, err)
	}

	return nil
}

var (
	utmServ     UTMServices
	utmServOnce sync.Once
)

func GetUTMServicesInstance() UTMServices {
	utmServOnce.Do(func() {
		utmServ = UTMServices{
			MasterVersion:   "",
			CurrentVersions: Version{},
			LatestVersions:  Version{},
			ServAttr:        configuration.GetServAttr(),
		}
	})
	return utmServ
}
