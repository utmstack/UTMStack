package feed

import (
	"bytes"
	"context"
	"fmt"
	osexec "os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/shared/logger"
)

// mirrorFailThreshold is the number of consecutive private-mirror update
// failures tolerated before one cycle falls back to the public database.
const mirrorFailThreshold = 3

type Feed struct {
	cfg          config.EDRConfig
	cache        *cache.Cache
	runUpdater   func() error
	currentSigDB func() (string, error)
	writeConf    func(useCDN bool) error

	mirrorFails   int  // consecutive private-mirror update failures
	onCDNFallback bool // this cycle's config points at the public database

	mu            sync.Mutex   // guards lastSuccess (cross-goroutine status reads)
	lastSuccess   time.Time    // time of the last successful signature update
	spool         *event.Spool // branded health-event sink (nil until SetSpool)
	staleNotified bool         // one health event per staleness crossing
}

func New(cfg config.EDRConfig, c *cache.Cache) *Feed {
	f := &Feed{cfg: cfg, cache: c}
	f.runUpdater = f.defaultRunUpdater
	f.currentSigDB = f.defaultCurrentSigDB
	f.writeConf = func(useCDN bool) error { return engine.WriteFreshclamConf(cfg, useCDN) }
	return f
}

// engineNameRedactor strips engine/vendor tokens from updater output so they
// never reach a surfaced error/log line (branding rule). The updater's stderr
// can name the engine or the public database host (e.g. on a CDN-fallback
// cycle), which the static labeling audit cannot catch in a runtime value.
var engineNameRedactor = strings.NewReplacer(
	"database.clamav.net", "the signature database",
	"ClamAV", "the signature engine", "clamav", "the signature engine",
	"clamd", "the signature engine", "freshclam", "the signature updater",
)

func (f *Feed) defaultRunUpdater() error {
	// Runs the signature updater bundled with the engine. Internal identifiers
	// here are fine; any updater output that reaches a surfaced error/log line
	// is redacted (engineNameRedactor) so the engine name never leaks.
	bin := filepath.Join(config.EngineDir, "freshclam.exe")
	cmd := osexec.Command(bin, "--config-file", filepath.Join(config.EngineDir, "freshclam.conf"))
	cmd.Dir = config.EngineDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("updater failed: %s: %w", engineNameRedactor.Replace(stderr.String()), err)
		}
		return fmt.Errorf("updater failed: %w", err)
	}
	return nil
}

func (f *Feed) defaultCurrentSigDB() (string, error) {
	// Read the signature-DB version from the engine (engine-name stripped for
	// branding). "" when the daemon is unreachable, so this cycle simply skips
	// cache invalidation.
	return engine.SigDBVersion(f.cfg.ClamdAddr), nil
}

func (f *Feed) updateOnce() (string, error) {
	err := f.runUpdater()
	if err != nil {
		logger.Error("UTMStack EDR: signature update failed: %v", err)
		if engine.ResolveMirrorURL(f.cfg) != "" && f.cfg.SignatureFallback == "cdn" {
			f.mirrorFails++
			if f.mirrorFails >= mirrorFailThreshold && !f.onCDNFallback {
				logger.Info("UTMStack EDR: signature mirror unreachable; falling back to public updates for one cycle")
				if werr := f.writeConf(true); werr == nil {
					f.onCDNFallback = true
					err = f.runUpdater()
				}
			}
		}
		if err != nil {
			return "", err
		}
	}
	if f.onCDNFallback {
		// Restore the mirror config so the next cycle retries the mirror.
		_ = f.writeConf(false)
		f.onCDNFallback = false
	}
	f.mirrorFails = 0
	f.setLastSuccess(time.Now())
	f.staleNotified = false
	ver, verr := f.currentSigDB()
	if verr != nil {
		return "", verr
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

// nextDelay picks the wait before the next update attempt: the configured
// interval on success, a short retry after a failure (so mirror failover is
// reachable in minutes, not multiples of sig_update_hours).
func nextDelay(err error, interval time.Duration) time.Duration {
	const retry = 15 * time.Minute
	if err != nil && retry < interval {
		return retry
	}
	return interval
}

func (f *Feed) Run(ctx context.Context) {
	interval := time.Duration(f.cfg.SigUpdateHours) * time.Hour
	if interval <= 0 {
		interval = 4 * time.Hour
	}
	_, err := f.updateOnce()
	f.noteHealth(interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(nextDelay(err, interval)):
			_, err = f.updateOnce()
			f.noteHealth(interval)
		}
	}
}
