package schema

type PluginConfig struct {
	LogInput    string `yaml:"log_input"`
	InternalKey string `yaml:"internal_key"`
	Backend     string `yaml:"backend"`
	CertsFolder string `yaml:"certs_folder"`
	LogLevel    int    `yaml:"log_level"`
}

type ModuleConfig struct {
	JsonKey        string
	ProjectID      string
	SubscriptionID string
	Topic          string
}

type LogEntry struct {
	Log    string `json:"log"`
	Tenant string `json:"tenant"`
}
