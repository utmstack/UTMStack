package domain

type RegexPattern struct {
	ID         string
	Definition string
}

type PatternsFile struct {
	Patterns map[string]string `yaml:"patterns"`
}
