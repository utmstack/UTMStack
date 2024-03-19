package update

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/utmstack/UTMStack/agent/self/configuration"
	"github.com/utmstack/UTMStack/agent/self/utils"
)

func UpdateUpdaterService(version, env string) error {
	path, err := utils.GetMyPath()
	if err != nil {
		return fmt.Errorf("failed to get current path: %v", err)
	}

	// Download new bin
	bin := configuration.GetUpdaterBin()
	url := configuration.Bucket + env + "/updater_service/v" + version + "/" + bin + "?time=" + utils.GetCurrentTime()
	err = utils.DownloadFile(url, filepath.Join(path, bin))
	if err != nil {
		return fmt.Errorf("error downloading new %s: %v", bin, err)
	}

	if runtime.GOOS == "linux" {
		if err = utils.Execute("chmod", path, "-R", "777", bin); err != nil {
			return fmt.Errorf("error executing chmod: %v", err)
		}
	}

	return nil
}
