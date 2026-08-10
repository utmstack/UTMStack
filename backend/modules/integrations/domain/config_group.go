package domain

import "github.com/google/uuid"

type TenantConfig struct {
	ID     uuid.UUID     `yaml:"id"`
	Groups []ConfigGroup `yaml:"groups"`
}

type ConfigGroup struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Config      map[string]string `yaml:"config"`
}
