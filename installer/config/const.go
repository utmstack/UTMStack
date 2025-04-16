package config

import (
	"path/filepath"
	"time"
)

const PUBLIC_KEY = ``

const (
	REPLACE                    = ""
	RegisterInstanceEndpoint   = "/api/v1/instances/register"
	GetInstanceDetailsEndpoint = "/api/v1/instances"
	GetUpdatesInfoEndpoint     = "/api/v1/updates"
	GetLicenseEndpoint         = "/api/v1/licenses"
	GetLatestVersionEndpoint   = "/api/v1/versions/latest"
	HealthEndpoint             = "/api/v1/health"
	ImagesPath                 = "/utmstack/images"

	CMServer = "https://customermanager.utmstack.com"

	RequiredMinCPUCores  = 2
	RequiredMinDiskSpace = 30
	RequiredDistroUbuntu = "ubuntu"
	RequiredDistroRHEL   = "redhat"
)

var (
	BackendConfigEndpoint = "https://127.0.0.1/api/utm-configuration-parameters?page=0&size=10000&sectionId.equals=%d&sort=id,asc"
	ConfigPath            = filepath.Join("/root", "utmstack.yml")
	InstanceConfigPath    = filepath.Join(GetConfig().UpdatesFolder, "instance-config.yml")
	ServiceLogPath        = filepath.Join(GetConfig().UpdatesFolder, "logs", "utmstack-updater.log")
	WindowConfigPath      = filepath.Join(GetConfig().UpdatesFolder, "update-window.yml")
	VersionFilePath       = filepath.Join(GetConfig().UpdatesFolder, "version.json")
	LicenseFilePath       = filepath.Join(GetConfig().UpdatesFolder, "LICENSE")
	CheckUpdatesEvery     = 5 * time.Minute
	ConnectedToInternet   = false
	Updating              = false
)

func GetCMServer() string {
	cnf := GetConfig()
	return CMServer + "/" + cnf.Branch
}
