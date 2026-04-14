//go:build linux
// +build linux

package auditd

import (
	libaudit "github.com/elastic/go-libaudit/v2"
)

const (
	// kernelBacklogLimit is the audit backlog limit to set in the kernel.
	// Higher values prevent event loss under high load.
	kernelBacklogLimit uint32 = 8192
)

// auditReceiver interface wraps the go-libaudit client for testability
type auditReceiver interface {
	Receive(nonBlocking bool) (*libaudit.RawAuditMessage, error)
	Close() error
}

// auditClientWrapper wraps the go-libaudit AuditClient to implement auditReceiver
type auditClientWrapper struct {
	client *libaudit.AuditClient
}

// Receive receives a raw audit message from the netlink socket
func (w *auditClientWrapper) Receive(nonBlocking bool) (*libaudit.RawAuditMessage, error) {
	return w.client.Receive(nonBlocking)
}

// Close closes the underlying audit client
func (w *auditClientWrapper) Close() error {
	return w.client.Close()
}

// newAuditClient creates a new multicast audit client wrapped in the auditReceiver interface
func newAuditClient() (auditReceiver, error) {
	client, err := libaudit.NewMulticastAuditClient(nil)
	if err != nil {
		return nil, err
	}
	return &auditClientWrapper{client: client}, nil
}

// setKernelBacklogLimit sets the kernel audit backlog limit using a separate
// control client. This requires root/CAP_AUDIT_CONTROL privileges.
// The multicast client cannot set kernel parameters, so we create a temporary
// unicast client specifically for this configuration.
func setKernelBacklogLimit(limit uint32) error {
	controlClient, err := libaudit.NewAuditClient(nil)
	if err != nil {
		return err
	}
	defer controlClient.Close()

	return controlClient.SetBacklogLimit(limit, libaudit.WaitForReply)
}

// setBacklogWaitTime configures the kernel to not block processes when the
// audit backlog queue is full. With waitTime=0, the kernel will drop events
// instead of stalling audited processes. This is the "kernel" backpressure
// mitigation strategy used by Elastic Auditbeat.
// Requires CAP_AUDIT_CONTROL and kernel 3.14+.
func setBacklogWaitTime(waitTime int32) error {
	controlClient, err := libaudit.NewAuditClient(nil)
	if err != nil {
		return err
	}
	defer controlClient.Close()

	return controlClient.SetBacklogWaitTime(waitTime, libaudit.WaitForReply)
}
