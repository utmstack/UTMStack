package recovery

import (
	"context"
	"fmt"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"gorm.io/gorm"
)

// configRecoveryDir and configGlobalConcurrency are package-level variables
// read from config at Init time. They exist so that service.go and dispatcher.go
// can access config values without importing config directly (avoiding potential
// future import-cycle issues as the package grows).
var (
	configRecoveryDir       string
	configGlobalConcurrency int
	configDispatchEnabled   bool
)

// streamProvider is the package-level StreamProvider, injected via
// SetStreamProvider. Stored separately from dispatcherState so it can be
// set before Init is called (e.g. by agent package init()).
var streamProvider StreamProvider

// SetStreamProvider is called by agent/agent_imp.go init() to inject the
// real provider. Stored package-level so dispatcher uses it lazily.
// Must be called before Init (or at least before the first dispatch fires).
func SetStreamProvider(p StreamProvider) {
	streamProvider = p
	// If the dispatcher is already running, update its provider too.
	disp.mu.Lock()
	disp.provider = p
	disp.mu.Unlock()
}

// Init is the package entrypoint. main.go calls it after MigrateDatabase.
//
// Flow:
//  1. Run BootReconcile (always, regardless of dispatch enabled, so the
//     DB is reconciled even when dispatch is off)
//  2. If config.RecoveryDispatchEnabled: dispatcher.Start(ctx, db, provider)
//     Else: log info that dispatch is disabled and return
//
// The provider is taken from the package-level streamProvider set via
// SetStreamProvider. It may be nil in Phase 2B — Phase 3 wires the real provider.
func Init(ctx context.Context, db *gorm.DB) error {
	// Cache config values into package-level vars so subpackage files
	// (service.go, dispatcher.go) can read them without importing config.
	configRecoveryDir = config.RecoveryDir
	configGlobalConcurrency = config.RecoveryGlobalConcurrency
	configDispatchEnabled = config.RecoveryDispatchEnabled

	// Step 1: Boot reconciliation — always runs.
	if err := BootReconcile(ctx, db); err != nil {
		return fmt.Errorf("recovery init: boot reconcile: %w", err)
	}

	// Step 2: Start dispatcher if enabled.
	if !configDispatchEnabled {
		catcher.Info("Recovery dispatch is disabled (RECOVERY_DISPATCH_ENABLED=false)", map[string]any{
			"event":   "recovery_dispatch_disabled",
			"process": logProcess,
		})
		return nil
	}

	Start(ctx, db, streamProvider)
	return nil
}

// Shutdown is the external entrypoint for graceful shutdown (main.go).
// It wraps doShutdown from shutdown.go.
func Shutdown(timeout time.Duration) error {
	return doShutdown(timeout)
}
