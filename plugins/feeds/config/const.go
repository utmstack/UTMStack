package config

import (
	"strings"

	"github.com/threatwinds/go-sdk/plugins"
)

func GetThreadWindsURL() string {
	if isDevEnvironment() {
		return "https://apis.dev.threatwinds.com"
	}
	return "https://apis.threatwinds.com"
}

func isDevEnvironment() bool {
	env := plugins.PluginCfg("com.utmstack").Get("env").String()
	if env != "" {
		if strings.Contains(env, "dev") ||
			strings.Contains(env, "qa") ||
			strings.Contains(env, "rc") {
			return true
		}
	}

	return false
}
