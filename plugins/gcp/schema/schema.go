package schema

type PluginConfig struct {
	InternalKey string `yaml:"internal_key"`
	Backend     string `yaml:"backend"`
	LogLevel    int    `yaml:"log_level"`
}

type ModuleConfig struct {
	JsonKey        string
	ProjectID      string
	SubscriptionID string
	Topic          string
}
