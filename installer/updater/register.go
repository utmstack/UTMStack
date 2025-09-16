package updater

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

type InstanceConfig struct {
	Server      string `yaml:"server"`
	InstanceID  string `yaml:"instance_id"`
	InstanceKey string `yaml:"instance_key"`
}

func RegisterInstance() error {
	instanceInfo := getInstanceInfo()

	v, err := GetVersion()
	if err != nil {
		return fmt.Errorf("error getting version: %v", err)
	}

	instanceInfo.Version = v.Version

	if config.ConnectedToInternet {
		instanceConf := InstanceConfig{
			Server: config.GetCMServer(),
		}

		instanceRegisterReq := InstanceDTOInput{
			Name:    instanceInfo.Name,
			Country: instanceInfo.Country,
			Email:   instanceInfo.Email,
			Edition: "community",
			Version: instanceInfo.Version,
		}

		serverConfig := config.GetConfig()
		if serverConfig != nil && (serverConfig.MappingName != nil && *serverConfig.MappingName != "") {
			instanceRegisterReq.MappingName = *serverConfig.MappingName
		}

		instanceJSON, err := json.Marshal(instanceRegisterReq)
		if err != nil {
			return fmt.Errorf("error marshalling instance register request: %v", err)
		}

		resp, status, err := utils.DoReq[Auth](fmt.Sprintf("%s%s", instanceConf.Server, config.RegisterInstanceEndpoint), instanceJSON, http.MethodPost, nil, nil)
		if err != nil || status != http.StatusOK {
			return fmt.Errorf("error registering instance: status code: %d, error %v", status, err)
		}

		instanceInfo.InstanceID = resp.ID
		instanceConf.InstanceID = resp.ID
		instanceConf.InstanceKey = resp.Key

		err = utils.WriteYAML(config.InstanceConfigPath, instanceConf)
		if err != nil {
			return fmt.Errorf("error writing instance config file: %v", err)
		}
	}

	err = updateInstanceInfo(instanceInfo)
	if err != nil {
		return fmt.Errorf("error updating instance info in backend: %v", err)
	}

	return nil
}

func getInstanceInfo() InstanceInfo {
	var instanceInfo InstanceInfo

	for {
		time.Sleep(30 * time.Second)
		backConf, err := getConfigFromBackend(6)
		if err != nil {
			// Only log if it's not a maintenance/backend down error
			if !IsBackendMaintenanceError(err) {
				config.Logger().Info("instance info not ready yet, retrying after error: %v", err)
			}
			continue
		}

		for _, c := range backConf {
			switch c.ConfParamShort {
			case "utmstack.instance.organization":
				instanceInfo.Name = c.ConfParamValue
			case "utmstack.instance.country":
				instanceInfo.Country = c.ConfParamValue
			case "utmstack.instance.contact_email":
				instanceInfo.Email = c.ConfParamValue
			}
		}

		if instanceInfo.Name == "" || instanceInfo.Country == "" || instanceInfo.Email == "" {
			config.Logger().Info("instance info not ready yet, retrying after incomplete data")
			continue
		}

		break
	}

	return instanceInfo
}

func updateInstanceInfo(instanceInfo InstanceInfo) error {
	jsonData, err := json.Marshal(instanceInfo)
	if err != nil {
		return fmt.Errorf("error marshalling instance info: %v", err)
	}
	instanceInfoBase64 := base64.StdEncoding.EncodeToString(jsonData)

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
			backConf[i].ConfParamValue = instanceInfoBase64
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
