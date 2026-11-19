package svc

import (
	"path"
	"strconv"
	"strings"
)

// validLaunchdName reports whether serviceName is a safe launchd label /
// systemd-style unit name: it must start with an ASCII letter or digit and
// contain only alphanumerics, dots, hyphens and underscores. This is a
// defense-in-depth guard so the plist path built from the name can never
// escape /Library/LaunchDaemons (no "..", no separators) even though all
// current callers pass trusted constants.
func validLaunchdName(serviceName string) bool {
	if serviceName == "" {
		return false
	}
	for i, r := range serviceName {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case i > 0 && (r == '.' || r == '-' || r == '_'):
		default:
			return false
		}
	}
	return true
}

// launchctlListNotFound reports whether a `launchctl list <label>` result means
// the service is not loaded. The not-loaded condition is signalled by the exit
// code 113 (macOS 12+) or, on older releases, by a "Could not find service ..."
// message; either indicates the daemon is absent, as opposed to a transient
// error (permission, launchd hiccup) that must NOT be treated as "stopped".
func launchctlListNotFound(exitCode int, out string) bool {
	if exitCode == 113 {
		return true
	}
	return strings.Contains(strings.ToLower(out), "could not find service")
}

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
