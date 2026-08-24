// Package configwatcher provides a shared config file watcher using fsnotify.
package configwatcher

import (
	"context"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
)

// FallbackInterval is used as a safety net in case fsnotify misses events.
const FallbackInterval = 5 * time.Minute

// Watch blocks until ctx is cancelled. collector.StopAll depends on that: the
// return of a collector's Start is its signal that no reconcile is still in
// flight. An implementation that returned early while the loop kept running in
// the background would silently reintroduce the shutdown race.
func Watch(ctx context.Context, name string, onConfigChange func()) {
	// Check without blocking whether we were cancelled before doing any work:
	// the initial onConfigChange below opens sockets, and a collector whose
	// goroutine had not been scheduled yet when StopAll ran must not open them.
	select {
	case <-ctx.Done():
		utils.Logger.Info("%s: stopping config watcher", name)
		return
	default:
	}

	// Initial call
	onConfigChange()

	// Set up fsnotify watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		utils.Logger.ErrorF("%s: failed to create fsnotify watcher: %v, falling back to polling", name, err)
		runPollingFallback(ctx, name, onConfigChange)
		return
	}
	defer watcher.Close()

	// Watch the directory containing the config file
	configDir := filepath.Dir(config.CollectorFileName)
	configBase := filepath.Base(config.CollectorFileName)

	if err := watcher.Add(configDir); err != nil {
		utils.Logger.ErrorF("%s: failed to watch config directory: %v, falling back to polling", name, err)
		runPollingFallback(ctx, name, onConfigChange)
		return
	}

	utils.Logger.Info("%s: watching config file for changes", name)

	// Fallback timer as safety net
	fallbackTicker := time.NewTicker(FallbackInterval)
	defer fallbackTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info("%s: stopping config watcher", name)
			return

		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if filepath.Base(event.Name) == configBase {
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					utils.Logger.Info("%s: config file changed, reconciling", name)
					onConfigChange()
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			utils.Logger.ErrorF("%s: fsnotify error: %v", name, err)

		case <-fallbackTicker.C:
			onConfigChange()
		}
	}
}

// runPollingFallback is used when fsnotify cannot be initialized.
func runPollingFallback(ctx context.Context, name string, onConfigChange func()) {
	utils.Logger.Info("%s: using polling fallback", name)
	ticker := time.NewTicker(FallbackInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info("%s: stopping config watcher", name)
			return
		case <-ticker.C:
			onConfigChange()
		}
	}
}
