//go:build linux
// +build linux

package auditd

import (
	"github.com/elastic/go-libaudit/v2/auparse"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
)

const (
	// eventsLostThreshold - only log when this many events are lost at once.
	// Small losses (1-10) are normal under high load and not worth logging.
	eventsLostThreshold = 50

	// eventsLostMaxReasonable is the maximum "reasonable" number of lost events.
	eventsLostMaxReasonable = 1000000
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
	select {
	case s.queue <- log:
		// Event sent successfully
	default:
		// Queue is full - drop event to prevent backpressure to kernel
		utils.Logger.ErrorF("auditd: queue full, dropping event (sequence=%d)", msgs[0].Sequence)
	}
}

// EventsLost is called when events were lost due to buffer overflow or rate limiting.
// We filter these out by checking against a reasonable maximum.
func (s *eventStream) EventsLost(count int) {
	// Filter out unreasonable values caused by sequence number rollover bug
	if count < eventsLostThreshold || count > eventsLostMaxReasonable {
		return
	}
	utils.Logger.ErrorF("auditd: %d events lost due to buffer overflow or rate limiting", count)
}
