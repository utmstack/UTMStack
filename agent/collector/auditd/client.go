//go:build linux
// +build linux

package auditd

import (
	libaudit "github.com/elastic/go-libaudit/v2"
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
