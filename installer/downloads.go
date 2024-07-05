package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/threatwinds/go-sdk/helpers"
	"github.com/utmstack/UTMStack/installer/types"
	"github.com/utmstack/UTMStack/installer/utils"
)

func Downloads(conf *types.Config, stack *types.StackConfig) error {
	var branch string

	switch conf.Branch {
	case "v10-dev":
		branch = "dev"
	case "v10-qa":
		branch = "qa"
	case "v10-rc":
		branch = "rc"
	default:
		branch = "release"
	}

	pluginsFolder := filepath.Join(stack.EventsEngineWorkdir, "plugins")
	err := os.MkdirAll(pluginsFolder, 0777)
	if err != nil {
		return err
	}

	pipelineFolder := filepath.Join(stack.EventsEngineWorkdir, "pipeline")
	err = os.MkdirAll(pipelineFolder, 0777)
	if err != nil {
		return err
	}

	var downloads = map[string]string{
		// Plugins
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.inputs.plugin", branch):      filepath.Join(pluginsFolder, "com.utmstack.inputs.plugin"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.alerts.plugin", branch):      filepath.Join(pluginsFolder, "com.utmstack.alerts.plugin"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.events.plugin", branch):      filepath.Join(pluginsFolder, "com.utmstack.events.plugin"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.geolocation.plugin", branch): filepath.Join(pluginsFolder, "com.utmstack.geolocation.plugin"),
		// Base Configurations
		fmt.Sprintf("https://cdn.utmstack.com/%s/pipeline/system_plugins_analysis.yaml", branch):     filepath.Join(pipelineFolder, "system_plugins_analysis.yaml"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/pipeline/system_plugins_correlation.yaml", branch):  filepath.Join(pipelineFolder, "system_plugins_correlation.yaml"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/pipeline/system_plugins_input.yaml", branch):        filepath.Join(pipelineFolder, "system_plugins_input.yaml"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/pipeline/system_plugins_notification.yaml", branch): filepath.Join(pipelineFolder, "system_plugins_notification.yaml"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/pipeline/system_plugins_parsing.yaml", branch):      filepath.Join(pipelineFolder, "system_plugins_parsing.yaml"),
	}

	for k, v := range downloads {
		if err := helpers.Download(k, v); err != nil {
			return fmt.Errorf(err.Message)
		}
	}

	if err := utils.RunCmd("chmod", "+x", "-R", filepath.Join(pluginsFolder)); err != nil {
		return err
	}

	return nil
}
