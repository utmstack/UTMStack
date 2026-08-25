package config

import (
	"path/filepath"

	"github.com/utmstack/UTMStack/collectors/utmstack/models"
	"github.com/utmstack/UTMStack/collectors/utmstack/utils"
)

const (
	REPLACE_KEY      string = ""
	DockerHost       string = ""
	DockerAPIVersion string = "1.41"
	MaxLogLength     int    = 32768
	BufferSize       int    = 50000

	DataType string = "utmstack"
)

var (
	// Everything the server exposes now answers on 443, behind the edge that
	// routes by path: gRPC by its /package.Service/ prefix, downloads by theirs.
	// One port is what a customer has to open, and it is the one their network
	// already allows.
	DependUrl        = "https://%s:%s/private/dependencies/collector/%s"
	AgentManagerPort = "443"
	LogAuthProxyPort = "443"
	DependenciesPort = "443"

	ServiceLogFile    = filepath.Join(utils.GetMyPath(), "logs", "utmstack_collector.log")
	UUIDFileName      = filepath.Join(utils.GetMyPath(), "uuid.yml")
	ConfigurationFile = filepath.Join(utils.GetMyPath(), "config.yml")
	VersionPath       = filepath.Join(utils.GetMyPath(), "version.json")
)

var FilterRules = []models.FilterRule{
	{
		Type:    "name",
		Pattern: ".*postgres.*", // Exclude containers with "postgres" in the name
		Action:  "exclude",
	},
	{
		Type:    "name",
		Pattern: ".*logstash.*", // Exclude containers with "logstash" in the name
		Action:  "exclude",
	},
	{
		Type:    "name",
		Pattern: ".*mutate.*", // Exclude containers with "mutate" in the name
		Action:  "exclude",
	},
	{
		Type:    "name",
		Pattern: ".*node1.*", // Exclude containers with "node1" in the name
		Action:  "exclude",
	},
	{
		Type:    "name",
		Pattern: ".*filebrowser.*", // Exclude containers with "filebrowser" in the name
		Action:  "exclude",
	},
	{
		Type:    "name",
		Pattern: ".*user-auditor.*", // Exclude containers with "user-auditor" in the name
		Action:  "exclude",
	},
	{
		Type:    "name",
		Pattern: ".*web-pdf.*", // Exclude containers with "web-pdf" in the name
		Action:  "exclude",
	},
	{
		Type:    "name",
		Pattern: ".*frontend.*", // Exclude containers with "frontend" in the name
		Action:  "exclude",
	},
}
