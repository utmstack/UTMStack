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
	enqueue  func(*plugins.Log) error
	hostname string
	cursor   *auditLogCursor
}

// newEventStream creates a new eventStream
func newEventStream(enqueue func(*plugins.Log) error, hostname string, cursor *auditLogCursor) *eventStream {
	return &eventStream{
		enqueue:  enqueue,
		hostname: hostname,
		cursor:   cursor,
	}
}

func (s *eventStream) ReassemblyComplete(msgs []*auparse.AuditMessage) {
	if len(msgs) == 0 {
		return
	}

	seq := msgs[0].Sequence

	jsonOutput, err := formatAuditEvent(msgs)
	if err != nil {
		utils.Logger.ErrorF("auditd: error formatting event: %v", err)
		s.cursor.resolve(seq, false)
		return
	}

	log := &plugins.Log{
		DataType:   string(config.DataTypeLinuxAgent),
		DataSource: s.hostname,
		Raw:        jsonOutput,
	}

	err = s.enqueue(log)
	if err != nil {
		utils.Logger.ErrorF("auditd: failed to persist event (sequence=%d): %v", seq, err)
	}
	// Only advance the audit.log resume position past this event once it's
	// durably persisted — resolve(seq, false) still drops the pending
	// entry so it doesn't leak, it just doesn't move the cursor forward.
	s.cursor.resolve(seq, err == nil)
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
