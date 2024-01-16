package depend

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/utmstack/UTMStack/agent/runner/configuration"
	"github.com/utmstack/UTMStack/agent/runner/utils"
)

func DownloadDependencies(servBins ServicesBin, ip string, skip string) error {
	path, err := utils.GetMyPath()
	if err != nil {
		return fmt.Errorf("failed to get current path: %v", err)
	}

	// Select environment
	env, err := configuration.ReadEnv()
	if err != nil {
		return fmt.Errorf("error reading environment configuration: %v", err)
	}

	currentVersion, err := getCurrentVersion(ip, env.Branch, skip)
	if err != nil {
		return fmt.Errorf("error getting current version: %v", err)
	}

	urlFiles := map[string]string{
		"dependencies.zip":         configuration.Bucket + env.Branch + "/agent_service/v" + currentVersion.AgentVersion + "/" + runtime.GOOS + "_dependencies.zip?time=" + utils.GetCurrentTime(),
		servBins.AgentServiceBin:   configuration.Bucket + env.Branch + "/agent_service/v" + currentVersion.AgentVersion + "/" + servBins.AgentServiceBin + "?time=" + utils.GetCurrentTime(),
		servBins.UpdaterServiceBin: configuration.Bucket + env.Branch + "/updater_service/v" + currentVersion.UpdaterVersion + "/" + servBins.UpdaterServiceBin + "?time=" + utils.GetCurrentTime(),
		servBins.RedlineServiceBin: configuration.Bucket + env.Branch + "/redline_service/v" + currentVersion.RedlineVersion + "/" + servBins.RedlineServiceBin + "?time=" + utils.GetCurrentTime(),
	}

	for filename, url := range urlFiles {
		err = utils.DownloadFile(url, filepath.Join(path, filename))
		if err != nil {
			return fmt.Errorf("error downloading dependencies and binaries: %v", err)
		}
		if runtime.GOOS == "linux" {
			if err = utils.Execute("chmod", path, "-R", "777", filename); err != nil {
				return fmt.Errorf("error executing chmod: %v", err)
			}
		}
	}

	err = utils.Unzip(filepath.Join(path, "dependencies.zip"), filepath.Join(path))
	if err != nil {
		return fmt.Errorf("error unzipping dependencies.zip: %v", err)
	}

	if runtime.GOOS == "linux" {
		if err = utils.Execute("chmod", path, "-R", "777", "utmstack_updater_self"); err != nil {
			return fmt.Errorf("error executing chmod: %v", err)
		}
	}

	err = os.Remove(filepath.Join(path, "dependencies.zip"))
	if err != nil {
		log.Printf("error deleting dependencies.zip file: %v\n", err)
	}

	return nil
}

func GetServicesBins() ServicesBin {
	servBins := ServicesBin{}
	switch runtime.GOOS {
	case "windows":
		servBins.AgentServiceBin = "utmstack_agent_service.exe"
		servBins.UpdaterServiceBin = "utmstack_updater_service.exe"
		servBins.RedlineServiceBin = "utmstack_redline_service.exe"
	case "linux":
		servBins.AgentServiceBin = "utmstack_agent_service"
		servBins.UpdaterServiceBin = "utmstack_updater_service"
		servBins.RedlineServiceBin = "utmstack_redline_service"
	}
	return servBins
}
