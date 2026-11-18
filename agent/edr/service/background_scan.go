package service

import (
	"runtime"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func backgroundScanConfig(cfg config.EDRConfig, engineThreads int) (config.EDRConfig, int) {
	return backgroundScanConfigForPlatform(cfg, engineThreads, runtime.GOOS)
}

// File/process scans share one scanner admission gate. On Linux, leave one of
// the owner's existing engine slots available to synchronous permission scans,
// which call the engine directly. The engine's total ceiling is unchanged.
// With only one engine slot there is no capacity to reserve. The returned
// status limit is zero when no reservation policy applies.
func backgroundScanConfigForPlatform(cfg config.EDRConfig, engineThreads int, platform string) (config.EDRConfig, int) {
	if platform != "linux" || !cfg.Enabled || !cfg.Sensors.FileWatcherOn() {
		return cfg, 0
	}
	f := cfg.LinuxProtection.Effective().Fanotify
	if !f.Enabled || (f.Mode != "permission_open" && f.Mode != "permission_exec") {
		return cfg, 0
	}
	total := max(1, min(16, engineThreads))
	if total < 2 {
		return cfg, 0
	}
	limit := total - 1
	if cfg.ScanConcurrency > 0 {
		limit = min(limit, cfg.ScanConcurrency)
	}
	cfg.ScanConcurrency = limit
	return cfg, limit
}
