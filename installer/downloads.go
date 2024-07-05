package main

import (
	"fmt"

	"github.com/threatwinds/go-sdk/helpers"
)

func Downloads(tag string) error {
	var branch string

	switch tag {
	case "v10-dev":
		branch = "dev"
	case "v10-qa":
		branch = "qa"
	case "v10-rc":
		branch = "rc"
	default:
		branch = "release"
	}

	var downloads = map[string]string{
		// Plugins
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.inputs.plugin", branch):      "/workdir/plugins/utmstack/com.utmstack.inputs.plugin",
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.alerts.plugin", branch):      "/workdir/plugins/utmstack/com.utmstack.alerts.plugin",
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.events.plugin", branch):      "/workdir/plugins/utmstack/com.utmstack.events.plugin",
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.geolocation.plugin", branch): "/workdir/plugins/utmstack/com.utmstack.geolocation.plugin",
		// Base Configurations
		fmt.Sprintf("https://cdn.utmstack.com/%s/pipeline/system_plugins_analysis.yaml", branch):     "/workdir/pipeline/utmstack/system_plugins_analysis.yaml",
		fmt.Sprintf("https://cdn.utmstack.com/%s/pipeline/system_plugins_correlation.yaml", branch):  "/workdir/pipeline/utmstack/system_plugins_correlation.yaml",
		fmt.Sprintf("https://cdn.utmstack.com/%s/pipeline/system_plugins_input.yaml", branch):        "/workdir/pipeline/utmstack/system_plugins_input.yaml",
		fmt.Sprintf("https://cdn.utmstack.com/%s/pipeline/system_plugins_notification.yaml", branch): "/workdir/pipeline/utmstack/system_plugins_notification.yaml",
		fmt.Sprintf("https://cdn.utmstack.com/%s/pipeline/system_plugins_parsing.yaml", branch):      "/workdir/pipeline/utmstack/system_plugins_parsing.yaml",
	}

	for k, v := range downloads {
		if err := helpers.Download(k, v); err != nil {
			return fmt.Errorf(err.Message)
		}
	}

	return nil
}
