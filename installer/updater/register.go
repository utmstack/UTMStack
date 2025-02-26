package updater

import (
	"fmt"
	"net/http"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

type InstanceConfig struct {
	Server         string `yaml:"server"`
	InstanceID     string `yaml:"instance_id"`
	InstanceKey    string `yaml:"instance_key"`
	OrganizationID string `yaml:"organization_id"`
}

func RegisterInstance(orgID string) error {
	cnf := config.GetConfig()
	instanceConf := InstanceConfig{}

	err := utils.ReadYAML(config.InstanceConfigPath, &instanceConf)
	if err != nil || instanceConf.Server == "" || instanceConf.InstanceID == "" || instanceConf.InstanceKey == "" {
		if orgID == "" {
			return fmt.Errorf("organization ID is required")
		}
		instanceConf.OrganizationID = orgID

		switch cnf.Branch {
		case "alpha":
			instanceConf.Server = config.CMAlpha
		case "beta":
			instanceConf.Server = config.CMBeta
		case "rc":
			instanceConf.Server = config.CMRc
		case "prod":
			instanceConf.Server = config.CMProd
		}

		url := fmt.Sprintf("%s%s?organization-id=%s", instanceConf.Server, config.RegisterInstanceEndpoint, instanceConf.OrganizationID)
		resp, status, err := utils.DoReq[Auth](url, nil, http.MethodPost, nil)
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
