package updates

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/agent/beats"
	"github.com/utmstack/UTMStack/agent/agent/config"
	"github.com/utmstack/UTMStack/agent/agent/models"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

const (
	checkEvery = 5 * time.Minute
)

func UpdateDependencies(cnf *config.Config) {
	RemoveOldServices()

	for {
		time.Sleep(checkEvery)
		versions := models.Version{}
		prepareForUpdate(&versions)

		zipDependencyResponse, err := getDependency(cnf, versions.DependenciesVersion, config.DEPEND_ZIP_LABEL)
		if err != nil {
			utils.Logger.ErrorF("error checking dependencies: %v", err)
			continue
		}

		serviceDependencyResponse, err := getDependency(cnf, versions.ServiceVersion, config.DEPEND_SERVICE_LABEL)
		if err != nil {
			utils.Logger.ErrorF("error checking service: %v", err)
			continue
		}

		err = updateDependencies(&zipDependencyResponse, cnf, &versions)
		if err != nil {
			utils.Logger.ErrorF("error updating dependencies: %v", err)
			continue
		}

		err = updateService(&serviceDependencyResponse, &versions)
		if err != nil {
			utils.Logger.ErrorF("error updating service: %v", err)
			continue
		}
	}
}

func prepareForUpdate(versions *models.Version) {
	isRecentAgentUpgradeDone := !utils.CheckIfPathExist(config.GetVersionPath())
	if isRecentAgentUpgradeDone {
		versions.ServiceVersion = "0.0.0"
		versions.DependenciesVersion = "0.0.0"
		os.RemoveAll(config.GetVersionOldPath())
	} else {
		err := utils.ReadJson(config.GetVersionPath(), versions)
		if err != nil {
			utils.Logger.ErrorF("error reading version file: %v", err)
		}
	}
}

func updateService(response *models.DependencyUpdateResponse, versions *models.Version) error {
	newUpdate, err := processUpdateResponse(response, config.GetDownloadFilePath(config.DEPEND_SERVICE_LABEL, "_new"))
	if err != nil {
		return fmt.Errorf("error processing agent service response: %v", err)
	}

	if newUpdate {
		utils.Logger.Info("New version of agent service found: %s", response.Version)
		versions.ServiceVersion = response.Version
		err = utils.WriteJSON(config.GetVersionPath(), versions)
		if err != nil {
			return fmt.Errorf("error writing version file: %v", err)
		}
		if runtime.GOOS == "linux" {
			if err = utils.Execute("chmod", utils.GetMyPath(), "-R", "777", config.GetDownloadFilePath(config.DEPEND_SERVICE_LABEL, "_new")); err != nil {
				utils.Logger.ErrorF("error executing chmod: %v", err)
			}
		}
		utils.Execute(config.GetSelfUpdaterPath(), utils.GetMyPath())
	}
	return nil
}

func updateDependencies(response *models.DependencyUpdateResponse, cnf *config.Config, versions *models.Version) error {
	newUpdate, err := processUpdateResponse(response, config.GetDownloadFilePath(config.DEPEND_ZIP_LABEL, ""))
	if err != nil {
		return fmt.Errorf("error processing dependencies response: %v", err)
	}

	if newUpdate {
		utils.Logger.Info("New version of dependencies found: %s", response.Version)
		if err = beats.UninstallBeats(); err != nil {
			return fmt.Errorf("error uninstalling beats: %v", err)
		}
		err = removeDependencies()
		if err != nil {
			return fmt.Errorf("error removing dependencies: %v", err)
		}
		err = utils.Unzip(config.GetDownloadFilePath(config.DEPEND_ZIP_LABEL, ""), utils.GetMyPath())
		if err != nil {
			return fmt.Errorf("error unzipping dependencies: %v", err)
		}
		if runtime.GOOS == "linux" {
			if err = utils.Execute("chmod", utils.GetMyPath(), "-R", "777", "utmstack_updater_self"); err != nil {
				return fmt.Errorf("error executing chmod: %v", err)
			}
		}
		err = os.Remove(config.GetDownloadFilePath(config.DEPEND_ZIP_LABEL, ""))
		if err != nil {
			utils.Logger.ErrorF("error removing dependencies file: %v", err)
		}
		err = beats.InstallBeats(*cnf)
		if err != nil {
			return fmt.Errorf("error installing beats: %v", err)
		}

		versions.DependenciesVersion = response.Version
		err = utils.WriteJSON(config.GetVersionPath(), &versions)
		if err != nil {
			return fmt.Errorf("error writing version file: %v", err)
		}
	}
	return nil
}

func processUpdateResponse(updateResponse *models.DependencyUpdateResponse, filepath string) (bool, error) {
	switch {
	case updateResponse.Message == "dependency not found", strings.Contains(updateResponse.Message, "error getting dependency file"):
		return false, fmt.Errorf("error getting dependency file: %v", updateResponse.Message)
	case updateResponse.Message == "dependency already up to date":
		return false, nil
	case updateResponse.Message == "dependency update available":
		if err := utils.WriteBytesToFile(filepath, updateResponse.FileContent); err != nil {
			return false, fmt.Errorf("error writing dependency file %s: %v", filepath, err)
		}
		return true, nil
	default:
		return false, fmt.Errorf("error processing update response: %v", updateResponse.Message)
	}
}

func removeDependencies() error {
	dependPaths := config.GetDependenPaths()
	for _, dep := range dependPaths {
		err := os.RemoveAll(dep)
		if err != nil {
			return fmt.Errorf("error removing file %s: %v", dep, err)
		}
	}
	return nil
}

func getDependency(cnf *config.Config, version, dependencyType string) (models.DependencyUpdateResponse, error) {
	queryParams := url.Values{}
	queryParams.Add("version", version)
	queryParams.Add("os", runtime.GOOS)
	queryParams.Add("type", dependencyType)

	headers := map[string]string{
		"key": cnf.AgentKey,
		"id":  fmt.Sprintf("%v", cnf.AgentID),
	}
	fmt.Printf(("Headers: %v\n"), headers)

	var tlsConfig *tls.Config
	if cnf.SkipCertValidation {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	} else {
		tlsConfig = &tls.Config{}
	}

	resp, status, err := utils.DoReq[models.DependencyUpdateResponse](fmt.Sprintf("https://%s/dependencies/agent?%s", cnf.Server, queryParams.Encode()), nil, http.MethodGet, headers, tlsConfig)
	if err != nil {
		return models.DependencyUpdateResponse{}, fmt.Errorf("error downloading dependencies: %v", err)
	}
	if status != http.StatusOK {
		return models.DependencyUpdateResponse{}, fmt.Errorf("error downloading dependencies: %v", resp.Message)
	}

	return resp, nil
}
