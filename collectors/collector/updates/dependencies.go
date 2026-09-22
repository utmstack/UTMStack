package updates

import (
	"fmt"

	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	sharedhttp "github.com/utmstack/UTMStack/shared/http"
	"github.com/utmstack/UTMStack/shared/fs"
)

func DownloadVersion(address string, insecure bool) error {
	url := fmt.Sprintf(config.DependUrl, address, config.DependenciesPort, "version.json")
	if err := sharedhttp.DownloadFile(url, map[string]string{}, "version.json", fs.GetExecutablePath(), insecure); err != nil {
		return fmt.Errorf("error downloading version.json: %v", err)
	}
	return nil
}
