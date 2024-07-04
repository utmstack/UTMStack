package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/utmstack/UTMStack/collector-installer/config"
	"github.com/utmstack/UTMStack/collector-installer/module"
	"github.com/utmstack/UTMStack/collector-installer/utils"
)

func main() {
	path := utils.GetMyPath()
	utils.InitBeautyLogger("UTMStack Collector Installer", filepath.Join(path, "logs", config.SERV_LOG))
	utils.Logger.PrintBanner()

	if len(os.Args) > 1 {
		mode := os.Args[1]
		if !config.IsValidCollector(os.Args[2]) {
			utils.Logger.WriteFatal("invalid collector type", fmt.Errorf("invalid collector type"))
		}
		collectorType := config.Collector(os.Args[2])
		collector := module.GetCollectorProcess(collectorType)

		switch mode {
		case "run":
			cnf, err := config.ReadConfig()
			if err != nil {
				utils.Logger.WriteFatal("error reading config", err)
			}
			collector := module.GetCollectorProcess(config.Collector(cnf.CollectorType))
			err = collector.Run()
			if err != nil {
				utils.Logger.WriteFatal(fmt.Sprintf("error running %s collector:", cnf.CollectorType), err)
			}
		case "install":
			ip, utmKey := os.Args[3], os.Args[4]

			if strings.Count(utmKey, "*") == len(utmKey) {
				utils.Logger.WriteFatal("the connection key provided is incorrect. Please make sure you use the 'copy' icon from the integrations section to get the value of the masked key value", nil)
			}

			utils.Logger.WriteSimpleMessage(fmt.Sprintf("Installing %s collector...", collectorType))
			if !utils.IsPortOpen(ip, config.AgentManagerPort) || !utils.IsPortOpen(ip, config.LogAuthProxyPort) {
				utils.Logger.WriteFatal("one or more of the requiered ports are closed. Please open ports 9000 and 50051.", nil)
			}

			config.SaveConfig(&config.ServiceTypeConfig{CollectorType: collectorType})
			err := collector.Install(ip, utmKey)
			if err != nil {
				utils.Logger.WriteFatal(fmt.Sprintf("error installing %s collector:", collectorType), err)
			}

			utils.Logger.WriteSimpleMessage(fmt.Sprintf("%s collector installed successfully", collectorType))

		case "uninstall":
			utils.Logger.WriteSimpleMessage(fmt.Sprintf("Uninstalling %s collector", collectorType))

			utils.Logger.WriteSimpleMessage(fmt.Sprintf("Uninstalling %s collector service...", collectorType))
			err := collector.Uninstall()
			if err != nil {
				utils.Logger.WriteFatal(fmt.Sprintf("error uninstalling %s collector:", collectorType), err)
			}

			utils.Logger.WriteSimpleMessage(fmt.Sprintf("%s collector uninstalled successfully", collectorType))

		}
	} else {
		cnf, err := config.ReadConfig()
		if err != nil {
			utils.Logger.WriteFatal("error reading config", err)
		}
		collector := module.GetCollectorProcess(config.Collector(cnf.CollectorType))
		err = collector.Run()
		if err != nil {
			utils.Logger.WriteFatal(fmt.Sprintf("error running %s collector:", cnf.CollectorType), err)
		}
	}
}
