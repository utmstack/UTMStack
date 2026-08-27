package agent

import (
	"context"
	"testing"

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
