package updater

import (
	"sync"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

var (
	version     = VersionFile{}
	versionOnce sync.Once
)

func GetVersion() VersionFile {
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

func SaveVersion(vers UpdateDTO) error {
	version.Changelog = vers.Version.Changelog
	version.Edition = vers.Instance.Edition
	version.Version = vers.Version.Version

	return utils.WriteJSON(config.VersionFilePath, &version)
}
