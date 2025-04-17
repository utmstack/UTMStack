package updater

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/docker"
	"github.com/utmstack/UTMStack/installer/utils"
)

type UpdaterClient struct {
	Config  InstanceConfig
	License string
}

var (
	updaterClient     *UpdaterClient
	updaterClientOnce sync.Once
)

func GetUpdaterClient() *UpdaterClient {
	updaterClientOnce.Do(func() {
		updaterClient = &UpdaterClient{
			Config:  InstanceConfig{},
			License: "",
		}

		if !utils.CheckIfPathExist(config.InstanceConfigPath) {
			err := RegisterInstance()
			if err != nil {
				config.Logger().ErrorF("error registering instance: %v", err)
				return
			}
		}

		if !utils.CheckIfPathExist(config.LicenseFilePath) {
			err := os.WriteFile(config.LicenseFilePath, []byte{}, 0644)
			if err != nil {
				config.Logger().ErrorF("error creating license file: %v", err)
				return
			}
		}

		cnf := InstanceConfig{}
		utils.ReadYAML(config.InstanceConfigPath, &cnf)
		updaterClient.Config = cnf

		licenseBytes, err := os.ReadFile(config.LicenseFilePath)
		if err != nil {
			config.Logger().ErrorF("error reading license file: %v", err)
			return
		}

		updaterClient.License = string(licenseBytes)
	})

	return updaterClient
}

func (c *UpdaterClient) UpdateProcess() {
	for {
		time.Sleep(config.CheckUpdatesEvery)
		if IsInMaintenanceWindow() {
			err := c.CheckUpdate()
			if err != nil {
				config.Logger().ErrorF("error checking update: %v", err)
			}
		}
	}
}

func (c *UpdaterClient) CheckUpdate() error {
	updates := make([]map[string]string, 0)

	url := fmt.Sprintf("%s%s", c.Config.Server, config.GetUpdatesInfoEndpoint)
	if config.ConnectedToInternet {
		resp, status, err := utils.DoReq[[]UpdateDTO](
			url,
			nil,
			http.MethodGet,
			map[string]string{"id": c.Config.InstanceID, "key": c.Config.InstanceKey},
			nil,
		)
		if err != nil || status != http.StatusOK {
			return fmt.Errorf("error getting updates from %s: status: %d, error: %v", url, status, err)
		}
		for _, update := range resp {
			newUpdate := make(map[string]string)
			newUpdate["version"] = update.Version.Version
			newUpdate["edition"] = update.Instance.Edition
			newUpdate["changelog"] = update.Version.Changelog
			updates = append(updates, newUpdate)
		}
	} else {
		v, err := ExtractVersionFromFolder(config.ImagesPath)
		if err != nil {
			return fmt.Errorf("error extracting version from folder: %v", err)
		}
		newUpdate := make(map[string]string)
		newUpdate["version"] = v
		newUpdate["edition"] = "enterprise"
		newUpdate["changelog"] = "No changelog available for offline version"
		updates = append(updates, newUpdate)
	}

	currentVersion, err := GetVersion()
	if err != nil {
		return fmt.Errorf("error getting current version: %v", err)
	}

	for _, update := range updates {
		if update["version"] != currentVersion.Version {
			err := c.UpdateToNewVersion(update["version"], update["edition"], update["changelog"])
			if err != nil {
				return fmt.Errorf("error updating to new version: %v", err)
			}
		}
	}

	return nil
}

func (c *UpdaterClient) UpdateToNewVersion(version, edition, changelog string) error {
	config.Logger().Info("Updating UTMStack to version %s-%s...", version, edition)
	config.Updating = true

	err := docker.StackUP(version + "-" + edition)
	if err != nil {
		return fmt.Errorf("error updating UTMStack: %v", err)
	}

	err = SaveVersion(version, edition, changelog)
	if err != nil {
		return fmt.Errorf("error saving new version: %v", err)
	}

	config.Logger().Info("UTMStack updated to version %s-%s", version, edition)
	config.Updating = false

	return nil
}
