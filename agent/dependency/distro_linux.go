//go:build linux
// +build linux

package dependency

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
)

// DistroInfo holds Linux distribution details for package manager selection.
type DistroInfo struct {
	ID             string // Primary distro ID (debian, ubuntu, rhel, centos, fedora, etc.)
	IDLike         string // Parent distro family (debian, rhel, etc.)
	PackageManager string // Package manager to use (apt, dnf, yum, zypper, pacman)
}

// DetectDistro parses /etc/os-release and determines the package manager to use.
// Falls back to binary detection if os-release parsing fails.
func DetectDistro() *DistroInfo {
	info := &DistroInfo{}

	// Try parsing /etc/os-release first
	if osRelease, err := parseOSRelease("/etc/os-release"); err == nil {
		info.ID = strings.ToLower(osRelease["ID"])
		info.IDLike = strings.ToLower(osRelease["ID_LIKE"])
	}

	// Map to package manager
	info.PackageManager = mapPackageManager(info.ID, info.IDLike)

	// Fallback to binary detection if mapping failed
	if info.PackageManager == "" {
		info.PackageManager = detectPackageManagerBinary()
	}

	return info
}

// parseOSRelease reads and parses /etc/os-release into a key-value map.
func parseOSRelease(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := strings.Trim(parts[1], `"'`)
		result[key] = value
	}

	return result, scanner.Err()
}

// mapPackageManager determines the package manager based on distro ID and ID_LIKE.
// Returns empty string if no match is found.
func mapPackageManager(id, idLike string) string {
	// Direct ID mapping (highest priority)
	switch id {
	// apt-based (Debian family)
	case "debian", "ubuntu", "linuxmint", "pop", "elementary", "zorin", "kali", "raspbian":
		return "apt"

	// dnf-based (modern Red Hat family)
	case "fedora", "rocky", "almalinux", "alma":
		return "dnf"

	// yum/dnf for RHEL and CentOS (depends on version, but dnf is preferred on modern)
	case "rhel", "centos":
		// Modern RHEL/CentOS use dnf, but we try dnf first in installPackage
		return "dnf"

	// yum-based (legacy Red Hat, Amazon Linux)
	case "amzn", "amazon":
		return "yum"

	// zypper-based (SUSE family)
	case "suse", "opensuse", "opensuse-leap", "opensuse-tumbleweed":
		return "zypper"

	// pacman-based (Arch family)
	case "arch", "manjaro", "endeavouros", "garuda":
		return "pacman"
	}

	// Check ID_LIKE for family membership
	idLikeParts := strings.Fields(idLike)
	for _, like := range idLikeParts {
		switch like {
		case "debian", "ubuntu":
			return "apt"
		case "rhel", "fedora", "centos":
			return "dnf"
		case "suse", "opensuse":
			return "zypper"
		case "arch":
			return "pacman"
		}
	}

	return ""
}

// detectPackageManagerBinary tries to find package manager binaries as a fallback.
// Returns empty string if no package manager is found.
func detectPackageManagerBinary() string {
	// Check in order of preference
	binaries := []struct {
		cmd string
		pm  string
	}{
		{"apt-get", "apt"},
		{"dnf", "dnf"},
		{"yum", "yum"},
		{"zypper", "zypper"},
		{"pacman", "pacman"},
	}

	for _, b := range binaries {
		if _, err := exec.LookPath(b.cmd); err == nil {
			return b.pm
		}
	}

	return ""
}
