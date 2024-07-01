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
	beautyLogger := utils.GetBeautyLogger(filepath.Join(path, "logs", config.SERV_LOG))
	beautyLogger.PrintBanner()

	if len(os.Args) > 1 {
		mode := os.Args[1]
		if !config.IsValidCollector(os.Args[2]) {
			beautyLogger.WriteFatal("Invalid collector type", fmt.Errorf("invalid collector type"))
		}
		collectorType := config.Collector(os.Args[2])
		collector := module.GetCollectorProcess(collectorType, beautyLogger)

		switch mode {
		case "run":
			err := collector.Run()
			if err != nil {
				beautyLogger.WriteFatal(fmt.Sprintf("Error running %s collector:", collectorType), err)
			}
		case "install":
			ip, utmKey := os.Args[3], os.Args[4]

			if strings.Count(utmKey, "*") == len(utmKey) {
				beautyLogger.WriteFatal("The connection key provided is incorrect. Please make sure you use the 'copy' icon from the integrations section to get the value of the masked key value.", nil)
			}

			beautyLogger.WriteSimpleMessage(fmt.Sprintf("Installing %s collector...", collectorType))
			if !utils.IsPortOpen(ip, config.AgentManagerPort) || !utils.IsPortOpen(ip, config.LogAuthProxyPort) {
				beautyLogger.WriteFatal("one or more of the requiered ports are closed. Please open ports 9000 and 50051.", nil)
			}

			config.SaveConfig(&config.ServiceTypeConfig{CollectorType: collectorType})
			err := collector.Install(ip, utmKey)
			if err != nil {
				beautyLogger.WriteFatal(fmt.Sprintf("Error installing %s collector:", collectorType), err)
			} else {
				beautyLogger.WriteSimpleMessage(fmt.Sprintf("%s collector installed successfully", collectorType))
			}
		case "uninstall":
			beautyLogger.WriteSimpleMessage(fmt.Sprintf("Uninstalling %s collector", collectorType))

			beautyLogger.WriteSimpleMessage(fmt.Sprintf("Uninstalling %s collector service...", collectorType))
			err := collector.Uninstall()
			if err != nil {
				beautyLogger.WriteFatal(fmt.Sprintf("Error uninstalling %s collector:", collectorType), err)
			} else {
				beautyLogger.WriteSimpleMessage(fmt.Sprintf("%s collector uninstalled successfully", collectorType))
			}
		}
	} else {
		cnf, err := config.ReadConfig()
		if err != nil {
			beautyLogger.WriteFatal("Error reading config", err)
		}
		collector := module.GetCollectorProcess(config.Collector(cnf.CollectorType), beautyLogger)
		err = collector.Run()
		if err != nil {
			beautyLogger.WriteFatal(fmt.Sprintf("Error running %s collector:", cnf.CollectorType), err)
		}
	}
}
