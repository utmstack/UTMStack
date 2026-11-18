package responder

import "sort"

type processParent struct{ pid, parent int }

// Merge ancestry before walking it. Concatenating two leaves-first lists can
// place a live parent before a child known only to the asynchronous table.
// The live subtree wins parent observations for PIDs it contains. Table-only
// historical descendants remain candidates, as with the prior union; absence
// from this subtree does not prove a process is gone or has reparented.
func mergeProcessTrees(root int, table, live []processParent) []int {
	parents := make(map[int]int, len(table)+len(live)+1)
	for _, p := range table {
		parents[p.pid] = p.parent
	}
	for _, p := range live {
		parents[p.pid] = p.parent
	}
	parents[root] = 0
	children := make(map[int][]int, len(parents))
	for pid, parent := range parents {
		if pid != root {
			children[parent] = append(children[parent], pid)
		}
	}
	for _, ids := range children {
		sort.Ints(ids)
	}
	order := make([]int, 0, len(parents))
	seen := make(map[int]bool, len(parents))
	var visit func(int)
	visit = func(pid int) {
		if seen[pid] {
			return
		}
		seen[pid] = true
		for _, child := range children[pid] {
			visit(child)
		}
		order = append(order, pid)
	}
	// Walk only this root; disconnected malformed inputs and cycles cannot
	// expand the response to unrelated roots.
	visit(root)
	return order
}
