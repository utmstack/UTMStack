//go:build linux
// +build linux

package auditd

import (
	"context"
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
	// Preflight check for audit capability
	if err := checkAuditCapability(); err != nil {
		utils.Logger.ErrorF("auditd: preflight check failed: %v", err)
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

		if exitCode == 0 {
			utils.Logger.Info("auditd client exited normally")
		} else {
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

// runAuditClient creates the audit client and runs the receive loop
func (a *AuditdCollector) runAuditClient(ctx context.Context, host string, queue chan *plugins.Log) int {
	a.mu.Lock()
	clientCtx, cancel := context.WithCancel(ctx)
	a.cancel = cancel

	// Create multicast audit client
	client, err := newAuditClient()
	if err != nil {
		a.mu.Unlock()
		utils.Logger.ErrorF("auditd: error creating audit client: %v", err)
		return -1
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
