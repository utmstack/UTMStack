//go:build linux
// +build linux

package auditd

import (
	"github.com/elastic/go-libaudit/v2/auparse"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
)

// eventStream implements libaudit.Stream interface for reassembled events
type eventStream struct {
	queue    chan *plugins.Log
	hostname string
}

// newEventStream creates a new eventStream
func newEventStream(queue chan *plugins.Log, hostname string) *eventStream {
	return &eventStream{
		queue:    queue,
		hostname: hostname,
	}
}

// ReassemblyComplete is called when a complete group of events has been received.
// Uses non-blocking send to prevent backpressure from propagating to the kernel.
// If the queue is full, events are dropped rather than blocking.
func (s *eventStream) ReassemblyComplete(msgs []*auparse.AuditMessage) {
	if len(msgs) == 0 {
		return
	}

	jsonOutput, err := formatAuditEvent(msgs)
	if err != nil {
		utils.Logger.ErrorF("auditd: error formatting event: %v", err)
		return
	}

	log := &plugins.Log{
		DataType:   string(config.DataTypeLinuxAgent),
		DataSource: s.hostname,
		Raw:        jsonOutput,
	}

	// Non-blocking send: drop events if queue is full to prevent backpressure
	// This is the "user-space" backpressure mitigation strategy from Elastic Auditbeat
	select {
	case s.queue <- log:
		// Event sent successfully
	default:
		// Queue is full - drop event to prevent backpressure to kernel
		utils.Logger.ErrorF("auditd: queue full, dropping event (sequence=%d)", msgs[0].Sequence)
	}
}

// EventsLost is called when events were lost due to buffer overflow or rate limiting
func (s *eventStream) EventsLost(count int) {
	// Ignore invalid counts - large values indicate sequence number rollover/overflow
	// not actual lost events. A reasonable max is 100K events lost in one batch.
	if count <= 0 || count > 100000 {
		return
	}
	utils.Logger.ErrorF("auditd: %d events lost due to buffer overflow or rate limiting", count)
}
