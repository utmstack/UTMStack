package module

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/utmstack/UTMStack/collector-installer/agent"
	"github.com/utmstack/UTMStack/collector-installer/config"
	serv "github.com/utmstack/UTMStack/collector-installer/service"
	"github.com/utmstack/UTMStack/collector-installer/utils"
)

var (
	as400Collector   AS400
	as400Once        sync.Once
	As400JavaVersion = "jdk-11.0.13.8"
)

type AS400 struct {
	Config ProcessConfig
}

func getAS400Collector(logger *utils.BeautyLogger) *AS400 {
	as400Once.Do(func() {
		currentPath := utils.GetMyPath()
		as400Collector = AS400{
			Config: ProcessConfig{
				ServiceInfo: config.ServiceConfig{
					Name:        "UTMStackCollectorAS400",
					DisplayName: "UTMStack Collector AS400",
					Description: "UTMStack Collector AS400",
					CMDRun:      getJavaCommand(),
					CMDArgs:     []string{"-jar", filepath.Join(currentPath, config.GetDownloadFilePath("service", config.AS400)), "-option=RUN"},
					CMDPath:     currentPath,
				},
				Logger: logger,
			},
		}
	})

	return &as400Collector
}

func (a *AS400) Run() error {
	err := serv.RunService(a.Config.ServiceInfo, a.Config.Logger)
	if err != nil {
		return fmt.Errorf("error running service: %v", err)
	}
	return nil
}

func (a *AS400) Install(ip string, utmKey string) error {
	currentPath := utils.GetMyPath()

	a.Config.Logger.WriteSimpleMessage("Downloading UTMStack dependencies...")
	err := agent.DownloadDependencies(ip, utmKey, config.AS400, a.Config.Logger)
	if err != nil {
		return fmt.Errorf("error downloading dependencies: %v", err)
	}
	a.Config.Logger.WriteSimpleMessage("Dependencies downloaded successfully")

	a.Config.Logger.WriteSimpleMessage("Installing service...")

	err = a.InstallDependencies()
	if err != nil {
		return fmt.Errorf("error installing dependencies: %v", err)
	}

	result, errB := utils.ExecuteWithResult(
		getJavaCommand(), "", "-jar", filepath.Join(currentPath, config.GetDownloadFilePath("service", config.AS400)), "-option=INSTALL",
		fmt.Sprintf("-collector-manager-host=%s", ip), fmt.Sprintf("-collector-manager-port=%s", config.AgentManagerPort),
		fmt.Sprintf("-logs-port=%s", config.LogAuthProxyPort), fmt.Sprintf("-connection-key=%s", utmKey),
	)
	if errB {
		return fmt.Errorf("error executing install command: %v", err)
	}

	err = utils.CheckErrorsInOutput(result)
	if err != nil {
		return fmt.Errorf("error executing install command: %v", err)
	}
	err = serv.InstallService(a.Config.ServiceInfo)
	if err != nil {
		return fmt.Errorf("error installing service: %v", err)
	}

	a.Config.Logger.WriteSuccessfull("Service installed successfully")

	return nil
}

func (a *AS400) InstallDependencies() error {
	err := InstallJava()
	if err != nil {
		return fmt.Errorf("error installing Java: %v", err)
	}
	return nil
}

func (a *AS400) Uninstall() error {
	currentPath := utils.GetMyPath()

	result, errB := utils.ExecuteWithResult(
		getJavaCommand(), "", "-jar", filepath.Join(currentPath, config.GetDownloadFilePath("service", config.AS400)), "-option=UNINSTALL",
	)
	if errB {
		return fmt.Errorf("error executing uninstall command: %v", result)
	}

	err := utils.CheckErrorsInOutput(result)
	if err != nil {
		return fmt.Errorf("error executing uninstall command: %v", err)
	}

	err = a.UninstallDependencies()
	if err != nil {
		return fmt.Errorf("error uninstalling dependencies: %v", err)
	}

	err = serv.UninstallService(a.Config.ServiceInfo)
	if err != nil {
		return fmt.Errorf("error uninstalling service: %v", err)
	}

	return nil
}

func (a *AS400) UninstallDependencies() error {
	err := UninstallJava()
	if err != nil {
		return fmt.Errorf("error uninstalling Java: %v", err)
	}
	return nil
}

func getJavaCommand() string {
	currentPath := utils.GetMyPath()
	javaCommand := ""
	switch runtime.GOOS {
	case "windows":
		javaCommand = filepath.Join(currentPath, "dependencies", As400JavaVersion, "bin", "java.exe")
	case "linux":
		javaCommand = "java"
	}
	return javaCommand
}
