// Package auditd provides a native collector for Linux Audit Framework events.
// It uses go-libaudit to receive events via netlink multicast and reassembles
// them before sending to the log queue.
package auditd

import "time"

const (
	// auditdRestartDelay is the initial delay between reconnection attempts
	auditdRestartDelay = 5 * time.Second

	// auditdMaxRestartDelay is the maximum backoff delay for reconnection
	auditdMaxRestartDelay = 5 * time.Minute

	// reassemblerMaxInFlight is the maximum number of events held for reassembly
	reassemblerMaxInFlight = 50

	// reassemblerTimeout is how long to wait for related messages before flushing
	reassemblerTimeout = 2 * time.Second

	// maintainInterval is how often to run Reassembler.Maintain() to flush stale events
	maintainInterval = 500 * time.Millisecond
)
