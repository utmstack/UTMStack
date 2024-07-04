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
	agentUpdater     AgentUpdater
	agentUpdaterOnce sync.Once
)

type AgentUpdater struct {
	Updater
}

func NewAgentUpdater() *AgentUpdater {
	agentUpdaterOnce.Do(func() {
		err := utils.CreatePathIfNotExist(filepath.Join(config.VOLPATH, "agent"))
		if err != nil {
			fmt.Printf("Error creating agent path: %v\n", err)
		}
		agentUpdater = AgentUpdater{
			Updater: Updater{
				Name:          "agent",
				MasterVersion: "0.0.0",
				DownloadPath:  filepath.Join(config.VOLPATH, "agent"),
				CurrentVersions: models.Version{
					MasterVersion:       "0.0.0",
					ServiceVersion:      "0.0.0",
					InstallerVersion:    "0.0.0",
					DependenciesVersion: "0.0.0",
				},
				LatestVersions: models.Version{},
				FileLookup: map[string]map[string]string{
					"windows": {
						config.DependInstallerLabel: "utmstack_agent_installer%s_windows.exe",
						config.DependServiceLabel:   "utmstack_agent_service%s_windows.exe",
						config.DependZipLabel:       "utmstack_agent_dependencies%s_windows.zip",
					},
					"linux": {
						config.DependInstallerLabel: "utmstack_agent_installer%s_linux",
						config.DependServiceLabel:   "utmstack_agent_service%s_linux",
						config.DependZipLabel:       "utmstack_agent_dependencies%s_linux.zip",
					},
				},
			},
		}
	})
	return &agentUpdater
}
