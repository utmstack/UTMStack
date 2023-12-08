package depend

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/utmstack/UTMStack/agent/runner/configuration"
	"github.com/utmstack/UTMStack/agent/runner/utils"
)

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

func getCurrentVersion(ip string, env string, skip string) (Version, error) {
	currentVersion := Version{}

	// Get master version
	skipB := skip == "yes"
	mastVers, err := getMasterVersion(ip, skipB)
	if err != nil {
		return currentVersion, fmt.Errorf("error getting master version: %v", err)
	}

	path, err := utils.GetMyPath()
	if err != nil {
		return currentVersion, fmt.Errorf("failed to get current path: %v", err)
	}

	err = utils.DownloadFile(configuration.Bucket+env+"/versions.json?time="+utils.GetCurrentTime(), filepath.Join(path, "versions.json"))
	if err != nil {
		return currentVersion, fmt.Errorf("error downloading versions.json: %v", err)
	}

	// Save data from versions.json
	var dataVersions DataVersions
	err = utils.ReadJson(filepath.Join(path, "versions.json"), &dataVersions)
	if err != nil {
		return currentVersion, fmt.Errorf("error reading versions.json: %v", err)
	}

	versionExist := false
	for _, vers := range dataVersions.Versions {
		versParts := strings.Split(vers.MasterVersion, ".")
		masterParts := strings.Split(mastVers, ".")

		if versParts[0] == masterParts[0] && versParts[1] == masterParts[1] {
			versionExist = true
			currentVersion = vers
		}
	}

	if versionExist {
		err = utils.WriteJSON(filepath.Join(path, "versions.json"), &currentVersion)
		if err != nil {
			return currentVersion, fmt.Errorf("error writing versions.json: %v", err)
		}
		return currentVersion, nil
	}
	return currentVersion, fmt.Errorf("error: master version not exist in versions.json")
}
