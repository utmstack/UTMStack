package main

import (
	"fmt"
	"path/filepath"
	"time"

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

	pluginsFolder := utils.MakeDir(0777, stack.EventsEngineWorkdir, "plugins")

	pipelineFolder := utils.MakeDir(0777, stack.EventsEngineWorkdir, "pipeline")

	var downloads = map[string]string{
		// Plugins
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.inputs.plugin", branch):      filepath.Join(pluginsFolder, "com.utmstack.inputs.plugin"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.alerts.plugin", branch):      filepath.Join(pluginsFolder, "com.utmstack.alerts.plugin"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.events.plugin", branch):      filepath.Join(pluginsFolder, "com.utmstack.events.plugin"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.geolocation.plugin", branch): filepath.Join(pluginsFolder, "com.utmstack.geolocation.plugin"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.gcp.plugin", branch):         filepath.Join(pluginsFolder, "com.utmstack.gcp.plugin"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.azure.plugin", branch):       filepath.Join(pluginsFolder, "com.utmstack.azure.plugin"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.sophos.plugin", branch):      filepath.Join(pluginsFolder, "com.utmstack.sophos.plugin"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.aws.plugin", branch):         filepath.Join(pluginsFolder, "com.utmstack.aws.plugin"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.stats.plugin", branch):       filepath.Join(pluginsFolder, "com.utmstack.stats.plugin"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.o365.plugin", branch):        filepath.Join(pluginsFolder, "com.utmstack.o365.plugin"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.config.plugin", branch):      filepath.Join(pluginsFolder, "com.utmstack.config.plugin"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/plugins/com.utmstack.bitdefender.plugin", branch): filepath.Join(pluginsFolder, "com.utmstack.bitdefender.plugin"),
		// Base Configurations
		fmt.Sprintf("https://cdn.utmstack.com/%s/pipeline/system_plugins_analysis.yaml", branch):     filepath.Join(pipelineFolder, "system_plugins_analysis.yaml"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/pipeline/system_plugins_correlation.yaml", branch):  filepath.Join(pipelineFolder, "system_plugins_correlation.yaml"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/pipeline/system_plugins_input.yaml", branch):        filepath.Join(pipelineFolder, "system_plugins_input.yaml"),
		fmt.Sprintf("https://cdn.utmstack.com/%s/pipeline/system_plugins_notification.yaml", branch): filepath.Join(pipelineFolder, "system_plugins_notification.yaml"),
	}

	if err := utils.RunCmd("systemctl", "stop", "docker"); err != nil {
		return err
	}

	for k, v := range downloads {
		if err := helpers.Download(k, v); err != nil {
			return fmt.Errorf(err.Message)
		}
	}

	if err := utils.RunCmd("chmod", "+x", "-R", filepath.Join(pluginsFolder)); err != nil {
		return err
	}

	if err := utils.RunCmd("systemctl", "start", "docker"); err != nil {
		return err
	}

	time.Sleep(60 * time.Second)

	return nil
}
