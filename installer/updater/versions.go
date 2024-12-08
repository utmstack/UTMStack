package updater

import (
	"sync"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

var (
	versions     = map[string]string{}
	versionsOnce sync.Once
)

func GetVersions() map[string]string {
	versionsOnce.Do(func() {
		if !utils.CheckIfPathExist(config.UpdatesInfoPath) {
			err := utils.WriteJSON(config.UpdatesInfoPath, versions)
			if err != nil {
				config.Logger().ErrorF("error writing versions file: %v", err)
			}
		} else {
			err := utils.ReadYAML(config.UpdatesInfoPath, &versions)
			if err != nil {
				config.Logger().ErrorF("error reading versions file: %v", err)
			}
		}
	})

	return versions
}

func SaveVersions(vers map[string]string) error {
	for k, v := range vers {
		versions[k] = v
	}

	return utils.WriteJSON(config.UpdatesInfoPath, versions)
}
