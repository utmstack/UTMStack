package updates

import "encoding/json"

type InfoResponse struct {
	Display  string          `json:"display-ribbon-on-profiles"`
	Git      json.RawMessage `json:"git"`
	Build    Build           `json:"build"`
	Profiles []string        `json:"activeProfiles"`
}

type Build struct {
	Artifact string `json:"artifact"`
	Name     string `json:"name"`
	Time     string `json:"time"`
	Version  string `json:"version"`
	Group    string `json:"group"`
}

type DataVersions struct {
	Versions []Version `json:"versions"`
}

type Version struct {
	MasterVersion  string `json:"master_version"`
	AgentVersion   string `json:"agent_version"`
	UpdaterVersion string `json:"updater_version"`
	RedlineVersion string `json:"redline_version"`
}
