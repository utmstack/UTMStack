package config

import "time"

const (
	PluginName  = "com.utmstack.enrichment"
	ProcessName = "plugin_com.utmstack.enrichment"

	CSVSubdir = "csv-datasets"

	MaxCSVRows   = 10000
	WarnCSVRows  = 5000
	PollInterval = 30 * time.Second
)
