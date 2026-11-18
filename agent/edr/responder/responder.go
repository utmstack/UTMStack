package responder

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"github.com/utmstack/UTMStack/shared/logger"
)

// ErrProcessGone means the target PID no longer exists — expected during a
// tree kill when a child was already terminated via its parent (or by its own
// independent guard). It is a benign race, not a failure.
var ErrProcessGone = errors.New("process already gone")
var ErrUnsupported = errors.New("process response is unavailable on this platform or kernel")
var ErrSuspensionsClosed = errors.New("process suspension is closed for service shutdown")

// errPIDReused means the PID now belongs to a different process than the one
// that was claimed: the claimed process is gone, and the process that took its
// PID must not be touched.
var errPIDReused = fmt.Errorf("%w: its id now belongs to another process", ErrProcessGone)

// claimTTL is how long a termination claim is kept. It covers the race it
// exists for (a child's own guard reaching the kill just after the parent's
// tree kill terminated it) and bounds how long a PID whose process cannot be
// identified stays reserved.
const claimTTL = time.Minute

// maxClaims bounds the claim memory; beyond it the oldest claim is dropped.
const maxClaims = 4096

type Killer interface{ KillPID(pid int) error }

// Identifier is implemented by a Killer that can tell apart two processes that
// held the same PID at different times. On Windows the identity is the process
// creation time (GetProcessTimes).
type Identifier interface {
	// ProcessIdentity returns the identity of the process that holds pid now,
	// or an error wrapping ErrProcessGone when no process holds it.
	ProcessIdentity(pid int) (int64, error)
	// KillIdentity terminates pid only while it is still the process with that
	// identity, and returns an error wrapping ErrProcessGone when it is not.
	KillIdentity(pid int, identity int64) error
}

type Suspender interface {
	SuspendPID(pid int) error
	ResumePID(pid int) error
}

type Responder struct {
	table *proctable.Table
	spool *event.Spool
	kill  Killer
	susp  Suspender

	claims claimSet // processes already claimed for termination (kill-once)
	now    func() time.Time
}

func New(table *proctable.Table, sp *event.Spool, k Killer, s Suspender) *Responder {
	return &Responder{table: table, spool: sp, kill: k, susp: s, now: time.Now}
}

// identitySource says where a processKey's identity came from. Identities
// from different sources are never compared with each other.
type identitySource uint8

const (
	identityUnknown identitySource = iota
	identityOS                     // OS process creation time
	identityTable                  // process-table generation (proctable StartTS)
)

// processKey names one process instance: a PID plus an identity that differs
// between two processes that held that PID over time.
type processKey struct {
	pid int
	id  int64
	src identitySource
}

// claimSet is the kill-once memory. A claim is keyed by process identity, not
// by PID, so a new process that reuses a killed PID can still be killed; it
// expires after claimTTL; and a failed kill releases it so the kill can be
// retried.
type claimSet struct {
	mu sync.Mutex
	at map[processKey]time.Time
}

// take reserves k at now and returns true, or returns false if k is already
// claimed, so each process is terminated once when a child's own guard races
// its parent's tree kill. When the identity is unknown, any live claim on the
// same PID counts: nothing says the process is a different one.
func (c *claimSet) take(k processKey, now time.Time) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.at == nil {
		c.at = make(map[processKey]time.Time)
	}
	var oldest processKey
	var oldestAt time.Time
	for o, at := range c.at {
		if now.Sub(at) >= claimTTL {
			delete(c.at, o)
			continue
		}
		if o == k || (k.src == identityUnknown && o.pid == k.pid) {
			return false
		}
		if oldestAt.IsZero() || at.Before(oldestAt) {
			oldest, oldestAt = o, at
		}
	}
	if len(c.at) >= maxClaims {
		delete(c.at, oldest)
	}
	c.at[k] = now
	return true
}

// release drops the claim on k, so a later request can try the kill again.
func (c *claimSet) release(k processKey) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.at, k)
}

// identify names the process that holds pid now: by its OS creation time when
// the killer can read it, otherwise by its process-table generation. It returns
// an error wrapping ErrProcessGone when the OS reports no process on pid.
func (r *Responder) identify(pid int) (processKey, error) {
	if idf, ok := r.kill.(Identifier); ok {
		id, err := idf.ProcessIdentity(pid)
		if err == nil {
			return processKey{pid: pid, id: id, src: identityOS}, nil
		}
		if errors.Is(err, ErrProcessGone) {
			return processKey{}, err
		}
		logger.Debug(100, "UTMStack EDR: pid %d not identified: %v", pid, err)
	}
	if p, ok := r.table.Get(pid); ok && p.StartTS != 0 {
		return processKey{pid: pid, id: p.StartTS, src: identityTable}, nil
	}
	return processKey{pid: pid}, nil
}

// terminate kills the process named by k. With an OS identity the kill applies
// only to that process, never to one that took its PID since the lookup.
func (r *Responder) terminate(k processKey) error {
	if k.src == identityOS {
		if idf, ok := r.kill.(Identifier); ok {
			return idf.KillIdentity(k.pid, k.id)
		}
	}
	return r.kill.KillPID(k.pid)
}

// treeDescendants returns rootPID and all transitive children, unioning a live OS
// process snapshot (authoritative for what is currently running — catches children
// the async proctable feed has not registered yet) with the proctable (for image/
// cmdline context and any process already gone from the OS). Leaves-first.
func (r *Responder) treeDescendants(rootPID int) []int {
	snapshot := r.table.DescendantProcesses(rootPID)
	table := make([]processParent, 0, len(snapshot))
	for _, p := range snapshot {
		table = append(table, processParent{p.PID, p.PPID})
	}
	live, ok := liveProcessTree(rootPID)
	if !ok {
		live = nil
	}
	return mergeProcessTrees(rootPID, table, live)
}

// KillTree terminates the whole tree rooted at rootPID, leaves-first, emitting
// one branded "killed" event per terminated PID.
func (r *Responder) KillTree(rootPID int, signature string) ([]int, error) {
	pids := r.treeDescendants(rootPID)
	var killed []int
	for _, pid := range pids {
		var p proctable.Proc
		key, err := r.identify(pid)
		if err == nil {
			if !r.claims.take(key, r.now()) {
				// Already being terminated by another path — skip (kill-once).
				continue
			}
			// Read the process context before the kill: once it is gone, a
			// racing path that finds it gone drops its table entry.
			p, _ = r.table.Get(pid)
			if err = r.terminate(key); err != nil && !errors.Is(err, ErrProcessGone) {
				// A failed kill must not block a later attempt.
				r.claims.release(key)
			}
		}
		if err != nil {
			if errors.Is(err, ErrProcessGone) {
				// Already terminated (e.g. by the parent's tree kill or its own
				// guard) — benign race, not a failure. When another process has
				// taken the PID, its table entry is live and stays.
				logger.Debug(100, "UTMStack EDR: pid %d already gone", pid)
				if !errors.Is(err, errPIDReused) {
					r.table.Remove(pid)
				}
			} else {
				logger.Error("UTMStack EDR: failed to terminate pid %d: %v", pid, err)
			}
			continue
		}
		killed = append(killed, pid)
		ev := event.NewProcessAction(event.ActionKilled, event.SourceProcessWatcher,
			event.ProcInfo{PID: pid, PPID: p.PPID, Image: p.Image, Cmdline: p.Cmdline},
			"malicious", signature)
		if js, err := ev.ToJSON(); err == nil {
			_ = r.spool.Append(js)
		}
		r.table.Remove(pid)
	}
	return killed, nil
}

func (r *Responder) Suspend(pid int) error {
	if r.susp == nil {
		return nil
	}
	return r.susp.SuspendPID(pid)
}

func (r *Responder) Resume(pid int) error {
	if r.susp == nil {
		return nil
	}
	return r.susp.ResumePID(pid)
}
