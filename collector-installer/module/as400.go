package module

import (
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/utmstack/UTMStack/collector-installer/config"
	"github.com/utmstack/UTMStack/collector-installer/models"
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

func getAS400Collector() *AS400 {
	as400Once.Do(func() {
		currentPath := utils.GetMyPath()
		as400Collector = AS400{
			Config: ProcessConfig{
				ServiceInfo: config.ServiceConfig{
					Name:        "UTMStackCollectorAS400",
					DisplayName: "UTMStack Collector AS400",
					Description: "UTMStack Collector AS400",
					CMDRun:      getJavaCommand(),
					CMDArgs:     []string{"-jar", filepath.Join(currentPath, config.GetDownloadFilePath(config.DependServiceLabel, config.AS400)), "-option=RUN"},
					CMDPath:     currentPath,
				},
			},
		}
	})

	return &as400Collector
}

func (a *AS400) Run() error {
	err := serv.RunService(a.Config.ServiceInfo)
	if err != nil {
		return fmt.Errorf("error running service: %v", err)
	}
	return nil
}

func (a *AS400) Install(ip, utmKey, skip string) error {
	utils.Logger.WriteSimpleMessage("Downloading UTMStack dependencies...")
	err := a.downloadDependencies(ip, utmKey, skip == "yes")
	if err != nil {
		return fmt.Errorf("error downloading dependencies: %v", err)
	}
	utils.Logger.WriteSimpleMessage("Dependencies downloaded successfully")

	utils.Logger.WriteSimpleMessage("Installing service...")

	err = a.installDependencies()
	if err != nil {
		return fmt.Errorf("error installing dependencies: %v", err)
	}

	result, errB := utils.ExecuteWithResult(
		getJavaCommand(), "", "-jar", config.GetDownloadFilePath(config.DependServiceLabel, config.AS400), "-option=INSTALL",
		fmt.Sprintf("-collector-manager-host=%s", ip), fmt.Sprintf("-collector-manager-port=%s", config.AgentManagerPort),
		fmt.Sprintf("-logs-port=%s", config.LogAuthProxyPort), fmt.Sprintf("-connection-key=%s", utmKey),
	)
	if errB {
		return fmt.Errorf("error executing install command: %s", result)
	}

	err = utils.CheckErrorsInOutput(result)
	if err != nil {
		return fmt.Errorf("error executing install command: %v", err)
	}
	err = serv.InstallService(a.Config.ServiceInfo)
	if err != nil {
		return fmt.Errorf("error installing service: %v", err)
	}

	utils.Logger.WriteSuccessfull("Service installed successfully")

	return nil
}

func (a *AS400) Uninstall() error {
	result, errB := utils.ExecuteWithResult(
		getJavaCommand(), "", "-jar", config.GetDownloadFilePath(config.DependServiceLabel, config.AS400), "-option=UNINSTALL",
	)
	if errB {
		return fmt.Errorf("error executing uninstall command: %v", result)
	}

	err := utils.CheckErrorsInOutput(result)
	if err != nil {
		return fmt.Errorf("error executing uninstall command: %v", err)
	}

	err = a.uninstallDependencies()
	if err != nil {
		return fmt.Errorf("error uninstalling dependencies: %v", err)
	}

	err = serv.UninstallService(a.Config.ServiceInfo)
	if err != nil {
		return fmt.Errorf("error uninstalling service: %v", err)
	}

	return nil
}

func (a *AS400) downloadDependencies(ip, utmKey string, skip bool) error {
	version := models.Version{}
	var err error

	headers := map[string]string{"connection-key": utmKey}

	if version.ServiceVersion, err = utils.DownloadFileByChunks(fmt.Sprintf(config.DEPEND_URL, ip, "0", runtime.GOOS, config.DependServiceLabel, string(config.AS400)), headers, config.GetDownloadFilePath(config.DependServiceLabel, config.AS400), skip); err != nil {
		return fmt.Errorf("error downloading service: %v", err)
	}

	switch runtime.GOOS {
	case "windows":
		if version.DependenciesVersion, err = utils.DownloadFileByChunks(fmt.Sprintf(config.DEPEND_URL, ip, "0", runtime.GOOS, config.DependZipLabel, string(config.AS400)), headers, config.GetDownloadFilePath(config.DependZipLabel, config.AS400), skip); err != nil {
			return fmt.Errorf("error downloading dependencies: %v", err)
		}
	case "linux":
		resp, _, err := utils.DoReq[models.DependencyUpdateResponse](fmt.Sprintf(config.DEPEND_URL, ip, "0", runtime.GOOS, config.DependZipLabel, string(config.AS400)), nil, http.MethodGet, headers, skip)
		if err != nil && !strings.Contains(resp.Message, "dependency not found") {
			return fmt.Errorf("error downloading dependencies: %v", err)
		}
		version.DependenciesVersion = resp.Version
	}

	if err := utils.WriteJSON(config.GetVersionPath(), &version); err != nil {
		return fmt.Errorf("error writing version file: %v", err)
	}

	err = HandleDependenciesPostDownload(config.AS400)
	if err != nil {
		return fmt.Errorf("error handling dependencies post download: %v", err)
	}

	return nil
}

func (a *AS400) installDependencies() error {
	err := InstallJava()
	if err != nil {
		return fmt.Errorf("error installing Java: %v", err)
	}
	return nil
}

func (a *AS400) uninstallDependencies() error {
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
		javaCommand = filepath.Join(currentPath, As400JavaVersion, "bin", "java.exe")
	case "linux":
		javaCommand = "java"
	}
	return javaCommand
}
