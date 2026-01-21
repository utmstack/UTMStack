package updater

import (
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
			Name:    serverConfig.ServerName,
			Edition: "community",
			Version: v.Version,
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

		err = updateInstanceInfo(resp.ID)
		if err != nil {
			return fmt.Errorf("error updating instance info in backend: %v", err)
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

// PollAndUpdateAdminEmail polls for admin email and updates instance details
func PollAndUpdateAdminEmail(instanceConf InstanceConfig) {
	serverConfig := config.GetConfig()
	if serverConfig == nil {
		config.Logger().ErrorF("error: server config is nil in PollAndUpdateAdminEmail")
		return
	}

	for {
		time.Sleep(5 * time.Minute)

		email, err := services.GetAdminEmail(serverConfig)
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

func updateInstanceInfo(id string) error {

	backConf, err := getConfigFromBackend(6)
	if err != nil {
		// If backend is in maintenance, just return without error
		if IsBackendMaintenanceError(err) {
			return nil
		}
		return fmt.Errorf("error getting instance auth from backend: %v", err)
	}

	for i, c := range backConf {
		if c.ConfParamShort == "utmstack.instance.data" {
			backConf[i].ConfParamValue = id
		}
	}

	err = updateConfigInBackend(backConf, 6)
	if err != nil {
		// If backend is in maintenance, just return without error
		if IsBackendMaintenanceError(err) {
			return nil
		}
		return fmt.Errorf("error updating instance info in backend: %v", err)
	}

	return nil
}
