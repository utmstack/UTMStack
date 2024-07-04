package utils

type Config struct {
	RulesFolder   string `yaml:"rules_folder"`
	GeoIPFolder   string `yaml:"geoip_folder"`
	Elasticsearch string `yaml:"elasticsearch"`
	PostgreSQL    struct {
		Server   string `yaml:"server"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"postgresql"`
}
