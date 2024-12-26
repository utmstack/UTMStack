package updates

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/utmstack/UTMStack/agent/installer/config"
	"github.com/utmstack/UTMStack/agent/installer/utils"
)

type Version struct {
	Version string `json:"name"`
}

func DownloadDependencies(address, authKey, skip string) error {
	versions := map[string]string{}
	headers := map[string]string{"connection-key": authKey}

	agentVersion, _, err := utils.DoReq[Version](fmt.Sprintf(config.VersionUrl, address, "agent-service"), nil, "GET", headers, skip == "yes")
	if err != nil {
		return fmt.Errorf("error getting agent version: %v", err)
	}

	installerVersion, _, err := utils.DoReq[Version](fmt.Sprintf(config.VersionUrl, address, "agent-installer"), nil, "GET", headers, skip == "yes")
	if err != nil {
		return fmt.Errorf("error getting installer version: %v", err)
	}

	versions["agent-service"] = agentVersion.Version
	versions["agent-installer"] = installerVersion.Version

	dependFiles := config.GetDependFiles()
	for _, file := range dependFiles {
		if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, address, file), headers, file, utils.GetMyPath(), skip == "yes"); err != nil {
			return fmt.Errorf("error downloading file %s: %v", file, err)
		}
	}

	if err := utils.WriteJSON(config.VersionPath, &versions); err != nil {
		return fmt.Errorf("error writing version file: %v", err)
	}

	if err := handleDependenciesPostDownload(dependFiles); err != nil {
		return err
	}

	return nil

}

func handleDependenciesPostDownload(dependencies []string) error {
	for _, file := range dependencies {
		if strings.HasSuffix(file, ".zip") {
			if err := utils.Unzip(filepath.Join(utils.GetMyPath(), file), utils.GetMyPath()); err != nil {
				return fmt.Errorf("error unzipping dependencies: %v", err)
			}

			if runtime.GOOS == "linux" {
				if err := utils.Execute("chmod", utils.GetMyPath(), "-R", "777", config.UpdaterSelfLinux); err != nil {
					utils.Logger.WriteError("error executing chmod: %v", err)
				}
			}

			if err := os.Remove(filepath.Join(utils.GetMyPath(), file)); err != nil {
				utils.Logger.WriteError("error deleting dependencies file", err)
			}
		} else if runtime.GOOS == "linux" {
			if err := utils.Execute("chmod", utils.GetMyPath(), "-R", "777", file); err != nil {
				utils.Logger.WriteError("error executing chmod: %v", err)
			}
		}
	}

	return nil
}
