package signatures

import (
	"bytes"
	"cmp"
	"slices"
)

// The backtracking search (matchesAtWith) remembers failed (node, position)
// states and gives up at maxPatternStates. A pattern with a long gap — the
// official Win.Packed.Gandcrab-6552923-4 allows up to 196,608 bytes after
// "MZ" — reaches that bound on ordinary programs with long zero runs. The
// forward search answers the same question (does the pattern match from this
// start?) by carrying, node by node in the pattern's order, the set of
// positions each node can be reached at, kept as ranges: a gap turns a
// position into a range instead of one state per position.

// maxForwardSteps bounds the positions one start's forward search examines.
// Past it the start stays undecided (errPatternStates), which leaves every
// other rule sound. It is a variable only so tests can reach it.
var maxForwardSteps = 1 << 22

// posRanges is a set of positions kept as ranges [lo, hi]. Additions in
// increasing order merge as they come; others are appended and the set is
// sorted and merged once, when it is read (ranges).
type posRanges struct {
	r        [][2]int
	unsorted bool
}

func (s *posRanges) add(lo, hi int) {
	n := len(s.r)
	if n == 0 || !s.unsorted && lo > s.r[n-1][1]+1 {
		s.r = append(s.r, [2]int{lo, hi})
		return
	}
	if !s.unsorted && lo >= s.r[n-1][0] {
		s.r[n-1][1] = max(s.r[n-1][1], hi)
		return
	}
	s.r = append(s.r, [2]int{lo, hi})
	s.unsorted = true
}

// ranges returns the set as sorted, disjoint, non-adjacent ranges.
func (s *posRanges) ranges() [][2]int {
	if !s.unsorted {
		return s.r
	}
	slices.SortFunc(s.r, func(a, b [2]int) int { return cmp.Compare(a[0], b[0]) })
	merged := s.r[:0]
	for _, x := range s.r {
		if n := len(merged); n > 0 && x[0] <= merged[n-1][1]+1 {
			merged[n-1][1] = max(merged[n-1][1], x[1])
			continue
		}
		merged = append(merged, x)
	}
	s.r, s.unsorted = merged, false
	return s.r
}

func (s *posRanges) union(o [][2]int) {
	for _, x := range o {
		s.add(x[0], x[1])
	}
}

// forwardOrder lists the nodes reachable from the start so that every node
// comes before the nodes it leads to (the pattern graph has no cycles).
func (p *bodyPattern) forwardOrder() []int {
	var order []int
	seen := make(map[int]bool)
	var walk func(int)
	walk = func(id int) {
		if seen[id] {
			return
		}
		seen[id] = true
		if id != 0 {
			for _, alt := range p.graph.nodes[id].alternatives {
				walk(alt)
			}
			walk(p.graph.nodes[id].next)
		}
		order = append(order, id)
	}
	walk(p.graph.start)
	for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {
		order[i], order[j] = order[j], order[i]
	}
	return order
}

// forwardMatchesAt reports whether the pattern matches from start, as
// matchesAt does, in time and memory bounded by the positions examined
// rather than by remembered states.
func (p *bodyPattern) forwardMatchesAt(data []byte, start int, w *matchWork) (bool, error) {
	if p.simple {
		return p.matchesAt(data, start, w)
	}
	if start < 0 || start > len(data) {
		return false, nil
	}
	reach := make(map[int]*posRanges)
	at := func(id int) *posRanges {
		s := reach[id]
		if s == nil {
			s = &posRanges{}
			reach[id] = s
		}
		return s
	}
	at(p.graph.start).add(start, start)
	budget := maxForwardSteps
	spend := func(n int) error {
		if n < 1 {
			n = 1
		}
		if budget -= n; budget < 0 {
			return errPatternStates
		}
		return w.step(int64(n))
	}
	// each visits every position of s, charging them.
	each := func(s [][2]int, visit func(pos int)) error {
		for _, r := range s {
			if err := spend(r[1] - r[0] + 1); err != nil {
				return err
			}
			for pos := r[0]; pos <= r[1]; pos++ {
				visit(pos)
			}
		}
		return nil
	}
	for _, id := range p.forwardOrder() {
		set, ok := reach[id]
		if !ok {
			continue
		}
		if set.unsorted {
			// Sorting is charged like visiting the ranges.
			if err := spend(len(set.r)); err != nil {
				return false, err
			}
		}
		s := set.ranges()
		if len(s) == 0 {
			continue
		}
		if id == 0 {
			return true, nil
		}
		node := p.graph.nodes[id]
		t := node.token
		next := func() *posRanges { return at(node.next) }
		var err error
		switch {
		case t.boundary != 0:
			err = each(s, func(pos int) {
				if boundaryHolds(t.boundary, data, pos) {
					next().add(pos, pos)
				}
			})
		case len(node.alternatives) > 0:
			if err = spend(len(s) * len(node.alternatives)); err == nil {
				for _, alt := range node.alternatives {
					at(alt).union(s)
				}
			}
		case len(t.literal) > 0:
			n := len(t.literal)
			if p.nocase {
				err = each(s, func(pos int) {
					if pos+n <= len(data) && equalBytes(data[pos:pos+n], t.literal, true) {
						next().add(pos+n, pos+n)
					}
				})
				break
			}
			for _, r := range s {
				if r[0]+n > len(data) {
					break
				}
				window := data[r[0]:min(len(data), r[1]+n)]
				if err = spend(len(window)); err != nil {
					break
				}
				for off := 0; ; {
					i := bytes.Index(window[off:], t.literal)
					if i < 0 {
						break
					}
					pos := r[0] + off + i
					next().add(pos+n, pos+n)
					off += i + 1
				}
			}
		case t.gap:
			if err = spend(len(s)); err == nil {
				for _, r := range s {
					lo, hi := r[0]+t.min, len(data)
					if t.max >= 0 {
						hi = min(hi, r[1]+t.max)
					}
					if lo <= hi {
						next().add(lo, hi)
					}
				}
			}
		case t.negative:
			var literals [][]byte
			for _, seq := range t.alternatives {
				var literal []byte
				for _, x := range seq {
					literal = append(literal, x.literal...)
				}
				literals = append(literals, literal)
			}
			err = each(s, func(pos int) {
				if pos+t.min > len(data) {
					return
				}
				for _, literal := range literals {
					if equalBytes(data[pos:pos+t.min], literal, p.nocase) {
						return
					}
				}
				next().add(pos+t.min, pos+t.min)
			})
		case t.mask == 0 && t.value == 0:
			// Any byte: every position before the end moves by one.
			if err = spend(len(s)); err == nil {
				for _, r := range s {
					if hi := min(r[1], len(data)-1); r[0] <= hi {
						next().add(r[0]+1, hi+1)
					}
				}
			}
		default:
			err = each(s, func(pos int) {
				if pos < len(data) && data[pos]&t.mask == t.value {
					next().add(pos+1, pos+1)
				}
			})
		}
		if err != nil {
			return false, err
		}
		delete(reach, id)
	}
	return false, nil
}
