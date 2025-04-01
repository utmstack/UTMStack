package updater

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/docker"
	"github.com/utmstack/UTMStack/installer/utils"
)

type UpdaterClient struct {
	Config InstanceConfig
}

var (
	updaterClient     *UpdaterClient
	updaterClientOnce sync.Once
)

func GetUpdaterClient() *UpdaterClient {
	updaterClientOnce.Do(func() {
		updaterClient = &UpdaterClient{
			Config: InstanceConfig{},
		}

		cnf := InstanceConfig{}
		utils.ReadYAML(config.InstanceConfigPath, &cnf)
		updaterClient.Config = cnf
	})

	return updaterClient
}

func (c *UpdaterClient) UpdateProcess() {
	for {
		time.Sleep(config.CheckUpdatesEvery)
		if IsInMaintenanceWindow() {
			err := c.CheckUpdate(false)
			if err != nil {
				config.Logger().ErrorF("error checking update: %v", err)
			}
		}
	}
}

func (c *UpdaterClient) CheckUpdate(wait bool) error {
	if wait {
		time.Sleep(time.Second * 20)
	}
	resp, status, err := utils.DoReq[[]UpdateDTO](
		c.Config.Server+config.GetUpdatesInfoEndpoint,
		nil,
		http.MethodGet,
		map[string]string{"id": c.Config.InstanceID, "key": c.Config.InstanceKey},
	)
	if err != nil || status != http.StatusOK {
		return fmt.Errorf("error getting updates: status: %d, error: %v", status, err)
	}

	currentVersion := GetVersion()
	for _, update := range resp {
		if update.Version.Version != currentVersion.Version || update.Instance.Edition != currentVersion.Edition {
			fmt.Printf("Updating UTMStack to version %s-%s...\n", update.Version.Version, update.Instance.Edition)
			config.Logger().Info("Updating UTMStack to version %s-%s...", update.Version.Version, update.Instance.Edition)
			err := docker.StackUP(update.Version.Version + "-" + update.Instance.Edition)
			if err != nil {
				return fmt.Errorf("error updating UTMStack: %v", err)
			}
		}

		err := SaveVersion(update)
		if err != nil {
			return fmt.Errorf("error saving new version: %v", err)
		}
		config.Logger().Info("UTMStack updated to version %s-%s", update.Version.Version, update.Instance.Edition)
	}

	return nil
}
