package utils

import (
	"fmt"
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

func GetOsInfo() (*OSInfo, error) {
	var info OSInfo

	hostInfo, err := sysinfo.Host()
	if err != nil {
		return nil, fmt.Errorf("error getting host info: %v", err)
	}

	hostName, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("error getting hostname: %v", err)
	}

	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("error getting user: %v", err)
	}

	aliases, err := GetHostAliases(hostInfo.Info().Hostname)
	if err != nil {
		aliases = aliases[:0]
		aliases = append(aliases, "")
	}

	info.Hostname = hostName
	info.OsType = hostInfo.Info().OS.Type
	info.Platform = hostInfo.Info().OS.Platform
	info.CurrentUser = currentUser.Username
	info.Mac = strings.Join(hostInfo.Info().MACs, ",")
	info.OsMajorVersion = strconv.Itoa(hostInfo.Info().OS.Major)
	info.OsMinorVersion = strconv.Itoa(hostInfo.Info().OS.Minor)
	if len(aliases) == 1 && strings.Contains(aliases[0], "any") {
		aliases = aliases[:0]
		aliases = append(aliases, "")
	}
	info.Aliases = strings.Join(aliases, ",")
	info.Addresses = strings.Join(hostInfo.Info().IPs, ",")

	return &info, nil
}
