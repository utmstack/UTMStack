package updates

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
)

func DownloadFirstDependencies(address string, authKey string, insecure bool) error {
	headers := map[string]string{"connection-key": authKey}

	if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, address, config.DependenciesPort, "version.json"), headers, "version.json", utils.GetMyPath(), insecure); err != nil {
		return fmt.Errorf("error downloading version.json : %v", err)
	}

	dependFiles := config.DependFiles
	for _, file := range dependFiles {
		if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, address, config.DependenciesPort, file), headers, file, utils.GetMyPath(), insecure); err != nil {
			return fmt.Errorf("error downloading file %s: %v", file, err)
		}
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
				if err := utils.Execute("chmod", utils.GetMyPath(), "-R", "777", fmt.Sprintf(config.UpdaterSelf, "")); err != nil {
					return fmt.Errorf("error executing chmod on %s: %v", fmt.Sprintf(config.UpdaterSelf, ""), err)
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
