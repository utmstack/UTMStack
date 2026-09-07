//go:build linux
// +build linux

package auditd

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	libaudit "github.com/elastic/go-libaudit/v2"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/utils"
)

type AuditdCollector struct {
	cursor      *auditLogCursor
	reassembler *libaudit.Reassembler
	tailer      *auditLogTailer
	cancel      context.CancelFunc
	mu          sync.Mutex
}

func New() *AuditdCollector {
	return &AuditdCollector{}
}

func (a *AuditdCollector) Name() string {
	return "auditd"
}

func (a *AuditdCollector) Start(ctx context.Context, enqueue func(*plugins.Log) error) {
	if err := checkAuditCapability(); err != nil {
		if !errors.Is(err, ErrAuditUnavailable) {
			utils.Logger.ErrorF("auditd: preflight check failed: %v", err)
		}
		return
	}

	host, err := os.Hostname()
	if err != nil {
		utils.Logger.ErrorF("auditd: error getting hostname: %v", err)
		host = "unknown"
	}

	a.cursor = newAuditLogCursor()
	go a.cursor.flushLoop(ctx)

	restartDelay := auditdRestartDelay

	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info("auditd collector stopping due to context cancellation")
			return
		default:
		}

		exitCode := a.runFileTailer(ctx, host, enqueue)

		if exitCode == 0 {
			utils.Logger.Info("auditd file tailer exited normally")
		} else {
			utils.Logger.ErrorF("auditd file tailer exited with code %d, restarting in %v", exitCode, restartDelay)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(restartDelay):
		}

		// Exponential backoff
		restartDelay *= 2
		if restartDelay > auditdMaxRestartDelay {
			restartDelay = auditdMaxRestartDelay
		}
	}
}

func (a *AuditdCollector) runFileTailer(ctx context.Context, host string, enqueue func(*plugins.Log) error) int {
	a.mu.Lock()
	clientCtx, cancel := context.WithCancel(ctx)
	a.cancel = cancel

	stream := newEventStream(enqueue, host, a.cursor)

	reassembler, err := libaudit.NewReassembler(reassemblerMaxInFlight, reassemblerTimeout, stream)
	if err != nil {
		a.mu.Unlock()
		utils.Logger.ErrorF("auditd: error creating reassembler: %v", err)
		return -1
	}
	a.reassembler = reassembler

	tailer := newAuditLogTailer(a.cursor, reassembler)
	if err := tailer.open(); err != nil {
		reassembler.Close()
		a.reassembler = nil
		a.mu.Unlock()
		utils.Logger.Info("auditd: cannot open %s (%v), will retry", auditLogPath, err)
		return -1
	}
	a.tailer = tailer
	a.mu.Unlock()

	utils.Logger.Info("auditd collector started (tailing %s)", auditLogPath)

	go a.runMaintenance(clientCtx)

	err = tailer.run(clientCtx)
	a.cleanup()
	if err != nil {
		utils.Logger.ErrorF("auditd: %v", err)
		return -1
	}
	return 0
}

// runMaintenance periodically calls Maintain() to flush stale events
func (a *AuditdCollector) runMaintenance(ctx context.Context) {
	ticker := time.NewTicker(maintainInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.mu.Lock()
			if a.reassembler != nil {
				if err := a.reassembler.Maintain(); err != nil {
					utils.Logger.ErrorF("auditd: error in reassembler maintenance: %v", err)
				}
			}
			a.mu.Unlock()
		}
	}
}

// cleanup closes the tailer and reassembler
func (a *AuditdCollector) cleanup() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.reassembler != nil {
		a.reassembler.Close()
		a.reassembler = nil
	}
	if a.tailer != nil {
		a.tailer.close()
		a.tailer = nil
	}
}

// Stop stops the collector
func (a *AuditdCollector) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancel != nil {
		a.cancel()
	}
}
