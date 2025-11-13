package updater

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
	ticker := time.NewTicker(config.CheckUpdatesEvery)
	defer ticker.Stop()

	for range ticker.C {
		if !config.Updating && IsInMaintenanceWindow() {
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
			newUpdate["id"] = update.ID
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
		newUpdate["id"] = "offline"
		updates = append(updates, newUpdate)
	}

	sortedUpdates := SortVersions(updates)

	for _, update := range sortedUpdates {
		// Apply all updates from the server regardless of current version
		// This allows for rollbacks, pre-release type changes (alpha→dev), and ensures all updates are applied in order
		// The server is responsible for only sending pending updates (marked as sent after application)
		err := c.UpdateToNewVersion(update["version"], update["edition"], update["changelog"])
		if err != nil {
			return fmt.Errorf("error updating to new version: %v", err)
		}
		if update["id"] != "offline" {
			err = c.MarkUpdateSent(update["id"])
			if err != nil {
				return fmt.Errorf("error marking update as sent: %v", err)
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

	err = utils.RunCmd("docker", "system", "prune", "-f")
	if err != nil {
		config.Logger().ErrorF("error cleaning up Docker system after update: %v", err)
	}

	return nil
}

func (c *UpdaterClient) MarkUpdateSent(updateId string) error {
	url := fmt.Sprintf("%s%s?id=%s", c.Config.Server, config.SetUpdateSentEndpoint, updateId)
	_, status, err := utils.DoReq[any](
		url,
		nil,
		http.MethodPost,
		map[string]string{"id": c.Config.InstanceID, "key": c.Config.InstanceKey},
		nil,
	)
	if err != nil || status != http.StatusOK {
		return fmt.Errorf("error marking update as sent: status: %d, error: %v", status, err)
	}
	return nil
}

func (c *UpdaterClient) UploadLogs(ctx context.Context, path string) error {
	url := fmt.Sprintf("%s%s", c.Config.Server, config.LogCollectorEndpoint)

	buf := &bytes.Buffer{}
	writer := io.MultiWriter(buf)

	zipFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	if _, err = io.Copy(writer, zipFile); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/zip")
	req.Header.Set("id", c.Config.InstanceID)
	req.Header.Set("key", c.Config.InstanceKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		if resp.StatusCode == 500 && strings.Contains(bodyStr, "log collector is not enabled for this instance") {
			return nil
		}

		return fmt.Errorf("%s: %s", resp.Status, bodyStr)
	}
	return nil
}
