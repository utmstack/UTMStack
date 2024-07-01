package updates

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/utmstack/UTMStack/agent/installer/configuration"
	"github.com/utmstack/UTMStack/agent/installer/models"
	"github.com/utmstack/UTMStack/agent/installer/utils"
)

func downloadAndUpdateVersion(address, authKey, fileType string, version *models.Version) error {
	url := fmt.Sprintf(configuration.DEPEND_URL, address, configuration.AgentManagerPort, "0", runtime.GOOS, fileType)
	resp, status, err := utils.DoReq[models.DependencyUpdateResponse](url, nil, http.MethodGet, map[string]string{"connection-key": authKey})
	if err != nil {
		return fmt.Errorf("error downloading %s: %v", fileType, err)
	}
	if status != http.StatusOK {
		return fmt.Errorf("error downloading %s: %v", fileType, resp.Message)
	}

	if fileType == "service" {
		version.ServiceVersion = resp.Version
	} else if fileType == "depend_zip" {
		version.DependenciesVersion = resp.Version
	}

	if err := utils.WriteBytesToFile(configuration.GetDownloadFilePath(fileType), resp.FileContent); err != nil {
		return fmt.Errorf("error writing %s file: %v", fileType, err)
	}

	return nil
}

func DownloadDependencies(address, authKey string) error {
	version := models.Version{}

	if err := downloadAndUpdateVersion(address, authKey, "service", &version); err != nil {
		return err
	}

	if err := downloadAndUpdateVersion(address, authKey, "depend_zip", &version); err != nil {
		return err
	}

	if err := utils.WriteJSON(configuration.GetVersionPath(), &version); err != nil {
		return fmt.Errorf("error writing version file: %v", err)
	}

	if err := handleDependenciesPostDownload(); err != nil {
		return err
	}

	return nil

}

func handleDependenciesPostDownload() error {
	if err := utils.Unzip(configuration.GetDownloadFilePath("dependencies"), utils.GetMyPath()); err != nil {
		return fmt.Errorf("error unzipping dependencies: %v", err)
	}

	if runtime.GOOS == "linux" {
		if err := utils.Execute("chmod", utils.GetMyPath(), "-R", "777", "utmstack_updater_self"); err != nil {
			return fmt.Errorf("error executing chmod: %v", err)
		}
	}

	if err := os.Remove(configuration.GetDownloadFilePath("dependencies")); err != nil {
		utils.GetBeautyLogger().WriteError("error deleting dependencies file", err)
	}

	return nil
}
