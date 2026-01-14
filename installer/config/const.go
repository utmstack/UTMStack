package config

import (
	"path/filepath"
	"time"
)

const (
	RegisterInstanceEndpoint      = "/api/v1/instances/register"
	GetInstanceDetailsEndpoint    = "/api/v1/instances"
	UpdateInstanceDetailsEndpoint = "/api/v1/instances/details"
	HeartbeatEndpoint             = "/api/v1/instances/heartbeat"
	GetUpdatesInfoEndpoint        = "/api/v1/updates"
	SetUpdateSentEndpoint         = "/api/v1/updates/sent"
	GetLicenseEndpoint            = "/api/v1/licenses"
	HealthEndpoint                = "/api/v1/health"
	LogCollectorEndpoint          = "/api/v1/logcollectors/upload"

	GitHubReleasesURL = "https://github.com/utmstack/UTMStack/releases/download/%s/installer"

	ImagesPath       = "/utmstack/images"
	InstallerBinPath = "/usr/local/bin/utmstack_installer"

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
	PendingUpdatesPath     = filepath.Join(GetConfig().UpdatesFolder, "pending-updates.json")
	LastAdminEmailPath     = filepath.Join(GetConfig().UpdatesFolder, "last-admin-email.txt")
	EventProcessorLogsPath = filepath.Join(GetConfig().DataDir, "events-engine-workdir", "logs")
	CheckUpdatesEvery      = 5 * time.Minute
	SyncSystemLogsEvery    = 5 * time.Minute
	ConnectedToInternet    = false
)

func GetCMServer() string {
	cnf := GetConfig()
	if cnf.Branch == "alpha" {
		return "https://cm.dev.utmstack.com"
	}
	return "https://cm.utmstack.com"
}
