package updater

import (
	"sync"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

var (
	version     = Release{}
	versionOnce sync.Once
)

func GetVersion() Release {
	versionOnce.Do(func() {
		if !utils.CheckIfPathExist(config.VersionFilePath) {
			err := utils.WriteJSON(config.VersionFilePath, &version)
			if err != nil {
				config.Logger().ErrorF("error writing version file: %v", err)
			}
		} else {
			err := utils.ReadJson(config.VersionFilePath, &version)
			if err != nil {
				config.Logger().ErrorF("error reading version file: %v", err)
			}
		}
	})

	return version
}

func SaveVersion(vers Release) error {
	version = vers
	return utils.WriteJSON(config.VersionFilePath, &version)
}
