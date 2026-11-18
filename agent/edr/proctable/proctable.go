package proctable

import (
	"container/list"
	"fmt"
	"sync"
)

type Proc struct {
	PID        int
	PPID       int
	Image      string
	Cmdline    string
	StartTS    int64
	OSIdentity int64
	// IdentityRequired prevents unverified native observations using PID fallback.
	IdentityRequired bool
}

const MaxProcesses = 32768
const MaxRetainedBytes = 16 << 20

type Table struct {
	mu        sync.RWMutex
	procs     map[int]Proc
	order     list.List
	nodes     map[int]*list.Element
	retained  int
	evictions uint64
}

func New() *Table { return &Table{procs: make(map[int]Proc), nodes: make(map[int]*list.Element)} }

func processBytes(p Proc) int { return len(p.Image) + len(p.Cmdline) + 256 }

// Add retains the most recently observed records within fixed byte/count
// budgets. Eviction is explicit coverage loss, reported by Health and Evictions.
func (t *Table) Add(p Proc) {
	t.mu.Lock()
	defer t.mu.Unlock()
	cost := processBytes(p)
	if cost > MaxRetainedBytes {
		t.evictions++
		return
	}
	t.removeLocked(p.PID)
	for len(t.procs) >= MaxProcesses || t.retained+cost > MaxRetainedBytes {
		oldest := t.order.Front()
		if oldest == nil {
			break
		}
		t.removeLocked(oldest.Value.(int))
		t.evictions++
	}
	t.procs[p.PID] = p
	t.nodes[p.PID] = t.order.PushBack(p.PID)
	t.retained += cost
}
func (t *Table) removeLocked(pid int) {
	if p, ok := t.procs[pid]; ok {
		t.retained -= processBytes(p)
		delete(t.procs, pid)
	}
	if node := t.nodes[pid]; node != nil {
		t.order.Remove(node)
		delete(t.nodes, pid)
	}
}
func (t *Table) Remove(pid int) { t.mu.Lock(); defer t.mu.Unlock(); t.removeLocked(pid) }

// RemoveObserved cannot delete a new occupant registered after an old process
// exited or was killed. OS identities refer to captured creation, not event time.
func (t *Table) RemoveObserved(pid int, identity int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if p, ok := t.procs[pid]; ok && identity > 0 && p.OSIdentity == identity {
		t.removeLocked(pid)
	}
}
func (t *Table) Evictions() uint64 { t.mu.RLock(); defer t.mu.RUnlock(); return t.evictions }
func (t *Table) Health() error {
	if n := t.Evictions(); n > 0 {
		return fmt.Errorf("process table capacity lost %d historical observations", n)
	}
	return nil
}

func (t *Table) Get(pid int) (Proc, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	p, ok := t.procs[pid]
	return p, ok
}

// Descendants returns pid and all transitive children, ordered leaves-first
// (deepest descendants first, pid itself last) so callers terminate safely.
func (t *Table) Descendants(pid int) []int {
	processes := t.DescendantProcesses(pid)
	order := make([]int, 0, len(processes))
	for _, p := range processes {
		order = append(order, p.PID)
	}
	return order
}

// DescendantProcesses captures ancestry and process context under one lock.
// The requested root remains present even when no table record exists for it.
func (t *Table) DescendantProcesses(pid int) []Proc {
	t.mu.RLock()
	defer t.mu.RUnlock()

	children := map[int][]int{}
	for _, p := range t.procs {
		children[p.PPID] = append(children[p.PPID], p.PID)
	}
	var order []Proc
	seen := map[int]bool{}
	var visit func(int)
	visit = func(id int) {
		if seen[id] {
			return
		}
		seen[id] = true
		for _, c := range children[id] {
			visit(c)
		}
		p := t.procs[id]
		p.PID = id
		order = append(order, p) // post-order => leaves first
	}
	visit(pid)
	return order
}

// FindProcessesByImage returns one atomic snapshot, including the observed OS
// identity. Looking up PIDs then reading them separately could select reused PIDs.
func (t *Table) FindProcessesByImage(imagePath string) []Proc {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var out []Proc
	for _, p := range t.procs {
		if sameImage(p.Image, imagePath) {
			out = append(out, p)
		}
	}
	return out
}

func (t *Table) FindByImage(imagePath string) []int {
	matches := t.FindProcessesByImage(imagePath)
	out := make([]int, 0, len(matches))
	for _, p := range matches {
		out = append(out, p.PID)
	}
	return out
}
