package domain

type Framework struct {
	Key         string             `yaml:"key" json:"key"`
	Name        string             `yaml:"name" json:"name"`
	Description string             `yaml:"description,omitempty" json:"description,omitempty"`
	Source      string             `yaml:"source,omitempty" json:"source,omitempty"`
	Sections    []FrameworkSection `yaml:"sections,omitempty" json:"sections,omitempty"`
}

type FrameworkSection struct {
	Key          string        `yaml:"key" json:"key"`
	Name         string        `yaml:"name" json:"name"`
	Requirements []Requirement `yaml:"requirements,omitempty" json:"requirements,omitempty"`
}

type Requirement struct {
	ID          string   `yaml:"id" json:"id"`
	Name        string   `yaml:"name" json:"name"`
	SatisfiedBy []string `yaml:"satisfiedBy,omitempty" json:"satisfiedBy,omitempty"`
}
