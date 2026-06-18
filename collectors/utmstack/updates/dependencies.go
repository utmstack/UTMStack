package updates

import (
	"fmt"

	"github.com/utmstack/UTMStack/collectors/utmstack/config"
	"github.com/utmstack/UTMStack/collectors/utmstack/utils"
)

func DownloadVersion(address string, insecure bool) error {
	if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, address, config.DependenciesPort, "version.json"), map[string]string{}, "version.json", utils.GetMyPath(), insecure); err != nil {
		return fmt.Errorf("error downloading version.json : %v", err)
	}

	return nil

}
