package updates

import (
	"fmt"

	"github.com/utmstack/UTMStack/as400/config"
	"github.com/utmstack/UTMStack/as400/utils"
)

func DownloadVersion(address string, insecure bool) error {
	if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, address, config.DependenciesPort, "version.json"), map[string]string{}, "version.json", utils.GetMyPath(), insecure); err != nil {
		return fmt.Errorf("error downloading version.json : %v", err)
	}

	return nil

}

func DownloadUpdater(address string, insecure bool) error {
	if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, address, config.DependenciesPort, "utmstack_updater_service"), map[string]string{}, "utmstack_updater_service", utils.GetMyPath(), insecure); err != nil {
		return fmt.Errorf("error downloading utmstack_updater_service : %v", err)
	}

	return nil
}

func DownloadJar(address string, insecure bool) error {
	if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, address, config.DependenciesPort, "as400-collector.jar"), map[string]string{}, "as400-collector.jar", utils.GetMyPath(), insecure); err != nil {
		return fmt.Errorf("error downloading as400-collector.jar : %v", err)
	}

	return nil

}
