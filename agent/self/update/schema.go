package update

type Version struct {
	MasterVersion  string `json:"master_version"`
	AgentVersion   string `json:"agent_version"`
	UpdaterVersion string `json:"updater_version"`
	RedlineVersion string `json:"redline_version"`
}
