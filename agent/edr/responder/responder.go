package responder

import (
	"errors"
	"sync"

	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"github.com/utmstack/UTMStack/shared/logger"
)

// ErrProcessGone means the target PID no longer exists — expected during a
// tree kill when a child was already terminated via its parent (or by its own
// independent guard). It is a benign race, not a failure.
var ErrProcessGone = errors.New("process already gone")

type Killer interface{ KillPID(pid int) error }

type Suspender interface {
	SuspendPID(pid int) error
	ResumePID(pid int) error
}

type Responder struct {
	table *proctable.Table
	spool *event.Spool
	kill  Killer
	susp  Suspender

	mu      sync.Mutex
	claimed map[int]bool // PIDs already claimed for termination (kill-once)
}

func New(table *proctable.Table, sp *event.Spool, k Killer, s Suspender) *Responder {
	return &Responder{table: table, spool: sp, kill: k, susp: s, claimed: make(map[int]bool)}
}

// claim atomically reserves a PID for termination; returns false if another
// tree kill (e.g. a child's own guard racing the parent's) already claimed it,
// so each PID is terminated exactly once.
func (r *Responder) claim(pid int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.claimed[pid] {
		return false
	}
	r.claimed[pid] = true
	return true
}

// treeDescendants returns rootPID and all transitive children, unioning a live OS
// process snapshot (authoritative for what is currently running — catches children
// the async proctable feed has not registered yet) with the proctable (for image/
// cmdline context and any process already gone from the OS). Leaves-first.
func (r *Responder) treeDescendants(rootPID int) []int {
	fromTable := r.table.Descendants(rootPID)
	live, ok := liveDescendants(rootPID)
	if !ok {
		return fromTable
	}
	seen := map[int]bool{}
	out := make([]int, 0, len(live)+len(fromTable))
	for _, pid := range live { // leaves-first from the live tree
		if !seen[pid] {
			seen[pid] = true
			out = append(out, pid)
		}
	}
	for _, pid := range fromTable { // any proctable-only pids (no longer in the OS list)
		if !seen[pid] {
			seen[pid] = true
			out = append(out, pid)
		}
	}
	return out
}

// KillTree terminates the whole tree rooted at rootPID, leaves-first, emitting
// one branded "killed" event per terminated PID.
func (r *Responder) KillTree(rootPID int, signature string) ([]int, error) {
	pids := r.treeDescendants(rootPID)
	var killed []int
	for _, pid := range pids {
		if !r.claim(pid) {
			// Already being terminated by another path — skip (kill-once).
			continue
		}
		if err := r.kill.KillPID(pid); err != nil {
			if errors.Is(err, ErrProcessGone) {
				// Already terminated (e.g. by the parent's tree kill or its own
				// guard) — benign race, not a failure.
				logger.Debug(100, "UTMStack EDR: pid %d already gone", pid)
				r.table.Remove(pid)
			} else {
				logger.Error("UTMStack EDR: failed to terminate pid %d: %v", pid, err)
			}
			continue
		}
		killed = append(killed, pid)
		p, _ := r.table.Get(pid)
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
