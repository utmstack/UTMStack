package updater

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/docker"
	"github.com/utmstack/UTMStack/installer/services"
	"github.com/utmstack/UTMStack/installer/utils"
)

type InstanceConfig struct {
	Server      string `yaml:"server"`
	InstanceID  string `yaml:"instance_id"`
	InstanceKey string `yaml:"instance_key"`
}

// usageDayLayout matches Customer Manager's UsageDayInput.Day format
// (domain.UsageDay in CustomerManager/backend): YYYY-MM-DD, UTC.
const usageDayLayout = "2006-01-02"

const usageReportInterval = 15 * time.Minute

// tenantUsageStat mirrors the local backend's internal-only usage-report
// response (datasources/dto.TenantUsageDTO in UTMStack/backend) — one
// tenant's ingest volume for the requested day.
type tenantUsageStat struct {
	TenantID   string `json:"tenant_id"`
	EventCount int64  `json:"event_count"`
	Bytes      int64  `json:"bytes"`
}

// usageDayInput mirrors Customer Manager's domain.UsageDayInput.
type usageDayInput struct {
	Day         string `json:"day"`
	TenantID    string `json:"tenant_id,omitempty"`
	IngestBytes int64  `json:"ingest_bytes"`
	EventCount  int64  `json:"event_count"`
}

// usageReportInput mirrors Customer Manager's domain.UsageReportInput.
type usageReportInput struct {
	Version string          `json:"version,omitempty"`
	Days    []usageDayInput `json:"days,omitempty"`
}

func RegisterInstance() error {
	if config.ConnectedToInternet {
		v, err := GetVersion()
		if err != nil {
			return fmt.Errorf("error getting version: %v", err)
		}

		instanceConf := InstanceConfig{
			Server: config.GetCMServer(),
		}

		serverConfig := config.GetConfig()
		if serverConfig == nil {
			return fmt.Errorf("error: server config is nil")
		}

		instanceRegisterReq := InstanceDTOInput{
			Name:        serverConfig.ServerName,
			Edition:     "community",
			Version:     v.Version,
			ProductSlug: config.GetProductSlug(),
		}

		if serverConfig.MappingName != nil && *serverConfig.MappingName != "" {
			instanceRegisterReq.MappingName = *serverConfig.MappingName
		}

		// Check if this is a SaaS instance
		stack := docker.GetStackConfig()
		saasLockPath := filepath.Join(stack.LocksDir, "saas.lock")
		if utils.CheckIfPathExist(saasLockPath) {
			instanceRegisterReq.Tags = "SAAS"
		}

		instanceJSON, err := json.Marshal(instanceRegisterReq)
		if err != nil {
			return fmt.Errorf("error marshalling instance register request: %v", err)
		}

		resp, status, err := utils.DoReq[Auth](fmt.Sprintf("%s%s", instanceConf.Server, config.RegisterInstanceEndpoint), instanceJSON, http.MethodPost, nil, nil)
		if err != nil || status != http.StatusOK {
			return fmt.Errorf("error registering instance: status code: %d, error %v", status, err)
		}

		instanceConf.InstanceID = resp.ID
		instanceConf.InstanceKey = resp.Key

		err = utils.WriteYAML(config.InstanceConfigPath, instanceConf)
		if err != nil {
			return fmt.Errorf("error writing instance config file: %v", err)
		}
	}

	return nil
}

// StartHeartbeat sends heartbeat to CM every minute
func StartHeartbeat(instanceConf InstanceConfig) {
	for {
		time.Sleep(1 * time.Minute)

		url := fmt.Sprintf("%s%s", instanceConf.Server, config.HeartbeatEndpoint)
		_, status, err := utils.DoReq[any](
			url,
			nil,
			http.MethodPost,
			map[string]string{"id": instanceConf.InstanceID, "key": instanceConf.InstanceKey},
			nil,
		)

		if err != nil || status != http.StatusOK {
			config.Logger().ErrorF("error sending heartbeat: status: %d, error: %v", status, err)
		}
	}
}

func StartUsageReporting(instanceConf InstanceConfig) {
	localTransport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	for {
		time.Sleep(usageReportInterval)

		today := time.Now().UTC().Format(usageDayLayout)

		stats, status, err := utils.DoReq[[]tenantUsageStat](
			fmt.Sprintf("https://127.0.0.1/api/v1/datasources/usage-report?day=%s", today),
			nil,
			http.MethodGet,
			map[string]string{"X-Internal-Key": config.GetConfig().InternalKey},
			localTransport,
		)
		if err != nil || status != http.StatusOK {
			config.Logger().ErrorF("error reading local usage stats: status: %d, error: %v", status, err)
			continue
		}
		if len(stats) == 0 {
			continue
		}

		days := make([]usageDayInput, 0, len(stats))
		for _, s := range stats {
			days = append(days, usageDayInput{
				Day:         today,
				TenantID:    s.TenantID,
				IngestBytes: s.Bytes,
				EventCount:  s.EventCount,
			})
		}

		version, err := GetVersion()
		if err != nil {
			config.Logger().ErrorF("error getting version for usage report: %v", err)
			continue
		}

		reportJSON, err := json.Marshal(usageReportInput{Version: version.Version, Days: days})
		if err != nil {
			config.Logger().ErrorF("error marshalling usage report: %v", err)
			continue
		}

		url := fmt.Sprintf("%s%s", instanceConf.Server, config.UsageReportEndpoint)
		_, status, err = utils.DoReq[any](
			url,
			reportJSON,
			http.MethodPost,
			map[string]string{"id": instanceConf.InstanceID, "key": instanceConf.InstanceKey},
			nil,
		)
		if err != nil || status != http.StatusOK {
			config.Logger().ErrorF("error sending usage report: status: %d, error: %v", status, err)
		}
	}
}

// PollAndUpdateAdminEmail polls for admin email and updates instance details
func PollAndUpdateAdminEmail(instanceConf InstanceConfig) {
	serverConfig := config.GetConfig()
	if serverConfig == nil {
		config.Logger().ErrorF("error: server config is nil in PollAndUpdateAdminEmail")
		return
	}

	for {
		time.Sleep(5 * time.Minute)

		email, err := services.GetAdminEmail()
		if err != nil {
			config.Logger().ErrorF("error getting admin email: %v", err)
			continue
		}

		if email == "" {
			continue
		}

		// Check if this email was already sent
		lastEmail, _ := os.ReadFile(config.LastAdminEmailPath)
		if strings.TrimSpace(string(lastEmail)) == email {
			return
		}

		// Email found, update instance details
		updateReq := InstanceDTOInput{
			Name:  serverConfig.ServerName,
			Email: email,
		}

		reqJSON, err := json.Marshal(updateReq)
		if err != nil {
			config.Logger().ErrorF("error marshalling update request: %v", err)
			continue
		}

		url := fmt.Sprintf("%s%s", instanceConf.Server, config.UpdateInstanceDetailsEndpoint)
		_, status, err := utils.DoReq[any](
			url,
			reqJSON,
			http.MethodPut,
			map[string]string{"id": instanceConf.InstanceID, "key": instanceConf.InstanceKey},
			nil,
		)

		if err != nil || status != http.StatusOK {
			config.Logger().ErrorF("error updating instance details: status: %d, error: %v", status, err)
			continue
		}

		// Save the email to avoid re-sending
		_ = os.WriteFile(config.LastAdminEmailPath, []byte(email), 0644)

		config.Logger().Info("Successfully updated instance with admin email: %s", email)
		return
	}
}
