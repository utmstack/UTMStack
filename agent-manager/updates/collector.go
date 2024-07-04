package updates

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/utils"
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
		err := utils.CreatePathIfNotExist(filepath.Join(config.VOLPATH, "collector_as400"))
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
					"windows": {
						config.DependServiceLabel: "utmstack_collector_as400_service%s.jar",
						config.DependZipLabel:     "utmstack_collector_as400_dependencies%s_windows.zip",
					},
					"linux": {
						config.DependServiceLabel: "utmstack_collector_as400_service%s.jar",
					},
				},
			},
		}
	})
	return &as400Updater
}

func NewCollectorUpdater() *CollectorUpdater {
	collectorUpdaterOnce.Do(func() {
		err := utils.CreatePathIfNotExist(filepath.Join(config.VOLPATH, "collector_installer"))
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
					"windows": {
						config.DependInstallerLabel: "utmstack_collector_installer%s_windows.exe",
					},
					"linux": {
						config.DependInstallerLabel: "utmstack_collector_installer%s_linux",
					},
				},
			},
		}
	})
	return &collectorUpdater
}
