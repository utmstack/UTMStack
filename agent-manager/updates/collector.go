package updates

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/util"
)

var (
	as400Updater         AS400Updater
	as400UpdaterOnce     sync.Once
	collectorUpdater     CollectorUpdater
	collectorUpdaterOnce sync.Once
)

type AS400Updater struct {
	Updater
}

type CollectorUpdater struct {
	Updater
}

func NewAs400Updater() *AS400Updater {
	as400UpdaterOnce.Do(func() {
		err := util.CreatePathIfNotExist(filepath.Join(config.VOLPATH, "collector_as400"))
		if err != nil {
			fmt.Printf("Error creating as400 path: %v\n", err)
		}
		as400Updater = AS400Updater{
			Updater: Updater{
				Name:          "collector_as400",
				MasterVersion: "0.0.0",
				DownloadPath:  filepath.Join(config.VOLPATH, "collector_as400"),
				CurrentVersions: models.Version{
					MasterVersion:       "0.0.0",
					ServiceVersion:      "0.0.0",
					DependenciesVersion: "0.0.0",
				},
				LatestVersions: models.Version{},
				FileLookup: map[string]map[string]string{
					"OS_WINDOWS": {
						"DEPEND_TYPE_SERVICE":    "utmstack_collector_as400_service_v%s.jar",
						"DEPEND_TYPE_DEPEND_ZIP": "utmstack_collector_as400_dependencies_v%s_windows.zip",
					},
					"OS_LINUX": {
						"DEPEND_TYPE_SERVICE":    "utmstack_collector_as400_service_v%s.jar",
						"DEPEND_TYPE_DEPEND_ZIP": "utmstack_collector_as400_dependencies_v%s_linux.zip",
					},
				},
			},
		}
	})
	return &as400Updater
}

func NewCollectorUpdater() *CollectorUpdater {
	collectorUpdaterOnce.Do(func() {
		err := util.CreatePathIfNotExist(filepath.Join(config.VOLPATH, "collector_installer"))
		if err != nil {
			fmt.Printf("Error creating collector path: %v\n", err)
		}
		collectorUpdater = CollectorUpdater{
			Updater: Updater{
				Name:          "collector_installer",
				MasterVersion: "0.0.0",
				DownloadPath:  filepath.Join(config.VOLPATH, "collector_installer"),
				CurrentVersions: models.Version{
					MasterVersion:    "0.0.0",
					InstallerVersion: "0.0.0",
				},
				LatestVersions: models.Version{},
				FileLookup: map[string]map[string]string{
					"OS_WINDOWS": {
						"DEPEND_TYPE_INSTALLER": "utmstack_collector_installer_v%s_windows.exe",
					},
					"OS_LINUX": {
						"DEPEND_TYPE_INSTALLER": "utmstack_collector_installer_v%s_linux",
					},
				},
			},
		}
	})
	return &collectorUpdater
}
