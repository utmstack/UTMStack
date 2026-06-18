package dto

type VersionInfo struct {
	Version    string `json:"version"`
	Edition    string `json:"edition"`
	Changelog  string `json:"changelog,omitempty"`
	Server     string `json:"server,omitempty"`
	InstanceID string `json:"instanceId,omitempty"`
}
