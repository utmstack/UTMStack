package config

import (
	"path/filepath"
	"time"
)

const (
	GetUpdatesInfoEndpoint   = "/api/customer-manager/v1/update-info"
	RegisterInstanceEndpoint = "/api/customer-manager/v1/instance"
	GetEditionEndpoint       = "/api/customer-manager/v1/instance-edition"
	GetFileEndpoint          = "/api/customer-manager/v1/file"
	ExposerPort              = "10091"
	RequiredMinCPUCores      = 2
	RequiredDistro           = "ubuntu"
	RequiredMinDiskSpace     = 30
	CMDev                    = "https://customermanagerdev.utmstack.com"
	CMQa                     = "https://customermanagerqa.utmstack.com"
	CMRc                     = "https://customermanagerrc.utmstack.com"
	CMProd                   = "https://customermanager.utmstack.com"
)

var (
	ConfigPath        = filepath.Join("/root", "utmstack.yml")
	UpdaterConfigPath = filepath.Join(GetConfig().UpdatesFolder, "updater.yml")
	ServiceLogPath    = filepath.Join(GetConfig().UpdatesFolder, "logs", "utmstack-updater.log")
	CertFilePath      = filepath.Join(GetStackConfig().Cert, "utm.crt")
	WindowConfigPath  = filepath.Join(GetConfig().UpdatesFolder, "window.yml")
	UpdatesInfoPath   = filepath.Join(GetConfig().UpdatesFolder, "updates_info.json")
	CheckUpdatesEvery = 5 * time.Minute
)
