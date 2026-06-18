package updater

import (
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
	ticker := time.NewTicker(config.CheckUpdatesEvery)
	defer ticker.Stop()

	for range ticker.C {
		if err := c.CheckLicense(); err != nil {
			config.Logger().ErrorF("error checking license: %v", err)
		}
	}
}

func (c *UpdaterClient) CheckLicense() error {
	newLicense := ""

	url := fmt.Sprintf("%s%s", c.Config.Server, config.GetLicenseEndpoint)
	if config.ConnectedToInternet {
		resp, status, err := utils.DoReq[string](
			url,
			nil,
			http.MethodGet,
			map[string]string{"id": c.Config.InstanceID, "key": c.Config.InstanceKey},
			nil,
		)
		if err != nil || status != http.StatusOK {
			return fmt.Errorf("error getting license from %s: status: %d, error: %v", url, status, err)
		}
		newLicense = resp
	} else if utils.CheckIfPathExist(config.LicenseFilePath) {
		newLicenseBytes, err := os.ReadFile(config.LicenseFilePath)
		if err != nil {
			return fmt.Errorf("error reading license file: %v", err)
		}

		newLicense = string(newLicenseBytes)
	}

	if newLicense != "" && newLicense != c.License {
		decryptedLicense, err := lm.DecryptAndVerifyFromBase64(newLicense, []string{c.Config.InstanceID, config.REPLACE}, config.PUBLIC_KEY)
		if err != nil {
			return fmt.Errorf("error decrypting and verifying license: %v", err)
		}

		finalLicense := LicenseEncrypted{}
		err = json.Unmarshal([]byte(decryptedLicense), &finalLicense)
		if err != nil {
			return fmt.Errorf("error unmarshalling decrypted license: %v", err)
		}

		if time.Now().After(finalLicense.ExpiresAt) {
			config.Logger().ErrorF("license has expired on %s, please renew it", finalLicense.ExpiresAt.Format(time.RFC3339))
			os.Remove(config.LicenseFilePath)
			c.License = ""

			err = SaveVersion("", "community", "")
			if err != nil {
				return fmt.Errorf("error saving version after license update: %v", err)
			}

			return nil
		}

		err = SaveVersion("", "enterprise", "")
		if err != nil {
			return fmt.Errorf("error saving version after license update: %v", err)
		}

		c.License = newLicense

		err = os.WriteFile(config.LicenseFilePath, []byte(newLicense), 0644)
		if err != nil {
			return fmt.Errorf("error writing new license: %v", err)
		}

		config.Logger().Info("License updated successfully")
	}

	return nil
}
