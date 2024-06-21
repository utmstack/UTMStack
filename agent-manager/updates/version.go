package updates

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/util"
)

func GetMasterVersion() (string, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	resp, status, err := util.DoReq[models.InfoResponse](config.GetPanelServiceName()+config.MASTERVERSIONENDPOINT, nil, http.MethodGet, map[string]string{}, tlsConfig)
	if err != nil {
		return "", err
	} else if status != http.StatusOK {
		return "", fmt.Errorf("status code %d: %v", status, resp)
	}
	return resp.Build.Version, nil
}

func FindLatestVersion(versions models.DataVersions, masterVersion string) models.Version {
	for _, vers := range versions.Versions {
		if vers.MasterVersion == masterVersion {
			return vers
		}
	}
	return models.Version{}
}

func IsVersionGreater(oldVersion, newVersion string) bool {
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
