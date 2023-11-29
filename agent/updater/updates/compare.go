package updates

import (
	"sort"
	"strconv"
	"strings"
)

func GetLatestVersion(versions DataVersions, masterVersions string) Version {
	latestVersions := Version{}

	sort.Slice(versions.Versions, func(i, j int) bool {
		s1 := strings.Split(versions.Versions[i].MasterVersion, ".")
		s2 := strings.Split(versions.Versions[j].MasterVersion, ".")
		for k := 0; k < 3; k++ {
			if s1[k] > s2[k] {
				return true
			} else if s1[k] < s2[k] {
				return false
			}
		}
		return true
	})

	for _, vers := range versions.Versions {
		if isVersionGreaterOrEqual(vers.MasterVersion, masterVersions) {
			latestVersions = vers
			break
		}
	}

	return latestVersions
}

func isVersionGreaterOrEqual(version1 string, version2 string) bool {
	v1Parts := strings.Split(version1, ".")
	v2Parts := strings.Split(version2, ".")

	for i := 0; i < len(v1Parts) && i < len(v2Parts); i++ {
		v1, _ := strconv.Atoi(v1Parts[i])
		v2, _ := strconv.Atoi(v2Parts[i])

		if v1 > v2 {
			return true
		} else if v1 < v2 {
			return false
		}
	}

	return true
}
