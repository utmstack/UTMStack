package config

import (
	"path/filepath"
	"time"
)

const (
	RegisterInstanceEndpoint   = "/api/v1/instances/register"
	GetInstanceDetailsEndpoint = "/api/v1/instances"
	GetUpdatesInfoEndpoint     = "/api/v1/updates"

	CMAlpha = "https://customermanager.utmstack.com/alpha"
	CMBeta  = "https://customermanager.utmstack.com/beta"
	CMRc    = "https://customermanager.utmstack.com/rc"
	CMProd  = "https://customermanager.utmstack.com/prod"

	RequiredMinCPUCores  = 2
	RequiredMinDiskSpace = 30
	RequiredDistroUbuntu = "ubuntu"
	RequiredDistroRHEL   = "redhat"
)

var (
	ConfigPath         = filepath.Join("/root", "utmstack.yml")
	InstanceConfigPath = filepath.Join(GetConfig().UpdatesFolder, "instance-config.yml")
	ServiceLogPath     = filepath.Join(GetConfig().UpdatesFolder, "logs", "utmstack-updater.log")
	WindowConfigPath   = filepath.Join(GetConfig().UpdatesFolder, "update-window.yml")
	VersionFilePath    = filepath.Join(GetConfig().UpdatesFolder, "version.json")
	CheckUpdatesEvery  = 5 * time.Minute
)
