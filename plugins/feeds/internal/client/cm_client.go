package client

import (
	"fmt"
	"net/http"
	"time"

	"github.com/threatwinds/go-sdk/catcher"

	sdkutils "github.com/threatwinds/go-sdk/utils"
	"github.com/utmstack/UTMStack/plugins/feeds/utils"
)

const (
	instanceConfigPath = "/updates/instance-config.yml"
)

type CustomersManagerClient struct {
	Server      string `yaml:"server"`
	InstanceID  string `yaml:"instance_id"`
	InstanceKey string `yaml:"instance_key"`
}

type RegistrationResponse struct {
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
}

func (c *CustomersManagerClient) LoadInstanceConfig() error {
	time.Sleep(10 * time.Second)

	loadFunc := func() error {
		if !utils.CheckIfPathExist(instanceConfigPath) {
			return catcher.Error("config file not found", nil, nil)
		}

		if err := utils.ReadYAML(instanceConfigPath, c); err != nil {
			return catcher.Error("failed to read or parse YAML config", err, nil)
		}

		if c.Server == "" || c.InstanceID == "" || c.InstanceKey == "" {
			return catcher.Error("missing required fields in config", nil, nil)
		}

		return nil
	}

	return utils.Retry(loadFunc, "instance config loading", utils.DefaultRetryConfig())
}

func (c *CustomersManagerClient) RegisterUserReporter() (*RegistrationResponse, error) {
	if c.Server == "" || c.InstanceID == "" || c.InstanceKey == "" {
		return nil, catcher.Error("instance configuration not loaded", nil, nil)
	}

	endpoint := fmt.Sprintf("%s/api/v1/intelligence/register", c.Server)

	headers := map[string]string{
		"accept": "application/json",
		"id":     c.InstanceID,
		"Key":    c.InstanceKey,
	}

	credentials, _, err := sdkutils.DoReq[RegistrationResponse](
		endpoint,
		nil,
		http.MethodPost,
		headers,
	)

	if err != nil {
		return nil, err
	}

	catcher.Info("Successfully registered ThreadWinds intelligence reporter", nil)

	return &credentials, nil
}
