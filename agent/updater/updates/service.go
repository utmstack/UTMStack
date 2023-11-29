package updates

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/updater/configuration"
	"github.com/utmstack/UTMStack/agent/updater/constants"
	"github.com/utmstack/UTMStack/agent/updater/utils"
)

type UTMServices struct {
	CurrentVersions Version
	LatestVersions  Version
	ServAttr        map[string]constants.ServiceAttribt
}

func (u *UTMServices) UpdateCurrentVersions() error {
	path, err := utils.GetMyPath()
	if err != nil {
		return fmt.Errorf("failed to get current path: %v", err)
	}

	// Save data from versions.json
	err = utils.ReadJson(filepath.Join(path, "versions.json"), &u.CurrentVersions)
	if err != nil {
		return fmt.Errorf("error reading current versions.json: %v", err)
	}

	return nil
}

func (u *UTMServices) UpdateLatestVersions() error {
	// Select environment
	env, err := configuration.ReadEnv()
	if err != nil {
		return fmt.Errorf("error reading environment configuration: %v", err)
	}

	path, err := utils.GetMyPath()
	if err != nil {
		return fmt.Errorf("failed to get current path: %v", err)
	}

	err = utils.DownloadFile(constants.Bucket+env.Branch+"/versions.json?time="+utils.GetCurrentTime(), filepath.Join(path, "versions.json"))
	if err != nil {
		return fmt.Errorf("error downloading latest versions.json: %v", err)
	}

	// Save data from versions.json
	newData := DataVersions{}
	err = utils.ReadJson(filepath.Join(path, "versions.json"), &newData)
	if err != nil {
		return fmt.Errorf("error reading latest versions.json: %v", err)
	}

	// Save master versions
	u.LatestVersions = GetLatestVersion(newData, u.CurrentVersions.MasterVersion)
	err = utils.WriteJSON(filepath.Join(path, "versions.json"), &u.LatestVersions)
	if err != nil {
		return fmt.Errorf("error writing versions.json: %v", err)
	}

	return nil
}

func (u *UTMServices) CheckUpdates(masterVersion string, h *holmes.Logger) error {
	if isNewOrEqualVersion(u.LatestVersions.MasterVersion, masterVersion) {
		path, err := utils.GetMyPath()
		if err != nil {
			return fmt.Errorf("failed to get current path: %v", err)
		}

		// Select environment
		env, err := configuration.ReadEnv()
		if err != nil {
			return fmt.Errorf("error reading environment configuration: %v", err)
		}

		if isVersionGreater(u.CurrentVersions.RedlineVersion, u.LatestVersions.RedlineVersion) {
			err := u.UpdateService(path, env.Branch, "redline")
			if err != nil {
				return fmt.Errorf("error updating UTMStackRedline service: %v", err)
			}
			h.Info("UTMStackRedline service updated correctly")
		}

		if isVersionGreater(u.CurrentVersions.AgentVersion, u.LatestVersions.AgentVersion) {
			err := u.UpdateService(path, env.Branch, "agent")
			if err != nil {
				return fmt.Errorf("error updating UTMStackAgent service: %v", err)
			}
			h.Info("UTMStackAgent service updated correctly")
		}

		if isVersionGreater(u.CurrentVersions.UpdaterVersion, u.LatestVersions.UpdaterVersion) {
			err := u.UpdateService(path, env.Branch, "updater")
			if err != nil {
				return fmt.Errorf("error updating UTMStackUpdater service: %v", err)
			}
			h.Info("UTMStackUpdater service updated correctly")
		}
	}

	return nil
}

func (u *UTMServices) UpdateService(path string, env string, servCode string) error {
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
		// Stop service
		err = utils.StopService(attr.ServName)
		if err != nil {
			return fmt.Errorf("error stoping %s service: %v", attr.ServName, err)
		}

		// Download new version
		vers := ""
		switch servCode {
		case "agent":
			vers = u.LatestVersions.AgentVersion
		case "redline":
			vers = u.LatestVersions.RedlineVersion
		}

		url := constants.Bucket + env + "/" + servCode + "_service/v" + vers + "/" + attr.ServBin + "?time=" + utils.GetCurrentTime()
		err = utils.DownloadFile(url, filepath.Join(path, attr.ServBin))
		if err != nil {
			return fmt.Errorf("error downloading new %s: %v", attr.ServBin, err)
		}

		if runtime.GOOS == "linux" {
			if err = utils.Execute("chmod", path, "-R", "777", attr.ServBin); err != nil {
				return fmt.Errorf("error executing chmod: %v", err)
			}
		}

		// Restart service
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
			CurrentVersions: Version{},
			LatestVersions:  Version{},
			ServAttr:        constants.GetServAttr(),
		}
	})
	return utmServ
}
