//go:build linux

package service

import (
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/procwatch"
)

func newProcessWatcher(cfg config.EDRConfig, h procwatch.Handler) *procwatch.Watcher {
	return procwatch.NewWithMode(h, cfg.LinuxProtection.Effective().ProcessEvents)
}
