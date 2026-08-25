package agent

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// fakeAgentStream stands in for one agent's connection. Recv blocks until the
// test feeds it something or closes it, which is how a real idle stream behaves
// and what makes the takeover observable.
type fakeAgentStream struct {
	grpc.ServerStream
	incoming chan *BidirectionalStream

	mu   sync.Mutex
	sent []*BidirectionalStream
}

func newFakeAgentStream() *fakeAgentStream {
	return &fakeAgentStream{incoming: make(chan *BidirectionalStream, 4)}
}

func (f *fakeAgentStream) Context() context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		"id": "1", "key": "k", "type": "agent",
	}))
}

func (f *fakeAgentStream) Recv() (*BidirectionalStream, error) {
	msg, ok := <-f.incoming
	if !ok {
		return nil, context.Canceled
	}
	return msg, nil
}

func (f *fakeAgentStream) Send(m *BidirectionalStream) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, m)
	return nil
}

func (f *fakeAgentStream) sentMessages() []*BidirectionalStream {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.sent)
}

func newTestService() *AgentService {
	return &AgentService{
		AgentStreamMap:       make(map[uint]AgentService_AgentStreamServer),
		CommandResultChannel: make(map[string]chan *CommandResult),
	}
}

// waitFor gives a goroutine a bounded chance to reach the state we care about.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func (s *AgentService) streamFor(id uint) AgentService_AgentStreamServer {
	s.AgentStreamMutex.Lock()
	defer s.AgentStreamMutex.Unlock()
	return s.AgentStreamMap[id]
}

// An agent that reconnects used to be turned away with AlreadyExists, because
// the server still held the dead stream it left behind. Commands kept failing
// with "agent not found or is disconnected" until Recv finally gave up on the
// corpse, so the fix for a mute agent has to include letting it back in.
func TestReconnectingAgentTakesOverItsStream(t *testing.T) {
	s := newTestService()

	first := newFakeAgentStream()
	go func() { _ = s.AgentStream(first) }()
	waitFor(t, "the first stream to register", func() bool { return s.streamFor(1) == AgentService_AgentStreamServer(first) })

	second := newFakeAgentStream()
	go func() { _ = s.AgentStream(second) }()
	waitFor(t, "the reconnecting agent to take over", func() bool { return s.streamFor(1) == AgentService_AgentStreamServer(second) })
}

// The old handler eventually notices its stream is dead. If it deleted the map
// entry unconditionally it would evict the connection that replaced it, and the
// agent would be unreachable despite being connected.
func TestDepartingHandlerDoesNotEvictTheStreamThatReplacedIt(t *testing.T) {
	s := newTestService()

	first := newFakeAgentStream()
	firstReturned := make(chan struct{})
	go func() { _ = s.AgentStream(first); close(firstReturned) }()
	waitFor(t, "the first stream to register", func() bool { return s.streamFor(1) == AgentService_AgentStreamServer(first) })

	second := newFakeAgentStream()
	go func() { _ = s.AgentStream(second) }()
	waitFor(t, "the reconnecting agent to take over", func() bool { return s.streamFor(1) == AgentService_AgentStreamServer(second) })

	close(first.incoming)
	select {
	case <-firstReturned:
	case <-time.After(3 * time.Second):
		t.Fatal("the first handler never returned")
	}

	if got := s.streamFor(1); got != AgentService_AgentStreamServer(second) {
		t.Fatalf("the departing handler evicted the live stream: map holds %v, want the second stream", got)
	}
}

// The response half of the stream needs traffic too: a proxy times out an idle
// response as readily as an idle request body, and only the server can keep
// that half warm.
func TestServerAnswersHeartbeats(t *testing.T) {
	s := newTestService()

	stream := newFakeAgentStream()
	go func() { _ = s.AgentStream(stream) }()
	waitFor(t, "the stream to register", func() bool { return s.streamFor(1) == AgentService_AgentStreamServer(stream) })

	stream.incoming <- &BidirectionalStream{
		StreamMessage: &BidirectionalStream_Heartbeat{Heartbeat: &Heartbeat{}},
	}

	waitFor(t, "the heartbeat to be answered", func() bool {
		for _, m := range stream.sentMessages() {
			if _, ok := m.StreamMessage.(*BidirectionalStream_Heartbeat); ok {
				return true
			}
		}
		return false
	})
}
