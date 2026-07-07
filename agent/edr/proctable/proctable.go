package proctable

import (
	"strings"
	"sync"
)

type Proc struct {
	PID     int
	PPID    int
	Image   string
	Cmdline string
	StartTS int64
}

type Table struct {
	mu    sync.RWMutex
	procs map[int]Proc
}

func New() *Table { return &Table{procs: make(map[int]Proc)} }

func (t *Table) Add(p Proc) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.procs[p.PID] = p
}

func (t *Table) Remove(pid int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.procs, pid)
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
	t.mu.RLock()
	defer t.mu.RUnlock()

	children := map[int][]int{}
	for _, p := range t.procs {
		children[p.PPID] = append(children[p.PPID], p.PID)
	}
	var order []int
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
		order = append(order, id) // post-order => leaves first
	}
	visit(pid)
	return order
}

func (t *Table) FindByImage(imagePath string) []int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	target := strings.ToLower(imagePath)
	var out []int
	for pid, p := range t.procs {
		if strings.ToLower(p.Image) == target {
			out = append(out, pid)
		}
	}
	return out
}
