package updater

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

type InstanceConfig struct {
	Server      string `yaml:"server"`
	InstanceID  string `yaml:"instance_id"`
	InstanceKey string `yaml:"instance_key"`
}

func RegisterInstance() error {
	instanceInfo := getInstanceInfoFromBackend()

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

	err = updateInstanceInfoInBackend(instanceInfo)
	if err != nil {
		return fmt.Errorf("error updating instance info in backend: %v", err)
	}

	return nil
}
