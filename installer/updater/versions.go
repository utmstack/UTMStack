package updater

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
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
				version.Version = config.INSTALLER_VERSION
				version.Changelog = ""
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

	// Regex pattern to find versions like 11_0_0
	versionRegex := regexp.MustCompile(`-(\d+_\d+_\d+)-enterprise\.tar$`)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasPrefix(name, "utmstack-") && strings.HasSuffix(name, "-enterprise.tar") {
			matches := versionRegex.FindStringSubmatch(name)
			if len(matches) >= 2 {
				version := strings.ReplaceAll(matches[1], "_", ".")
				return version, nil
			}
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

// Version sorting functions
type Version struct {
	Major          int
	Minor          int
	Patch          int
	PrereleaseName string // alpha, beta, rc, or empty
	PrereleaseNum  int    // the number after alpha.X, beta.X, rc.X
	Original       string
}

func ParseVersion(versionStr string) Version {
	v := Version{Original: versionStr}

	// Remove v prefix if present
	versionStr = strings.TrimPrefix(versionStr, "v")
	versionStr = strings.TrimPrefix(versionStr, "V")

	// Parse version with regex: X.Y.Z or X.Y.Z-type.num
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-(alpha|beta|rc)\.(\d+))?$`)
	matches := re.FindStringSubmatch(versionStr)

	if len(matches) > 1 {
		v.Major, _ = strconv.Atoi(matches[1])
	}
	if len(matches) > 2 {
		v.Minor, _ = strconv.Atoi(matches[2])
	}
	if len(matches) > 3 {
		v.Patch, _ = strconv.Atoi(matches[3])
	}
	if len(matches) > 4 && matches[4] != "" {
		v.PrereleaseName = matches[4]
	}
	if len(matches) > 5 && matches[5] != "" {
		v.PrereleaseNum, _ = strconv.Atoi(matches[5])
	}

	return v
}

func CompareVersions(v1, v2 Version) int {
	// Compare major.minor.patch first
	if v1.Major != v2.Major {
		return v1.Major - v2.Major
	}
	if v1.Minor != v2.Minor {
		return v1.Minor - v2.Minor
	}
	if v1.Patch != v2.Patch {
		return v1.Patch - v2.Patch
	}

	if v1.PrereleaseNum != v2.PrereleaseNum {
		return v1.PrereleaseNum - v2.PrereleaseNum
	}

	return 0
}

func SortVersions(versions []map[string]string) []map[string]string {
	if len(versions) <= 1 {
		return versions
	}

	for i := range len(versions) - 1 {
		for j := range len(versions) - i - 1 {
			v1 := ParseVersion(versions[j]["version"])
			v2 := ParseVersion(versions[j+1]["version"])

			if CompareVersions(v1, v2) > 0 {
				versions[j], versions[j+1] = versions[j+1], versions[j]
			}
		}
	}

	return versions
}
