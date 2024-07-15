package updates

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/utmstack/UTMStack/agent/installer/config"
	"github.com/utmstack/UTMStack/agent/installer/models"
	"github.com/utmstack/UTMStack/agent/installer/utils"
)

func DownloadDependencies(address, authKey, skip string) error {
	version := models.Version{}
	var err error

	if version.ServiceVersion, err = downloadAndUpdateVersion(address, authKey, skip, config.DependServiceLabel); err != nil {
		return err
	}

	if version.DependenciesVersion, err = downloadAndUpdateVersion(address, authKey, skip, config.DependZipLabel); err != nil {
		return err
	}

	if err := utils.WriteJSON(config.GetVersionPath(), &version); err != nil {
		return fmt.Errorf("error writing version file: %v", err)
	}

	if err := handleDependenciesPostDownload(); err != nil {
		return err
	}

	return nil

}

func downloadAndUpdateVersion(address, authKey, skip, fileType string) (string, error) {
	url := fmt.Sprintf(config.DEPEND_URL, address, "0", runtime.GOOS, fileType)
	var skipTlsVerification bool
	if skip == "yes" {
		skipTlsVerification = true
	}
	resp, status, err := utils.DoReq[models.DependencyUpdateResponse](url, nil, http.MethodGet, map[string]string{"connection-key": authKey}, skipTlsVerification)
	if err != nil {
		return "", fmt.Errorf("error downloading %s: %v", fileType, err)
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("error downloading %s: %v", fileType, resp.Message)
	}

	if err := utils.WriteBytesToFile(config.GetDownloadFilePath(fileType), resp.FileContent); err != nil {
		return "", fmt.Errorf("error writing %s file: %v", fileType, err)
	}

	return resp.Version, nil
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
