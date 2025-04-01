package updater

import "time"

type Auth struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

type UpdateDTO struct {
	ID        string      `json:"id"`
	Instance  InstanceDTO `json:"instance,omitempty"`
	Version   VersionDTO  `json:"version"`
	Sent      bool        `json:"sent"`
	Approved  bool        `json:"approved"`
	ApproveAt time.Time   `json:"aprove_at,omitempty"`
}

type InstanceDTO struct {
	ID           string          `json:"id,omitempty"`
	Name         string          `json:"name,omitempty"`
	Organization OrganizationDTO `json:"organization,omitempty"`
	Version      string          `json:"version"`
	Edition      string          `json:"edition"`
	CurrentIp    string          `json:"current_ip"`
	CreatedAt    string          `json:"created_at,omitempty"`
}

type VersionDTO struct {
	ID        string `json:"id,omitempty"`
	Version   string `json:"version"`
	Changelog string `json:"changelog"`
}

type OrganizationDTO struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Country string `json:"country"`
}

type VersionFile struct {
	Version   string `json:"version"`
	Changelog string `json:"changelog"`
	Edition   string `json:"edition"`
}
