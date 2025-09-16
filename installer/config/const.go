package config

import (
	"path/filepath"
	"time"
)

const (
	RegisterInstanceEndpoint   = "/api/v1/instances/register"
	GetInstanceDetailsEndpoint = "/api/v1/instances"
	GetUpdatesInfoEndpoint     = "/api/v1/updates"
	SetUpdateSentEndpoint      = "/api/v1/updates/sent"
	GetLicenseEndpoint         = "/api/v1/licenses"
	HealthEndpoint             = "/api/v1/health"
	LogCollectorEndpoint       = "/api/v1/logcollectors/upload"

	ImagesPath = "/utmstack/images"

	CMServer = "https://customermanager.utmstack.com"

	RequiredMinCPUCores  = 2
	RequiredMinDiskSpace = 30
	RequiredDistroUbuntu = "ubuntu"
	RequiredDistroRHEL   = "redhat"
)

var (
	DEFAULT_BRANCH    string
	INSTALLER_VERSION string
	REPLACE           string
	PUBLIC_KEY        string

	BackendConfigEndpoint  = "https://127.0.0.1/api/utm-configuration-parameters?page=0&size=10000&sectionId.equals=%d&sort=id,asc"
	ConfigPath             = filepath.Join("/root", "utmstack.yml")
	InstanceConfigPath     = filepath.Join(GetConfig().UpdatesFolder, "instance-config.yml")
	ServiceLogPath         = filepath.Join(GetConfig().UpdatesFolder, "logs", "utmstack-updater.log")
	VersionFilePath        = filepath.Join(GetConfig().UpdatesFolder, "version.json")
	LicenseFilePath        = filepath.Join(GetConfig().UpdatesFolder, "LICENSE")
	EventProcessorLogsPath = filepath.Join(GetConfig().DataDir, "events-engine-workdir", "logs")
	CheckUpdatesEvery      = 5 * time.Minute
	SyncSystemLogsEvery    = 5 * time.Minute
	ConnectedToInternet    = false
	Updating               = false
)

func GetCMServer() string {
	cnf := GetConfig()
	return CMServer + "/" + cnf.Branch
}
