package updates

import (
	"fmt"
	"os"
	"runtime"

	"github.com/utmstack/UTMStack/agent/installer/config"
	"github.com/utmstack/UTMStack/agent/installer/models"
	"github.com/utmstack/UTMStack/agent/installer/utils"
)

func DownloadDependencies(address, authKey, skip string) error {
	version := models.Version{}
	var err error
	headers := map[string]string{"connection-key": authKey}

	if version.ServiceVersion, err = utils.DownloadFileByChunks(fmt.Sprintf(config.DEPEND_URL, address, "0", runtime.GOOS, config.DependServiceLabel), headers, config.GetDownloadFilePath(config.DependServiceLabel), skip == "yes"); err != nil {
		return fmt.Errorf("error downloading service: %v", err)
	}

	if version.DependenciesVersion, err = utils.DownloadFileByChunks(fmt.Sprintf(config.DEPEND_URL, address, "0", runtime.GOOS, config.DependZipLabel), headers, config.GetDownloadFilePath(config.DependZipLabel), skip == "yes"); err != nil {
		return fmt.Errorf("error downloading dependencies: %v", err)
	}

	if err := utils.WriteJSON(config.GetVersionPath(), &version); err != nil {
		return fmt.Errorf("error writing version file: %v", err)
	}

	if err := handleDependenciesPostDownload(); err != nil {
		return err
	}

	return nil

}

func handleDependenciesPostDownload() error {
	if err := utils.Unzip(config.GetDownloadFilePath(config.DependZipLabel), utils.GetMyPath()); err != nil {
		return fmt.Errorf("error unzipping dependencies: %v", err)
	}

	if runtime.GOOS == "linux" {
		if err := utils.Execute("chmod", utils.GetMyPath(), "-R", "777", config.UpdaterSelfLinux); err != nil {
			utils.Logger.WriteError("error executing chmod: %v", err)
		}
	}

	if err := os.Remove(config.GetDownloadFilePath(config.DependZipLabel)); err != nil {
		utils.Logger.WriteError("error deleting dependencies file", err)
	}

	return nil
}
