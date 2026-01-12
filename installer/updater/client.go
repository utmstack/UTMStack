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
		// config.Logger().Info("[DEBUG] Initializing UpdaterClient...")
		updaterClient = &UpdaterClient{
			Config:  InstanceConfig{},
			License: "",
		}

		// config.Logger().Info("[DEBUG] Checking instance config at: %s", config.InstanceConfigPath)
		if !utils.CheckIfPathExist(config.InstanceConfigPath) {
			// config.Logger().Info("[DEBUG] Instance config not found, registering instance...")
			err := RegisterInstance()
			if err != nil {
				config.Logger().ErrorF("error registering instance: %v", err)
				return
			}
			// config.Logger().Info("[DEBUG] Instance registered successfully")
		}

		// config.Logger().Info("[DEBUG] Checking license file at: %s", config.LicenseFilePath)
		if !utils.CheckIfPathExist(config.LicenseFilePath) {
			//config.Logger().Info("[DEBUG] License file not found, creating empty license file...")
			err := os.WriteFile(config.LicenseFilePath, []byte{}, 0644)
			if err != nil {
				config.Logger().ErrorF("error creating license file: %v", err)
				return
			}
		}

		cnf := InstanceConfig{}
		utils.ReadYAML(config.InstanceConfigPath, &cnf)
		updaterClient.Config = cnf
		// config.Logger().Info("[DEBUG] Instance config loaded - InstanceID=%s, Server=%s", cnf.InstanceID, cnf.Server)

		licenseBytes, err := os.ReadFile(config.LicenseFilePath)
		if err != nil {
			config.Logger().ErrorF("error reading license file: %v", err)
			return
		}

		updaterClient.License = string(licenseBytes)
		// config.Logger().Info("[DEBUG] UpdaterClient initialized successfully")
	})

	return updaterClient
}

func (c *UpdaterClient) UpdateProcess() {
	// config.Logger().Info("[DEBUG] UpdateProcess started - checking updates every %v", config.CheckUpdatesEvery)
	ticker := time.NewTicker(config.CheckUpdatesEvery)
	defer ticker.Stop()

	for range ticker.C {
		// config.Logger().Info("[DEBUG] Update tick triggered - Updating=%v, ConnectedToInternet=%v", config.Updating, config.ConnectedToInternet)
		inWindow := IsInMaintenanceWindow()
		// config.Logger().Info("[DEBUG] IsInMaintenanceWindow=%v", inWindow)

		if !config.Updating && inWindow {
			// config.Logger().Info("[DEBUG] Conditions met, calling CheckUpdate...")
			err := c.CheckUpdate()
			if err != nil {
				config.Logger().ErrorF("error checking update: %v", err)
			}
		} //else {
		// config.Logger().Info("[DEBUG] Skipping update check - Updating=%v, InMaintenanceWindow=%v", config.Updating, inWindow)
		//}
	}
}

func (c *UpdaterClient) CheckUpdate() error {
	// config.Logger().Info("[DEBUG] CheckUpdate started - Server=%s, InstanceID=%s", c.Config.Server, c.Config.InstanceID)
	updates := make([]map[string]string, 0)

	url := fmt.Sprintf("%s%s", c.Config.Server, config.GetUpdatesInfoEndpoint)
	// config.Logger().Info("[DEBUG] Checking updates from URL: %s", url)

	if config.ConnectedToInternet {
		// config.Logger().Info("[DEBUG] Connected to internet, fetching updates from server...")
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
		// config.Logger().Info("[DEBUG] Got %d updates from server", len(resp))
		for _, update := range resp {
			// config.Logger().Info("[DEBUG] Update found: ID=%s, Version=%s, Edition=%s", update.ID, update.Version.Version, update.Instance.Edition)
			newUpdate := make(map[string]string)
			newUpdate["version"] = update.Version.Version
			newUpdate["edition"] = update.Instance.Edition
			newUpdate["changelog"] = update.Version.Changelog
			newUpdate["id"] = update.ID
			updates = append(updates, newUpdate)
		}
	} else {
		// config.Logger().Info("[DEBUG] Not connected to internet, checking offline updates...")
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
	// config.Logger().Info("[DEBUG] After sorting: %d updates to apply", len(sortedUpdates))

	// if len(sortedUpdates) == 0 {
	// 	config.Logger().Info("[DEBUG] No updates to apply")
	// 	return nil
	// }

	for _, update := range sortedUpdates {
		// config.Logger().Info("[DEBUG] Processing update %d/%d: version=%s, edition=%s, id=%s",
		// 	i+1, len(sortedUpdates), update["version"], update["edition"], update["id"])
		// Apply all updates from the server regardless of current version
		// This allows for rollbacks, pre-release type changes (alpha→dev), and ensures all updates are applied in order
		// The server is responsible for only sending pending updates (marked as sent after application)
		err := c.UpdateToNewVersion(update["version"], update["edition"], update["changelog"])
		if err != nil {
			return fmt.Errorf("error updating to new version: %v", err)
		}
		if update["id"] != "offline" {
			// config.Logger().Info("[DEBUG] Marking update %s as sent", update["id"])
			err = c.MarkUpdateSent(update["id"])
			if err != nil {
				return fmt.Errorf("error marking update as sent: %v", err)
			}
		}
	}

	// config.Logger().Info("[DEBUG] All updates processed successfully")
	return nil
}

func (c *UpdaterClient) UpdateToNewVersion(version, edition, changelog string) error {
	config.Logger().Info("Updating UTMStack to version %s-%s...", version, edition)
	config.Updating = true
	defer func() {
		config.Updating = false
	}()

	// Update installer binary first (only in prod branch)
	cnf := config.GetConfig()
	if cnf.Branch == "prod" || cnf.Branch == "" {
		if err := c.UpdateInstaller(version); err != nil {
			config.Logger().ErrorF("error updating installer: %v", err)
		}
	}

	err := docker.StackUP(version + "-" + edition)
	if err != nil {
		return fmt.Errorf("error updating UTMStack: %v", err)
	}

	err = SaveVersion(version, edition, changelog)
	if err != nil {
		return fmt.Errorf("error saving new version: %v", err)
	}

	config.Logger().Info("UTMStack updated to version %s-%s", version, edition)

	time.Sleep(3 * time.Minute)

	err = utils.RunCmd("docker", "image", "prune", "-a", "-f")
	if err != nil {
		config.Logger().ErrorF("error cleaning up old Docker images after update: %v", err)
	}

	// Restart service to load new installer binary
	if cnf.Branch == "prod" || cnf.Branch == "" {
		go func() {
			time.Sleep(5 * time.Second)
			utils.RestartService("UTMStackComponentsUpdater")
		}()
	}

	return nil
}

func (c *UpdaterClient) UpdateInstaller(version string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %v", err)
	}

	// Download new installer from GitHub
	url := fmt.Sprintf(config.GitHubReleasesURL, version)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading installer from %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error downloading installer: status %d", resp.StatusCode)
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "installer-*")
	if err != nil {
		return fmt.Errorf("error creating temp file: %v", err)
	}
	tmpPath := tmpFile.Name()

	// Download to temp file
	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("error writing installer to temp file: %v", err)
	}

	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("error making installer executable: %v", err)
	}

	// Replace current binary
	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("error replacing installer binary: %v", err)
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
