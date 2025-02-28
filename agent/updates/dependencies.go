package updates

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/models"
	"github.com/utmstack/UTMStack/agent/utils"
)

func DownloadFirstDependencies(address string, authKey string, insecure bool) error {
	headers := map[string]string{"connection-key": authKey}

	version, _, err := utils.DoReq[models.Version](fmt.Sprintf(config.VersionUrl, address, config.DependenciesPort), nil, "GET", headers, insecure)
	if err != nil {
		return fmt.Errorf("error getting agent version: %v", err)
	}

	dependFiles := GetDependFiles()
	for _, file := range dependFiles {
		if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, address, config.DependenciesPort, file), headers, file, utils.GetMyPath(), insecure); err != nil {
			return fmt.Errorf("error downloading file %s: %v", file, err)
		}
	}

	if err := utils.WriteJSON(config.VersionPath, &version); err != nil {
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
					return fmt.Errorf("error executing chmod on %s: %v", config.UpdaterSelfLinux, err)
				}
			}

			if err := os.Remove(filepath.Join(utils.GetMyPath(), file)); err != nil {
				return fmt.Errorf("error removing file %s: %v", file, err)
			}
		} else if runtime.GOOS == "linux" {
			if err := utils.Execute("chmod", utils.GetMyPath(), "-R", "777", file); err != nil {
				return fmt.Errorf("error executing chmod on %s: %v", file, err)
			}
		}
	}

	return nil
}

func GetDependFiles() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{
			"utmstack_agent_dependencies_windows.zip",
		}
	case "linux":
		return []string{
			"utmstack_agent_dependencies_linux.zip",
		}
	}
	return nil
}
