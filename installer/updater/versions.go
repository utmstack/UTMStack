package updater

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

var (
	version     = VersionFile{}
	versionOnce sync.Once
)

func GetVersion() (VersionFile, error) {
	var err error
	versionOnce.Do(func() {
		if !utils.CheckIfPathExist(config.VersionFilePath) {
			if config.ConnectedToInternet {
				resp, status, errB := utils.DoReq[VersionDTO](config.GetCMServer()+config.GetLatestVersionEndpoint, nil, http.MethodGet, nil, nil)
				if errB != nil || status != http.StatusOK {
					err = fmt.Errorf("error getting latest version: status: %d, error: %v", status, err)
					return
				}
				version.Version = resp.Version
				version.Changelog = resp.Changelog
				version.Edition = "community"

			} else {
				versionFromTar, errB := ExtractVersionFromFolder(config.ImagesPath)
				if errB == nil {
					version.Version = versionFromTar
					version.Edition = "enterprise"
				} else {
					err = fmt.Errorf("error extracting version from folder: %v", err)
					return
				}
			}

			errB := utils.WriteJSON(config.VersionFilePath, &version)
			if errB != nil {
				err = fmt.Errorf("error writing version file: %v", err)
				return
			}
		} else {
			errB := utils.ReadJson(config.VersionFilePath, &version)
			if errB != nil {
				err = fmt.Errorf("error reading version file: %v", err)
				return
			}
		}
	})

	return version, err
}

func SaveVersion(vers, edition, changelog string) error {
	version.Changelog = changelog
	version.Edition = edition
	version.Version = vers

	return utils.WriteJSON(config.VersionFilePath, &version)
}

func ExtractVersionFromFolder(folder string) (string, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return "", fmt.Errorf("error reading directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasPrefix(name, "utmstack-") && strings.HasSuffix(name, "-enterprise.tar") {
			base := strings.TrimPrefix(name, "utmstack-")
			base = strings.TrimSuffix(base, "-enterprise.tar")

			version := strings.ReplaceAll(base, "_", ".")

			return version, nil
		}
	}

	return "", fmt.Errorf("valid version not found in folder")
}

func IsEnterpriseImage(serviceName string) (bool, error) {
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer

	cmd := exec.Command("docker", "service", "inspect", serviceName, "--format", "{{.Spec.TaskTemplate.ContainerSpec.Image}}")
	cmd.Env = os.Environ()
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("error running docker inspect: %v - %s", err, errBuf.String())
	}

	image := strings.TrimSpace(outBuf.String())
	return strings.Contains(image, "-enterprise"), nil
}
