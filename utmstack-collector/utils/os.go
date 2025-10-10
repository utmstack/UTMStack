package utils

import (
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/elastic/go-sysinfo"
)

type OSInfo struct {
	Hostname       string
	OsType         string
	Platform       string
	CurrentUser    string
	Mac            string
	OsMajorVersion string
	OsMinorVersion string
	Aliases        string
	Addresses      string
}

func GetOsInfo() (OSInfo, error) {
	var info OSInfo

	hostInfo, err := sysinfo.Host()
	if err != nil {
		return info, Logger.ErrorF("error getting host info: %v", err)
	}
	info.OsType = hostInfo.Info().OS.Type
	info.Platform = hostInfo.Info().OS.Platform
	info.Mac = strings.Join(hostInfo.Info().MACs, ",")
	info.OsMajorVersion = strconv.Itoa(hostInfo.Info().OS.Major)
	info.OsMinorVersion = strconv.Itoa(hostInfo.Info().OS.Minor)
	info.Addresses = strings.Join(hostInfo.Info().IPs, ",")

	hostName, err := os.Hostname()
	if err != nil {
		return info, Logger.ErrorF("error getting hostname: %v", err)
	}
	info.Hostname = hostName

	currentUser, err := user.Current()
	if err != nil {
		return info, Logger.ErrorF("error getting user: %v", err)
	}
	info.CurrentUser = currentUser.Username

	aliases, err := GetHostAliases(hostInfo.Info().Hostname)
	if err != nil {
		aliases = aliases[:0]
		aliases = append(aliases, "")
	}
	if len(aliases) == 1 && strings.Contains(aliases[0], "any") {
		aliases = aliases[:0]
		aliases = append(aliases, "")
	}
	info.Aliases = strings.Join(aliases, ",")

	return info, nil
}
