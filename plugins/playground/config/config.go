package config

import (
	"time"

	"github.com/threatwinds/go-sdk/plugins"
)

const (
	HTTPPort          = ":8091"
	WorkDir           = "/playground-workdir"
	ProductionWorkDir = "/workdir"
	PlaygroundBinary  = "/usr/local/bin/playground"
	WatchRestartDelay = 5 * time.Second
	EventTimeout      = 15 * time.Second
	AlertGrace        = 2 * time.Second
	PollInterval      = 100 * time.Millisecond
	MaxRunBodyBytes   = 1 << 20 // 1 MiB
	ProcessName       = "plugin_com.utmstack.playground"
	DefaultTenant     = "ce66672c-e36d-4761-a8c8-90058fee1a24"
)

func InternalKey() string {
	return plugins.PluginCfg("com.utmstack").Get("internalKey").String()
}
