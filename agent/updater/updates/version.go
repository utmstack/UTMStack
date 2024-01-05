package updates

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/utmstack/UTMStack/agent/updater/configuration"
	"github.com/utmstack/UTMStack/agent/updater/utils"
)

// getLatestVersion returns the latest version of the service
func getLatestVersion(versions DataVersions, masterVersions string) Version {
	for _, vers := range versions.Versions {
		if vers.MasterVersion == masterVersions {
			return vers
		}
	}
	return Version{}
}

// getMasterVersion returns the master version of the service
func getMasterVersion(ip string, skip bool) (string, error) {
	config := &tls.Config{InsecureSkipVerify: skip}
	if !skip {
		var err error
		config, err = utils.LoadTLSCredentials(configuration.GetCertPath())
		if err != nil {
			return "", fmt.Errorf("error loading tls credentials: %v", err)
		}
	}
	resp, status, err := utils.DoReq[InfoResponse]("https://"+ip+configuration.MASTERVERSIONENDPOINT, nil, http.MethodGet, map[string]string{}, config)
	if err != nil {
		return "", err
	} else if status != http.StatusOK {
		return "", fmt.Errorf("status code %d: %v", status, resp)
	}
	return resp.Build.Version, nil
}

// isVersionGreater returns true if newVersion is greater than oldVersion
func isVersionGreater(oldVersion, newVersion string) bool {
	oldParts := strings.Split(oldVersion, ".")
	newParts := strings.Split(newVersion, ".")

	for i, oldPart := range oldParts {
		nOld, _ := strconv.Atoi(oldPart)
		if i < len(newParts) {
			nNew, _ := strconv.Atoi(newParts[i])
			if nNew > nOld {
				return true
			} else if nNew < nOld {
				return false
			}
		}
	}
	return false
}

/*
func isNewOrEqualVersion(oldVersion, newVersion string) bool {
	oldParts := strings.Split(oldVersion, ".")
	newParts := strings.Split(newVersion, ".")

	minLength := len(oldParts)
	if len(newParts) < minLength {
		minLength = len(newParts)
	}

	for i := 0; i < minLength; i++ {
		nOld, _ := strconv.Atoi(oldParts[i])
		nNew, _ := strconv.Atoi(newParts[i])

		if nNew > nOld {
			return true
		} else if nNew < nOld {
			return false
		}
	}

	return len(newParts) >= len(oldParts)
}
*/
