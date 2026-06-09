package updates

import (
	"fmt"

	"github.com/utmstack/UTMStack/collectors/as400/config"
	"github.com/utmstack/UTMStack/collectors/as400/utils"
)

func DownloadVersion(address string, insecure bool) error {
	if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, address, config.DependenciesPort, "version.json"), map[string]string{}, "version.json", utils.GetMyPath(), insecure); err != nil {
		return fmt.Errorf("error downloading version.json : %v", err)
	}

	return nil

}

func DownloadJar(address string, insecure bool) error {
	if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, address, config.DependenciesPort, "as400-collector.jar"), map[string]string{}, "as400-collector.jar", utils.GetMyPath(), insecure); err != nil {
		return fmt.Errorf("error downloading as400-collector.jar : %v", err)
	}

	return nil

}
