package agent

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/metadata"
)

// fakeAgentStream is the smallest thing that satisfies
// AgentService_AgentStreamServer for identity-comparison tests.
type fakeAgentStream struct{ id int }

func (fakeAgentStream) Send(*BidirectionalStream) error     { return nil }
func (fakeAgentStream) Recv() (*BidirectionalStream, error) { return nil, nil }
func (fakeAgentStream) SetHeader(metadata.MD) error         { return nil }
func (fakeAgentStream) SendHeader(metadata.MD) error        { return nil }
func (fakeAgentStream) SetTrailer(metadata.MD)              {}
func (fakeAgentStream) Context() context.Context            { return context.Background() }
func (fakeAgentStream) SendMsg(any) error                   { return nil }
func (fakeAgentStream) RecvMsg(any) error                   { return nil }

// TestEvictIfOwner_LeavesForeignStream: an old goroutine returning long after
// a fresh reconnect must NOT clobber the fresh entry.
// TestEvictIfOwner_RemovesOwnedStream: the current owner cleans up on exit.
func TestEvictIfOwner(t *testing.T) {
	s := &AgentService{AgentStreamMap: map[uint]AgentService_AgentStreamServer{}}
	old := &fakeAgentStream{id: 1}
	fresh := &fakeAgentStream{id: 2}

	s.AgentStreamMap[42] = fresh
	s.evictIfOwner(42, old)
	if _, ok := s.AgentStreamMap[42]; !ok {
		t.Fatal("evictIfOwner clobbered a fresh stream owned by a different goroutine")
	}
	if s.AgentStreamMap[42] != fresh {
		t.Fatal("evictIfOwner replaced the fresh entry with something else")
	}

	s.evictIfOwner(42, fresh)
	if _, ok := s.AgentStreamMap[42]; ok {
		t.Fatal("evictIfOwner did not remove the owned entry")
	}
}

// runWithTimeout reports whether fn returns within d. A blocking delivery
// (the pre-fix deadlock) fails the test instead of hanging it for minutes.
func runWithTimeout(t *testing.T, d time.Duration, fn func()) bool {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()
	select {
	case <-done:
		return true
	case <-time.After(d):
		t.Fatalf("operation blocked for %s — the AgentStream result delivery path deadlocks", d)
		return false
	}
}

// TestTryDeliverResult_WaitingPanel: the normal path — a panel is blocked on
// the slot's channel and receives the result. The reader reports what it got
// over an unbuffered channel, so the delivery is what unblocks it.
func TestTryDeliverResult_WaitingPanel(t *testing.T) {
	s := &AgentService{
		AgentStreamMap:       map[uint]AgentService_AgentStreamServer{},
		CommandResultChannel: map[string]chan *CommandResult{},
	}
	s.CommandResultChannel["cmd-1"] = make(chan *CommandResult, 1)

	received := make(chan *CommandResult) // unbuffered: sent only by the reader
	go func() {
		received <- <-s.CommandResultChannel["cmd-1"]
	}()
	// Let the reader block in its receive before delivering.
	select {
	case <-received:
		t.Fatal("reader returned without any delivery")
	case <-time.After(50 * time.Millisecond):
	}

	if !s.tryDeliverResult(&CommandResult{AgentId: "7", CmdId: "cmd-1", Result: "ok"}) {
		t.Fatal("tryDeliverResult returned false for a waiting panel")
	}

	select {
	case got := <-received:
		if got == nil || got.Result != "ok" {
			t.Fatalf("panel received %v, want the delivered result", got)
		}
	case <-time.After(time.Second):
		t.Fatal("panel did not receive the delivered result")
	}
}

// TestTryDeliverResult_NoSlot: an unknown cmd_id must not block and must
// report undelivered.
func TestTryDeliverResult_NoSlot(t *testing.T) {
	s := &AgentService{CommandResultChannel: map[string]chan *CommandResult{}}
	if !runWithTimeout(t, time.Second, func() {
		if s.tryDeliverResult(&CommandResult{CmdId: "missing", Result: "x"}) {
			t.Error("tryDeliverResult returned true with no slot registered")
		}
	}) {
		t.Fatal("delivery with no slot blocked")
	}
}

// TestTryDeliverResult_PanelGone is the regression test for the freeze: the
// panel disconnected without claiming its result. The first result is
// absorbed by the buffer; a second delivery must drop the stale slot and
// return immediately. Before the fix, any delivery into a slot with no
// reader blocked the AgentStream goroutine forever, holding the global
// CommandResultChannelM and the whole command system with it.
func TestTryDeliverResult_PanelGone(t *testing.T) {
	s := &AgentService{CommandResultChannel: map[string]chan *CommandResult{}}
	s.CommandResultChannel["cmd-2"] = make(chan *CommandResult, 1)

	// First result: absorbed into the buffer, nobody reads it.
	if !runWithTimeout(t, time.Second, func() {
		if !s.tryDeliverResult(&CommandResult{CmdId: "cmd-2", Result: "first"}) {
			t.Error("first result should be absorbed by the buffer")
		}
	}) {
		t.Fatal("first delivery blocked")
	}

	// Second result: must not block, and must reclaim the orphaned slot.
	if !runWithTimeout(t, 2*time.Second, func() {
		if s.tryDeliverResult(&CommandResult{CmdId: "cmd-2", Result: "second"}) {
			t.Error("second result delivered into a full stale slot")
		}
	}) {
		t.Fatal("second delivery blocked")
	}

	s.CommandResultChannelM.Lock()
	_, still := s.CommandResultChannel["cmd-2"]
	s.CommandResultChannelM.Unlock()
	if still {
		t.Fatal("orphaned slot was not reclaimed after a dropped delivery")
	}
}

// TestReclaimResultSlot: every handlePanelCommand exit path runs this via
// defer; it must remove the slot, and be safe to call twice.
func TestReclaimResultSlot(t *testing.T) {
	s := &AgentService{CommandResultChannel: map[string]chan *CommandResult{}}
	s.CommandResultChannel["cmd-3"] = make(chan *CommandResult, 1)

	s.reclaimResultSlot("cmd-3")

	s.CommandResultChannelM.Lock()
	_, still := s.CommandResultChannel["cmd-3"]
	s.CommandResultChannelM.Unlock()
	if still {
		t.Fatal("reclaimResultSlot did not remove the slot")
	}

	// Reclaiming an unknown id is a harmless no-op (double-reclaim safety).
	if !runWithTimeout(t, time.Second, func() { s.reclaimResultSlot("cmd-3") }) {
		t.Fatal("reclaiming a missing slot blocked")
	}
}
