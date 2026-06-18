package updater

import (
	"fmt"
	"io"
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

		go PollAndUpdateAdminEmail(cnf)
		go StartHeartbeat(cnf)

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
		if err := c.CheckUpdate(); err != nil {
			config.Logger().ErrorF("error checking update: %v", err)
		}
	}
}

func (c *UpdaterClient) CheckUpdate() error {
	var update *PendingUpdate

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

		if len(resp) > 0 {
			// CM returns only one update at a time (the next one to apply)
			u := resp[0]
			update = &PendingUpdate{
				ID:          u.ID,
				Version:     u.Version.Version,
				Edition:     u.Instance.Edition,
				Changelog:   u.Version.Changelog,
				UpdateLocks: u.UpdateLocks,
			}
		}
	} else {
		v, err := ExtractVersionFromFolder(config.ImagesPath)
		if err != nil {
			return fmt.Errorf("error extracting version from folder: %v", err)
		}
		update = &PendingUpdate{
			ID:        "offline",
			Version:   v,
			Edition:   "enterprise",
			Changelog: "No changelog available for offline version",
		}
	}

	if update == nil {
		return nil
	}

	config.Logger().Info("Update available: %s-%s", update.Version, update.Edition)

	// Save pending update
	if err := SavePendingUpdate(*update); err != nil {
		return fmt.Errorf("error saving pending update: %v", err)
	}

	// Remove locks if provided
	if update.UpdateLocks != "" {
		stack := docker.GetStackConfig()
		if err := utils.RemoveLocks(update.UpdateLocks, stack.LocksDir); err != nil {
			config.Logger().ErrorF("error removing locks: %v", err)
		}
		config.Logger().Info("Removed locks: %s", update.UpdateLocks)
	}

	// Download the installer for this version
	cnf := config.GetConfig()
	if cnf.Branch == "prod" {
		if err := c.UpdateInstaller(update.Version); err != nil {
			return fmt.Errorf("error updating installer: %v", err)
		}
	}

	// Save the version
	if err := SaveVersion(update.Version, update.Edition, update.Changelog); err != nil {
		return fmt.Errorf("error saving new version: %v", err)
	}

	config.Logger().Info("Update prepared, restarting service to apply changes...")

	// Restart service - Apply will run on startup and mark as sent after success
	go func() {
		time.Sleep(5 * time.Second)
		utils.RestartService("UTMStackComponentsUpdater")
	}()

	return nil
}

func (c *UpdaterClient) UpdateInstaller(version string) error {
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

	// Replace binary at standard location
	if err := os.Rename(tmpPath, config.InstallerBinPath); err != nil {
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
