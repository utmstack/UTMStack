package client

import (
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/feeds/utils"
)

const (
	instanceConfigPath   = "/updates/instance-config.yml"
	instanceLoadInterval = 5 * time.Second
	instanceLoadDeadline = 10 * time.Minute
)

type CustomersManagerClient struct {
	Server      string `yaml:"server"`
	InstanceID  string `yaml:"instance_id"`
	InstanceKey string `yaml:"instance_key"`
}

func (c *CustomersManagerClient) LoadInstanceConfig() error {
	deadline := time.Now().Add(instanceLoadDeadline)
	var lastErr error

	for {
		lastErr = c.tryLoadInstanceConfig()
		if lastErr == nil {
			catcher.Info("instance configuration loaded", nil)
			return nil
		}

		if time.Now().Add(instanceLoadInterval).After(deadline) {
			break
		}

		_ = catcher.Error("instance configuration load failed, retrying", lastErr, map[string]any{
			"next_attempt_in": instanceLoadInterval.String(),
		})
		time.Sleep(instanceLoadInterval)
	}

	return lastErr
}

func (c *CustomersManagerClient) tryLoadInstanceConfig() error {
	if !utils.CheckIfPathExist(instanceConfigPath) {
		return catcher.Error("instance config file not found", nil, map[string]any{
			"path": instanceConfigPath,
		})
	}

	if err := utils.ReadYAML(instanceConfigPath, c); err != nil {
		return catcher.Error("failed to read or parse instance config", err, nil)
	}

	if c.Server == "" || c.InstanceID == "" || c.InstanceKey == "" {
		return catcher.Error("missing required fields in instance config", nil, map[string]any{
			"path": instanceConfigPath,
		})
	}

	return nil
}
