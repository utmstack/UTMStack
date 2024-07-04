package updates

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/utmstack/UTMStack/collector-installer/config"
	"github.com/utmstack/UTMStack/collector-installer/models"
	"github.com/utmstack/UTMStack/collector-installer/utils"
)

func DownloadAndUpdateVersion(collectorType config.Collector, address, authKey, fileType string) (string, error) {
	url := fmt.Sprintf(config.DEPEND_URL, address, config.AgentManagerPort, "0", runtime.GOOS, fileType, string(collectorType))
	resp, status, err := utils.DoReq[models.DependencyUpdateResponse](url, nil, http.MethodGet, map[string]string{"connection-key": authKey})
	if err != nil {
		return "", fmt.Errorf("error downloading %s: %v", fileType, err)
	}
	if status != http.StatusOK {
		return resp.Version, fmt.Errorf("error downloading %s: %v", fileType, resp.Message)
	}

	if err := utils.WriteBytesToFile(config.GetDownloadFilePath(fileType, collectorType), resp.FileContent); err != nil {
		return resp.Version, fmt.Errorf("error writing %s file: %v", fileType, err)
	}

	return resp.Version, nil
}

func HandleDependenciesPostDownload(collectorType config.Collector) error {
	zipName := config.GetDownloadFilePath(config.DependZipLabel, collectorType)
	if zipName == "" {
		if err := utils.Unzip(zipName, utils.GetMyPath()); err != nil {
			return fmt.Errorf("error unzipping dependencies: %v", err)
		}
		if err := os.Remove(zipName); err != nil {
			fmt.Printf("error deleting dependencies file: %v\n", err)
		}
	}

	return nil
}
