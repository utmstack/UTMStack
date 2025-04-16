package updater

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"

	lm "github.com/utmstack/license-manager-sdk"
)

func (c *UpdaterClient) LicenseProcess() {
	for {
		time.Sleep(config.CheckUpdatesEvery)
		if IsInMaintenanceWindow() && !config.Updating {
			err := c.CheckLicense()
			if err != nil {
				config.Logger().ErrorF("error checking license: %v", err)
			}
		}
	}
}

func (c *UpdaterClient) CheckLicense() error {
	newLicense := ""

	if config.ConnectedToInternet {
		resp, status, err := utils.DoReq[string](
			c.Config.Server+config.GetLicenseEndpoint,
			nil,
			http.MethodGet,
			map[string]string{"id": c.Config.InstanceID, "key": c.Config.InstanceKey},
		)
		if err != nil || status != http.StatusOK {
			return fmt.Errorf("error getting license: status: %d, error: %v", status, err)
		}
		newLicense = resp
	} else {
		if !utils.CheckIfPathExist(config.InstanceConfigPath) {
			auth, err := getInstanceAuthFromBackend()
			if err != nil {
				return fmt.Errorf("error getting instance auth: %v", err)
			}

			if auth == "" {
				return fmt.Errorf("instance auth is empty, since you are not connected to the internet, please contact support to get your license")
			}

			authB, err := base64.StdEncoding.DecodeString(auth)
			if err != nil {
				return fmt.Errorf("error decoding instance auth: %v", err)
			}

			newAuth := Auth{}
			err = json.Unmarshal(authB, &newAuth)
			if err != nil {
				return fmt.Errorf("error unmarshalling instance auth: %v", err)
			}

			c.Config.InstanceID = newAuth.ID
			c.Config.InstanceKey = newAuth.Key
			c.Config.Server = config.GetCMServer()
			err = utils.WriteYAML(config.InstanceConfigPath, c.Config)
			if err != nil {
				return fmt.Errorf("error writing instance config file: %v", err)
			}
		}

		license, err := getLicenseFromBackend()
		if err != nil {
			return fmt.Errorf("error getting license from backend: %v", err)
		}
		newLicense = license
	}

	if newLicense != "" && newLicense != c.License {
		config.Logger().Info("Updating license...")
		err := os.WriteFile(config.LicenseFilePath, []byte(newLicense), 0644)
		if err != nil {
			return fmt.Errorf("error writing new license: %v", err)
		}
		c.License = newLicense

		decryptedLicense, err := lm.DecryptAndVerifyFromBase64(c.License, []string{c.Config.InstanceID, config.REPLACE}, config.PUBLIC_KEY)
		if err != nil {
			return fmt.Errorf("error decrypting and verifying license: %v", err)
		}

		finalLicense := LicenseEncrypted{}
		err = json.Unmarshal([]byte(decryptedLicense), &finalLicense)
		if err != nil {
			return fmt.Errorf("error unmarshalling decrypted license: %v", err)
		}

		isCurrentlyEnterprise, err := IsEnterpriseImage("utmstack_event-processor-manager")
		if err != nil {
			return fmt.Errorf("error checking if image is enterprise: %v", err)
		}

		v, err := GetVersion()
		if err != nil {
			return fmt.Errorf("error getting version: %v", err)
		}

		if time.Now().After(finalLicense.ExpiresAt) {
			config.Logger().ErrorF("license has expired on %s, please renew it", finalLicense.ExpiresAt.Format(time.RFC3339))
			// TODO: send notification to backend and email to user
			if isCurrentlyEnterprise {
				config.Logger().Info("downgrading to community edition after license expiration")
				err := c.UpdateToNewVersion(v.Version, "community", v.Changelog)
				if err != nil {
					return fmt.Errorf("error updating to new version: %v", err)
				}
			}
		} else if time.Now().After(finalLicense.ExpiresAt.Add(-72 * time.Hour)) {
			config.Logger().Info("license will expire on %s, please renew it", finalLicense.ExpiresAt.Format(time.RFC3339))
			// TODO: send notification to backend and email to user
		}

		if !isCurrentlyEnterprise {
			config.Logger().Info("Upgrading to UTMStack to enterprise after license verification...")
			err := c.UpdateToNewVersion(v.Version, "enterprise", v.Changelog)
			if err != nil {
				return fmt.Errorf("error updating to new version: %v", err)
			}
		}

		config.Logger().Info("License updated successfully")
	}

	return nil
}
