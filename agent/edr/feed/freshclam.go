package feed

import (
	"context"
	"path/filepath"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/shared/exec"
	"github.com/utmstack/UTMStack/shared/logger"
)

type Feed struct {
	cfg          config.EDRConfig
	cache        *cache.Cache
	runUpdater   func() error
	currentSigDB func() (string, error)
}

func New(cfg config.EDRConfig, c *cache.Cache) *Feed {
	f := &Feed{cfg: cfg, cache: c}
	f.runUpdater = f.defaultRunUpdater
	f.currentSigDB = f.defaultCurrentSigDB
	return f
}

func (f *Feed) defaultRunUpdater() error {
	// Runs the signature updater bundled with the engine. Naming here is
	// internal only; nothing is surfaced in events/logs.
	bin := filepath.Join(config.EngineDir, "freshclam.exe")
	return exec.Run(bin, config.EngineDir, "--config-file", filepath.Join(config.EngineDir, "freshclam.conf"))
}

func (f *Feed) defaultCurrentSigDB() (string, error) {
	// Read the signature-DB version from the engine (engine-name stripped for
	// branding). "" when the daemon is unreachable, so this cycle simply skips
	// cache invalidation.
	return engine.SigDBVersion(f.cfg.ClamdAddr), nil
}

func (f *Feed) updateOnce() (string, error) {
	if err := f.runUpdater(); err != nil {
		logger.Error("UTMStack EDR: signature update failed: %v", err)
		return "", err
	}
	ver, err := f.currentSigDB()
	if err != nil {
		return "", err
	}
	if ver != "" {
		n, err := f.cache.MarkStaleBySigDB(ver)
		if err != nil {
			return ver, err
		}
		logger.Info("UTMStack EDR: signatures updated to %s, %d cached entries invalidated", ver, n)
	}
	return ver, nil
}

func (f *Feed) Run(ctx context.Context) {
	interval := time.Duration(f.cfg.SigUpdateHours) * time.Hour
	if interval <= 0 {
		interval = 4 * time.Hour
	}
	_, _ = f.updateOnce()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = f.updateOnce()
		}
	}
}
