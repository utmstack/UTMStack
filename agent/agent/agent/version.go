package agent

import (
	"fmt"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/agent/utils"
)

type Version struct {
	MasterVersion  string `json:"master_version"`
	AgentVersion   string `json:"agent_version"`
	UpdaterVersion string `json:"updater_version"`
	RedlineVersion string `json:"redline_version"`
}

func GetVersion() (string, error) {
	path, err := utils.GetMyPath()
	if err != nil {
		return "", fmt.Errorf("failed to get current path: %v", err)
	}

	versions := &Version{}
	err = utils.ReadJson(filepath.Join(path, "versions.json"), &versions)
	if err != nil {
		return "", fmt.Errorf("error reading current versions.json: %v", err)
	}

	return versions.AgentVersion, nil
}
