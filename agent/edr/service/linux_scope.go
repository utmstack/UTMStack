package service

import (
	"github.com/utmstack/UTMStack/agent/edr/config"
	"path"
	"strings"
)

func linuxCanaryScope(cfg config.EDRConfig) []string {
	roots := cfg.LinuxProtection.Fanotify.Roots
	dirs := append([]string(nil), roots...)
	for _, dir := range cfg.Ransomware.CanaryDirs {
		for _, root := range roots {
			if dir == root || root == "/" && strings.HasPrefix(dir, "/") || strings.HasPrefix(path.Clean(dir), root+"/") {
				dirs = append(dirs, dir)
				break
			}
		}
	}
	return dirs
}
