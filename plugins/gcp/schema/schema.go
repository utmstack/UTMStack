package schema

type PluginConfig struct {
	InternalKey string `yaml:"internal_key"`
	Backend     string `yaml:"backend"`
}

type ModuleConfig struct {
	JsonKey        string
	ProjectID      string
	SubscriptionID string
	Topic          string
}
