package schema

type AWSConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	LogGroupName    string
	LogStreamName   string
}

type PluginConfig struct {
	InternalKey string `yaml:"internal_key"`
	Backend     string `yaml:"backend"`
}
