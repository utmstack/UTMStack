package docker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
	"gopkg.in/yaml.v3"
)

type storageYAML struct {
	Primary   StorageEngine  `yaml:"primary"`
	Secondary *storageSecond `yaml:"secondary,omitempty"`
}

type StorageEngine struct {
	Engine     string     `yaml:"engine"`
	ClickHouse *storageCH `yaml:"clickhouse,omitempty"`
	OpenSearch *storageOS `yaml:"opensearch,omitempty"`
}

type storageCH struct {
	Addr           []string          `yaml:"addr"`
	Database       string            `yaml:"database"`
	Username       string            `yaml:"username"`
	Password       string            `yaml:"password"`
	TenantColumn   string            `yaml:"tenantColumn"`
	TimeColumn     string            `yaml:"timeColumn"`
	DataTypeColumn string            `yaml:"dataTypeColumn"`
	Tables         map[string]string `yaml:"tables"`
	TextColumns    map[string]string `yaml:"textColumns"`
	IDColumns      map[string]string `yaml:"idColumns"`
}

type storageOS struct {
	URLs              []string                 `yaml:"urls"`
	Username          string                   `yaml:"username"`
	Password          string                   `yaml:"password"`
	TenantField       string                   `yaml:"tenantField"`
	TimestampField    string                   `yaml:"timestampField"`
	RetentionPolicyID string                   `yaml:"retentionPolicyId"`
	Layout            map[string]storageLayout `yaml:"layout"`
}

type storageLayout struct {
	Prefix          string `yaml:"prefix"`
	IncludeDataType bool   `yaml:"includeDataType"`
	DateFormat      string `yaml:"dateFormat"`
}

type storageSecond struct {
	StorageEngine   `yaml:",inline"`
	Cutover         string   `yaml:"cutover"`
	Expires         string   `yaml:"expires"`
	Tenant          string   `yaml:"tenant"`
	Datasets        []string `yaml:"datasets"`
	TolerateFailure bool     `yaml:"tolerateFailure"`
}

type osSecondaryInfo struct {
	URL      string
	Username string
	Password string
}

func buildStorageConfig(password string, secondary *osSecondaryInfo) storageYAML {
	cfg := storageYAML{
		Primary: StorageEngine{
			Engine: "clickhouse",
			ClickHouse: &storageCH{
				Addr:           []string{"clickhouse:9000"},
				Database:       "utmstack",
				Username:       "default",
				Password:       password,
				TenantColumn:   "tenantId",
				TimeColumn:     "@timestamp",
				DataTypeColumn: "dataType",
				Tables:         map[string]string{"logs": "logs", "alerts": "alerts"},
				TextColumns:    map[string]string{"logs": "raw"},
				IDColumns:      map[string]string{"logs": "id"},
			},
		},
	}
	if secondary != nil {
		now := time.Now().UTC()
		cfg.Secondary = &storageSecond{
			StorageEngine: StorageEngine{
				Engine: "opensearch",
				OpenSearch: &storageOS{
					URLs:              []string{"https://" + secondary.URL},
					Username:          secondary.Username,
					Password:          secondary.Password,
					TenantField:       "tenantId",
					TimestampField:    "@timestamp",
					RetentionPolicyID: "utmstack-logs-retention",
					Layout:            map[string]storageLayout{"logs": {Prefix: "v11-log", IncludeDataType: true, DateFormat: "2006.01.02"}},
				},
			},
			Cutover:         now.Format(time.RFC3339),
			Expires:         now.Add(90 * 24 * time.Hour).Format(time.RFC3339),
			Tenant:          "00000000-0000-0000-0000-000000000000",
			Datasets:        []string{"logs"},
			TolerateFailure: false,
		}
	}
	return cfg
}

func EnsureStorageConfig(conf *config.Config, withOpenSearchSecondary bool) error {
	var secondary *osSecondaryInfo
	if withOpenSearchSecondary {
		found, info, err := findOpenSearch()
		if err != nil {
			return fmt.Errorf("detect opensearch: %w", err)
		}
		if found {
			secondary = info
		}
	}
	body, err := yaml.Marshal(buildStorageConfig(conf.Password, secondary))
	if err != nil {
		return err
	}
	dir := utils.MakeDir(0777, conf.DataDir, "storage")
	return os.WriteFile(filepath.Join(dir, "storage.yaml"), body, 0644)
}

var dockerRun = func(arg ...string) ([]string, error) {
	return utils.RunCmdWithOutput("docker", arg...)
}

// DetectOpenSearch reports whether a running container with an opensearch
// image exists
func DetectOpenSearch() (bool, error) {
	found, _, err := findOpenSearch()
	return found, err
}

type psEntry struct {
	Id    string `json:"Id"`
	Image string `json:"Image"`
	State string `json:"State"`
}

func findOpenSearch() (bool, *osSecondaryInfo, error) {
	lines, err := dockerRun("ps", "--format", "json")
	if err != nil {
		return false, nil, err
	}
	for _, l := range lines {
		var e psEntry
		if json.Unmarshal([]byte(l), &e) != nil {
			continue
		}
		if e.State != "running" || !strings.Contains(strings.ToLower(e.Image), "opensearch") {
			continue
		}
		info, err := inspectOpenSearch(e.Id)
		return err == nil, info, err
	}
	return false, nil, nil
}

func inspectOpenSearch(id string) (*osSecondaryInfo, error) {
	lines, err := dockerRun("inspect", "--format",
		"{{json .Config.Env}}\n{{json .NetworkSettings.Ports}}\n{{json .NetworkSettings.Networks}}", id)
	if err != nil {
		return nil, err
	}
	if len(lines) < 3 {
		return nil, fmt.Errorf("unexpected docker inspect output for %s", id)
	}
	var env []string
	var ports map[string][]struct {
		HostIp   string `json:"HostIp"`
		HostPort string `json:"HostPort"`
	}
	var networks map[string]struct {
		IPAddress string `json:"IPAddress"`
	}
	if err := json.Unmarshal([]byte(lines[0]), &env); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(lines[1]), &ports); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(lines[2]), &networks); err != nil {
		return nil, err
	}

	username, password := "admin", ""
	for _, kv := range env {
		if v, ok := strings.CutPrefix(kv, "OS_USERNAME="); ok {
			username = v
		}
		if v, ok := strings.CutPrefix(kv, "OS_PASSWORD="); ok {
			password = v
		}
	}

	host := ""
	if m, ok := ports["9200/tcp"]; ok && len(m) > 0 && m[0].HostIp != "" && m[0].HostPort != "" {
		host = m[0].HostIp + ":" + m[0].HostPort
	}
	if host == "" {
		for _, n := range networks {
			if n.IPAddress != "" {
				host = n.IPAddress + ":9200"
				break
			}
		}
	}
	if host == "" {
		return nil, fmt.Errorf("no 9200 mapping or container IP for opensearch %s", id)
	}
	return &osSecondaryInfo{URL: host, Username: username, Password: password}, nil
}
