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
	resp, status, err := utils.DoReq[[]Release](
		c.Config.Server+config.GetUpdatesInfoEndpoint,
		nil,
		http.MethodGet,
		map[string]string{"id": c.Config.InstanceID, "key": c.Config.InstanceKey},
	)
	if err != nil || status != http.StatusOK {
		return fmt.Errorf("error getting updates: status: %d, error: %v", status, err)
	}

	currentVersion := GetVersion()
	for _, release := range resp {
		if release.Version != currentVersion.Version || release.Edition != currentVersion.Edition {
			fmt.Printf("Updating UTMStack to version %s-%s...\n", release.Version, release.Edition)
			config.Logger().Info("Updating UTMStack to version %s-%s...", release.Version, release.Edition)
			err := docker.StackUP(release.Version + "-" + release.Edition)
			if err != nil {
				return fmt.Errorf("error updating UTMStack: %v", err)
			}
		}

		err := SaveVersion(release)
		if err != nil {
			return fmt.Errorf("error saving new version: %v", err)
		}
		config.Logger().Info("UTMStack updated to version %s-%s", release.Version, release.Edition)
	}

	return nil
}
