package domain

type Tenant struct {
	Name   string            `yaml:"name"`
	Config map[string]string `yaml:",inline"`
}
