package svc

import (
	"path"
	"strconv"
	"strings"
)

// daemonPlistPath returns the launch daemon plist path for a service installed
// by kardianos/service with the default options (optionUserService=false),
// i.e. a system daemon in /Library/LaunchDaemons/<name>.plist. launchctl
// load/unload take this path; `launchctl list <label>` takes the bare service
// name.
func daemonPlistPath(serviceName string) string {
	return path.Join("/Library/LaunchDaemons", serviceName+".plist")
}

// plistRunningPID reports whether a `launchctl list <label>` output shows the
// daemon currently running, i.e. contains a PID. Both output shapes seen
// across macOS versions are handled: the table form (PID in the first column):
//
//	PID	Status	LastExitCode
//	12345	0
//
// and the dictionary form printed by older releases (a quoted "PID" key):
//
//	{
//		"PID" = 12345;
//	}
//
// When the daemon is registered but not running, the table's first column is
// "-" and the dictionary form has no "PID" key, so both report false. launchctl
// also prints a plain "Could not find service ..." message for a daemon that
// is not loaded, which reports false too.
func plistRunningPID(out string) bool {
	if strings.Contains(out, `"PID" = `) {
		return true
	}
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if n, err := strconv.Atoi(fields[0]); err == nil && n > 0 {
			return true
		}
	}
	return false
}
