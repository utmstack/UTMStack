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
	"github.com/elastic/go-libaudit/v2/auparse"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/utils"
)

// AuditdCollector collects Linux Audit events via netlink multicast
type AuditdCollector struct {
	client      auditReceiver
	reassembler *libaudit.Reassembler
	cancel      context.CancelFunc
	mu          sync.Mutex
}

// New creates a new AuditdCollector
func New() *AuditdCollector {
	return &AuditdCollector{}
}

// Name returns the collector name
func (a *AuditdCollector) Name() string {
	return "auditd"
}

// Start begins collecting audit events and sending them to the queue
func (a *AuditdCollector) Start(ctx context.Context, queue chan *plugins.Log) {
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

	restartDelay := auditdRestartDelay

	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info("auditd collector stopping due to context cancellation")
			return
		default:
		}

		exitCode := a.runAuditClient(ctx, host, queue)

		switch exitCode {
		case 0:
			utils.Logger.Info("auditd client exited normally")
		case auditdExitPermanent:
			// Environment cannot run auditd (e.g. missing CAP_AUDIT_*, kernel
			// audit disabled). Retrying will never succeed — exit silently.
			utils.Logger.Info("auditd collector disabled: audit subsystem not accessible in this environment")
			return
		default:
			utils.Logger.ErrorF("auditd client exited with code %d, restarting in %v", exitCode, restartDelay)
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

// auditdExitPermanent signals that the collector cannot run in this
// environment and must not be retried.
const auditdExitPermanent = -2

// runAuditClient creates the audit client and runs the receive loop
func (a *AuditdCollector) runAuditClient(ctx context.Context, host string, queue chan *plugins.Log) int {
	a.mu.Lock()
	clientCtx, cancel := context.WithCancel(ctx)
	a.cancel = cancel

	// Attempt to set kernel backlog limit and wait time. Both are best-effort
	// tuning that require CAP_AUDIT_CONTROL — if they fail we just continue
	// with kernel defaults, so log at Info level to avoid noise on restricted
	// hosts (e.g. containers without audit capabilities).
	if err := setKernelBacklogLimit(kernelBacklogLimit); err != nil {
		utils.Logger.Info("auditd: could not set kernel backlog limit (%v), using default", err)
	} else {
		utils.Logger.Info("auditd: kernel backlog limit set to %d", kernelBacklogLimit)
	}

	if err := setBacklogWaitTime(0); err != nil {
		utils.Logger.Info("auditd: could not set backlog wait time (%v), using default", err)
	} else {
		utils.Logger.Info("auditd: backlog wait time set to 0 (non-blocking mode)")
	}

	// Create multicast audit client. Failure here (typically EPERM when the
	// agent lacks CAP_AUDIT_READ) is permanent for the current process — the
	// outer Start loop treats auditdExitPermanent as a no-retry condition.
	client, err := newAuditClient()
	if err != nil {
		a.mu.Unlock()
		utils.Logger.Info("auditd: cannot open audit netlink socket (%v); collector will not run", err)
		return auditdExitPermanent
	}
	a.client = client

	// Create event stream for reassembled events
	stream := newEventStream(queue, host)

	// Create reassembler
	reassembler, err := libaudit.NewReassembler(reassemblerMaxInFlight, reassemblerTimeout, stream)
	if err != nil {
		client.Close()
		a.mu.Unlock()
		utils.Logger.ErrorF("auditd: error creating reassembler: %v", err)
		return -1
	}
	a.reassembler = reassembler
	a.mu.Unlock()

	utils.Logger.Info("auditd collector started (netlink multicast)")

	// Start maintenance goroutine for reassembler
	go a.runMaintenance(clientCtx)

	// Main receive loop
	for {
		select {
		case <-clientCtx.Done():
			a.cleanup()
			return 0
		default:
		}

		// Receive with non-blocking to allow checking context
		msg, err := client.Receive(false)
		if err != nil {
			utils.Logger.ErrorF("auditd: error receiving message: %v", err)
			a.cleanup()
			return -1
		}

		if msg == nil {
			// No message available, brief sleep to avoid busy loop
			time.Sleep(10 * time.Millisecond)
			continue
		}

		// Parse message type from raw data
		msgType := auparse.AuditMessageType(msg.Type)

		// Push to reassembler for event grouping
		if err := reassembler.Push(msgType, msg.Data); err != nil {
			utils.Logger.ErrorF("auditd: error pushing to reassembler: %v", err)
		}
	}
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

// cleanup closes the client and reassembler
func (a *AuditdCollector) cleanup() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.reassembler != nil {
		a.reassembler.Close()
		a.reassembler = nil
	}
	if a.client != nil {
		a.client.Close()
		a.client = nil
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
