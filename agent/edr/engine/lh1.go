package engine

// LH1's adaptive tree and fixed position-code layout follow Simon Howard's
// ISC-licensed Lhasa decoder (https://github.com/fragglet/lhasa, lh1_decoder.c).
// This is a Go decoder with no runtime dependency on Lhasa or ClamAV.
//
// Copyright (c) 2011, 2012, Simon Howard
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

import "context"

const (
	lh1Codes = 314
	lh1Nodes = lh1Codes*2 - 1
)

type lh1Node struct {
	leaf   bool
	child  int // code for a leaf, right-child index otherwise
	parent int
	freq   int
	group  int
}

type lh1State struct {
	nodes       [lh1Nodes]lh1Node
	leaves      [lh1Codes]int
	groups      [lh1Nodes]int
	leaders     [lh1Nodes]int
	numGroups   int
	offsetCode  [256]byte
	offsetWidth [64]byte
	in          []byte
	bit         int
	history     [4096]byte
	write       int
}

func (d *lh1State) group() int {
	g := d.groups[d.numGroups]
	d.numGroups++
	return g
}

func (d *lh1State) release(g int) {
	d.numGroups--
	d.groups[d.numGroups] = g
}

func (d *lh1State) resetGroups() {
	for i := range d.groups {
		d.groups[i] = i
	}
	d.numGroups = 0
}

func (d *lh1State) init(in []byte) {
	d.in = in
	d.resetGroups()
	g := d.group()
	at := lh1Nodes - 1
	for code := 0; code < lh1Codes; code++ {
		d.nodes[at] = lh1Node{leaf: true, child: code, freq: 1, group: g}
		d.leaders[g] = at
		d.leaves[code] = at
		at--
	}
	child := lh1Nodes - 1
	for at >= 0 {
		freq := d.nodes[child].freq + d.nodes[child-1].freq
		if freq != d.nodes[at+1].freq {
			g = d.group()
		}
		d.nodes[at] = lh1Node{child: child, freq: freq, group: g}
		d.nodes[child].parent = at
		d.nodes[child-1].parent = at
		d.leaders[g] = at
		at--
		child -= 2
	}
	code := 0
	offset := 0
	for i, count := range [...]int{1, 3, 8, 12, 24, 16} {
		width := i + 3
		step := 1 << (8 - width)
		for range count {
			for j := 0; j < step; j++ {
				d.offsetCode[(code|j)&255] = byte(offset)
			}
			d.offsetWidth[offset] = byte(width)
			code += step
			offset++
		}
	}
	for i := range d.history {
		d.history[i] = ' '
	}
}

func (d *lh1State) read(n int) (int, bool) {
	if len(d.in)*8-d.bit < n {
		return 0, false
	}
	v := 0
	for range n {
		v = v<<1 | int(d.in[d.bit>>3]>>(7-(d.bit&7))&1)
		d.bit++
	}
	return v, true
}

func (d *lh1State) peek8() int {
	v := 0
	for i := 0; i < 8; i++ {
		v <<= 1
		at := d.bit + i
		if at < len(d.in)*8 {
			v |= int(d.in[at>>3] >> (7 - (at & 7)) & 1)
		}
	}
	return v
}

func (d *lh1State) lead(at int) int {
	g := d.nodes[at].group
	leader := d.leaders[g]
	if leader == at {
		return at
	}
	a, b := &d.nodes[at], &d.nodes[leader]
	a.leaf, b.leaf = b.leaf, a.leaf
	a.child, b.child = b.child, a.child
	for _, at := range [...]int{at, leader} {
		node := &d.nodes[at]
		if node.leaf {
			d.leaves[node.child] = at
		} else {
			d.nodes[node.child].parent = at
			d.nodes[node.child-1].parent = at
		}
	}
	return leader
}

func (d *lh1State) increment(at int) {
	node := &d.nodes[at]
	previous := &d.nodes[at-1]
	node.freq++
	if at < lh1Nodes-1 && node.group == d.nodes[at+1].group {
		d.leaders[node.group]++
		if node.freq == previous.freq {
			node.group = previous.group
		} else {
			node.group = d.group()
			d.leaders[node.group] = at
		}
	} else if node.freq == previous.freq {
		d.release(node.group)
		node.group = previous.group
	}
}

func (d *lh1State) rebuild() {
	leaf := 0
	for i := 0; i < lh1Nodes; i++ {
		if d.nodes[i].leaf {
			d.nodes[leaf].leaf = true
			d.nodes[leaf].child = d.nodes[i].child
			d.nodes[leaf].freq = (d.nodes[i].freq + 1) / 2
			leaf++
		}
	}
	leaf = lh1Codes - 1
	child := lh1Nodes - 1
	for i := lh1Nodes - 1; i >= 0; {
		for child-i < 2 {
			d.nodes[i] = d.nodes[leaf]
			d.leaves[d.nodes[leaf].child] = i
			i--
			leaf--
		}
		freq := d.nodes[child].freq + d.nodes[child-1].freq
		for leaf >= 0 && freq >= d.nodes[leaf].freq {
			d.nodes[i] = d.nodes[leaf]
			d.leaves[d.nodes[leaf].child] = i
			i--
			leaf--
		}
		d.nodes[i].leaf = false
		d.nodes[i].freq = freq
		d.nodes[i].child = child
		d.nodes[child].parent = i
		d.nodes[child-1].parent = i
		i--
		child -= 2
	}
	d.resetGroups()
	g := d.group()
	d.nodes[0].group = g
	d.leaders[g] = 0
	for i := 1; i < lh1Nodes; i++ {
		if d.nodes[i].freq != d.nodes[i-1].freq {
			g = d.group()
			d.leaders[g] = i
		}
		d.nodes[i].group = g
	}
}

func (d *lh1State) code() (int, bool) {
	at := 0
	for !d.nodes[at].leaf {
		bit, ok := d.read(1)
		if !ok {
			return 0, false
		}
		at = d.nodes[at].child - bit
	}
	code := d.nodes[at].child
	if d.nodes[0].freq >= 32*1024 {
		d.rebuild()
	}
	d.nodes[0].freq++
	at = d.leaves[code]
	for at != 0 {
		at = d.lead(at)
		d.increment(at)
		at = d.nodes[at].parent
	}
	return code, true
}

func (d *lh1State) offset() (int, bool) {
	high := int(d.offsetCode[d.peek8()])
	if _, ok := d.read(int(d.offsetWidth[high])); !ok {
		return 0, false
	}
	low, ok := d.read(6)
	return high<<6 | low, ok
}

// lh1Expand decodes the original LHarc adaptive-Huffman method. The caller
// has already checked that size is within the scan's per-file and aggregate
// limits; output is charged even when decoding fails.
func lh1Expand(ctx context.Context, in []byte, size int) ([]byte, error) {
	d := new(lh1State)
	d.init(in)
	out := arjInitialOutput(size, len(in))
	nextContextCheck := 0
	put := func(c byte) {
		out = append(out, c)
		d.history[d.write] = c
		d.write = (d.write + 1) & 4095
	}
	for len(out) < size {
		if len(out) >= nextContextCheck {
			if err := ctx.Err(); err != nil {
				return out, err
			}
			nextContextCheck = len(out) + 4096
		}
		code, ok := d.code()
		if !ok {
			return out, ErrMalformed
		}
		if code < 256 {
			put(byte(code))
			continue
		}
		offset, ok := d.offset()
		if !ok {
			return out, ErrMalformed
		}
		start := (d.write - offset - 1 + 4096) & 4095
		for i := 0; i < code-256+3 && len(out) < size; i++ {
			put(d.history[(start+i)&4095])
		}
	}
	return out, nil
}
