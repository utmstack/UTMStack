package updater

import "time"

type Auth struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

type InstanceDTOOutput struct {
	ID string `json:"id"`
	InstanceDTOInput
	CreatedAt string `json:"created_at"`
}

type InstanceDTOInput struct {
	Name        string `json:"name"`
	Country     string `json:"country"`
	Email       string `json:"email"`
	Version     string `json:"version"`
	Edition     string `json:"edition"`
	CurrentIp   string `json:"current_ip"`
	MappingName string `json:"mapping_name,omitempty"`
}

type InstanceInfo struct {
	Name       string `json:"name"`
	Country    string `json:"country"`
	Email      string `json:"email"`
	Version    string `json:"version"`
	InstanceID string `json:"instance_id"`
}

type UpdateDTO struct {
	ID        string           `json:"id"`
	Instance  InstanceDTOInput `json:"instance,omitempty"`
	Version   VersionDTO       `json:"version"`
	Edition   string           `json:"edition"`
	Sent      bool             `json:"sent"`
	Approved  bool             `json:"approved"`
	ApproveAt time.Time        `json:"aprove_at,omitempty"`
}

type VersionDTO struct {
	Version   string `json:"version"`
	Changelog string `json:"changelog"`
}

type VersionFile struct {
	Version   string `json:"version"`
	Changelog string `json:"changelog"`
	Edition   string `json:"edition"`
}

type LicenseEncrypted struct {
	ExpiresAt   time.Time `json:"expires_at"`
	Datasources int64     `json:"datasources"`
	CPU         int64     `json:"cpu"`
	Memory      int64     `json:"memory"`
	Type        string    `json:"type"`
	CreatedAt   string    `json:"created_at"`
}

type SectionBackend struct {
	ID          int    `json:"id"`
	Section     string `json:"section"`
	Description string `json:"description"`
	ShortName   string `json:"shortName"`
}

type ConfigBackend struct {
	ID                   int            `json:"id"`
	SectionID            int            `json:"sectionId"`
	ConfParamShort       string         `json:"confParamShort"`
	ConfParamLarge       string         `json:"confParamLarge"`
	ConfParamDescription string         `json:"confParamDescription"`
	ConfParamValue       string         `json:"confParamValue"`
	ConfParamRegexp      string         `json:"confParamRegexp"`
	ConfParamRequired    bool           `json:"confParamRequired"`
	ConfParamDatatype    string         `json:"confParamDatatype"`
	ModificationTime     string         `json:"modificationTime"`
	ModificationUser     string         `json:"modificationUser"`
	ConfParamOption      string         `json:"confParamOption"`
	Section              SectionBackend `json:"section"`
}
