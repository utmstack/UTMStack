package repository

type WidgetDefinition struct {
	Layout map[string]any `yaml:"layout"`
	Spec   map[string]any `yaml:"spec"`
	Config map[string]any `yaml:"config"`
}

type DashboardDefinition struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Widgets     []WidgetDefinition `yaml:"widgets"`
}
