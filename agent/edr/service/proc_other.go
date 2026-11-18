//go:build !linux

package service

import (
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/procwatch"
)

func newProcessWatcher(_ config.EDRConfig, h procwatch.Handler) *procwatch.Watcher {
	return procwatch.New(h)
}
