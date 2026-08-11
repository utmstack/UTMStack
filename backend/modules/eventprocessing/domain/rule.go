package domain

import "time"

type AdversaryType string

const (
	AdversaryOrigin AdversaryType = "origin"
	AdversaryTarget AdversaryType = "target"
)

func (a AdversaryType) Valid() bool {
	switch a {
	case AdversaryOrigin, AdversaryTarget:
		return true
	default:
		return false
	}
}

type Impact struct {
	Confidentiality int `yaml:"confidentiality"`
	Integrity       int `yaml:"integrity"`
	Availability    int `yaml:"availability"`
}

type Rule struct {
	DataTypes     []string      `yaml:"dataTypes"`
	Name          string        `yaml:"name"`
	Impact        Impact        `yaml:"impact"`
	Category      string        `yaml:"category"`
	Technique     string        `yaml:"technique"`
	Adversary     AdversaryType `yaml:"adversary"`
	References    []any         `yaml:"references,omitempty"`
	Description   string        `yaml:"description,omitempty"`
	Where         string        `yaml:"where"`
	Correlation   any           `yaml:"correlation,omitempty"`
	GroupBy       []string      `yaml:"groupBy,omitempty"`
	DeduplicateBy []string      `yaml:"deduplicateBy,omitempty"`
	TenantId      string        `yaml:"tenantId,omitempty"`
}

type StoredRule struct {
	Rule
	RelPath  string    `yaml:"-"`
	Modified time.Time `yaml:"-"`
	System   bool      `yaml:"-"`
	Enabled  bool      `yaml:"-"`
}
