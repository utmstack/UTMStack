package recovery

import "sync"

// StreamMutex provides per-agent serialization for AgentStream.Send calls.
//
// The agent-manager's AgentStreamMap is shared between panel-driven
// ProcessCommand and recovery dispatch. Concurrent Send calls on the same
// gRPC stream cause undefined behavior — the gRPC server does NOT serialize
// concurrent writes on a single ServerStream. This mutex map serializes
// per-agent.
//
// SECURITY: ALL Send calls to an agent stream MUST acquire StreamMutex.For(id)
// before invoking Send and release it after.
//
// Usage:
//
//	mu := recovery.StreamMutex.For(agentID)
//	mu.Lock()
//	defer mu.Unlock()
//	stream.Send(...)
var StreamMutex = &streamMutexMap{}

// streamMutexMap is a concurrency-safe map of per-agent mutexes.
type streamMutexMap struct {
	m sync.Map // map[uint]*sync.Mutex
}

// For returns the *sync.Mutex for the given agent ID, lazily creating it on
// first access. Safe for concurrent use — uses sync.Map.LoadOrStore to avoid
// races on first creation.
//
// Entries are never removed. A disconnected agent's leaked mutex is ~8 bytes,
// negligible in practice (cleanup is a future improvement if ever needed).
func (s *streamMutexMap) For(agentID uint) *sync.Mutex {
	mu := &sync.Mutex{}
	actual, _ := s.m.LoadOrStore(agentID, mu)
	return actual.(*sync.Mutex)
}
