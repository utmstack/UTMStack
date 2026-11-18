package signatures

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"slices"
	"sort"
)

// hashWildcard encodes an unknown object size; it sorts after every real size.
const hashWildcard = ^uint32(0)

// hashTable is a compact, immutable-after-freeze digest index. Official feeds
// carry millions of unique digest rules, so records are packed into sorted
// byte arrays instead of per-entry map nodes and string headers. Records are
// grouped by digest width (MD5, SHA1, SHA256); each record is the digest
// followed by the big-endian object size. The first rule loaded for a
// (digest, size) pair wins, as with the earlier map representation.
type hashTable struct {
	groups [3]hashGroup
	// A whole-file database can consist of ordered immutable source tables.
	// This avoids retaining a second copy of every digest during feed reloads.
	shards         []*hashTable
	distinctCount  int
	frozen         bool
	referenceOrder bool
	referenceReady bool
	referenceNames map[referenceHashKey]string
	sourcePriority bool
}

type hashGroup struct {
	width   int
	records []byte   // len = n*(width+4), sorted after freeze
	offsets []uint32 // unsorted name starts before freeze
	arena   []byte
	// spans maps sorted records to length-prefixed names in the arena.
	spans           []uint32
	count           int
	keepSourceOrder bool
	sourceOrder     []uint32
	alternateNames  map[uint32]string
}

func newHashTable() *hashTable {
	t := &hashTable{}
	for i, w := range [3]int{16, 20, 32} {
		t.groups[i].width = w
	}
	return t
}

func newDetectionHashTable() *hashTable {
	t := newHashTable()
	t.referenceOrder = true
	for i := range t.groups {
		t.groups[i].keepSourceOrder = true
	}
	return t
}

// Whole-file sources have an explicit publisher order: an earlier local or
// official label takes precedence over a redundant supplemental IOC label.
// Candidate classification still follows the reference when duplicate classes
// differ; a source-name preference cannot turn weak evidence into malware.
func newWholeHashIndex() *hashTable {
	t := newDetectionHashTable()
	t.sourcePriority = true
	return t
}

func (t *hashTable) group(width int) *hashGroup {
	switch width {
	case 16:
		return &t.groups[0]
	case 20:
		return &t.groups[1]
	case 32:
		return &t.groups[2]
	}
	return nil
}

func encodeHashSize(size int64) uint32 {
	if size < 0 {
		return hashWildcard
	}
	return uint32(size)
}

func hashNameEncodedSize(n int) uint64 {
	var prefix [binary.MaxVarintLen64]byte
	return uint64(binary.PutUvarint(prefix[:], uint64(n))) + uint64(n)
}

// add appends a rule; duplicates are resolved when the table is frozen.
func (t *hashTable) add(digest string, size int64, name string) bool {
	return t.addBytes([]byte(digest), size, []byte(name))
}

func (t *hashTable) addBytes(digest []byte, size int64, name []byte) bool {
	g := t.group(len(digest))
	if g == nil || t.frozen || size < -1 || size >= int64(hashWildcard) || uint64(len(g.arena))+hashNameEncodedSize(len(name)) > uint64(^uint32(0)) {
		return false
	}
	g.records = append(g.records, digest...)
	g.records = binary.BigEndian.AppendUint32(g.records, encodeHashSize(size))
	g.offsets = append(g.offsets, uint32(len(g.arena)))
	g.arena = binary.AppendUvarint(g.arena, uint64(len(name)))
	g.arena = append(g.arena, name...)
	return true
}

// reserve uses a bounded, publisher-authenticated member preflight to avoid
// repeated growth of large packed hash arrays during full-feed loading.
func (t *hashTable) reserve(hint hashReserve) {
	if t == nil || t.frozen || len(t.shards) != 0 {
		return
	}
	for i, need := range hint {
		if need.count == 0 {
			continue
		}
		g := &t.groups[i]
		if want := len(g.records) + need.count*(g.width+4); want > cap(g.records) {
			buf := make([]byte, len(g.records), want)
			copy(buf, g.records)
			g.records = buf
		}
		if want := len(g.offsets) + need.count; want > cap(g.offsets) {
			buf := make([]uint32, len(g.offsets), want)
			copy(buf, g.offsets)
			g.offsets = buf
		}
		if want := len(g.arena) + need.nameBytes; want > cap(g.arena) && uint64(want) <= uint64(^uint32(0)) {
			buf := make([]byte, len(g.arena), want)
			copy(buf, g.arena)
			g.arena = buf
		}
	}
}

// freeze sorts each group and drops later duplicates. It must run once, after
// the last add and before any lookup.
func (t *hashTable) freeze() error {
	t.freezeStorage()
	if t.sourcePriority {
		for _, shard := range t.shards {
			if err := shard.freeze(); err != nil {
				return err
			}
		}
	}
	if t.referenceOrder && !t.referenceReady {
		if err := t.prepareReferenceNames(); err != nil {
			return err
		}
		t.referenceReady = true
	}
	return nil
}

func (t *hashTable) freezeStorage() {
	if t.frozen {
		return
	}
	if len(t.shards) > 0 {
		for _, shard := range t.shards {
			shard.freezeStorage()
		}
		t.distinctCount = distinctShardRecords(t.shards)
		t.frozen = true
		return
	}
	t.frozen = true
	for i := range t.groups {
		t.groups[i].freeze()
	}
}

func (g *hashGroup) freeze() {
	rw := g.width + 4
	n := len(g.records) / rw
	if n == 0 {
		g.records, g.arena, g.offsets = nil, nil, nil
		g.spans, g.count = nil, 0
		return
	}
	order := make([]uint32, n)
	for i := range order {
		order[i] = uint32(i)
	}
	rec := func(i uint32) []byte { return g.records[int(i)*rw : int(i+1)*rw] }
	// Ties break on load order, so the first-loaded rule sorts first.
	slices.SortFunc(order, func(a, b uint32) int {
		if c := bytes.Compare(rec(a), rec(b)); c != 0 {
			return c
		}
		return cmp.Compare(a, b)
	})
	// Sort records and their name offsets together. The offset slice is
	// compile-only storage until this point; reusing it as the final spans
	// avoids a second four-byte-per-rule array at peak load.
	spans := g.offsets
	visited := newBitset(n)
	scratch := make([]byte, rw)
	for start := 0; start < n; start++ {
		if visited.has(start) {
			continue
		}
		copy(scratch, rec(uint32(start)))
		scratchSpan := spans[start]
		for dst := start; ; {
			src := int(order[dst])
			visited.set(dst)
			if src == start {
				copy(rec(uint32(dst)), scratch)
				spans[dst] = scratchSpan
				break
			}
			copy(rec(uint32(dst)), rec(uint32(src)))
			spans[dst] = spans[src]
			dst = src
		}
	}
	kept, nameBytes := 0, 0
	if g.keepSourceOrder {
		g.sourceOrder = make([]uint32, n)
	}
	for k := 0; k < n; k++ {
		if k > 0 && bytes.Equal(rec(uint32(k-1)), rec(uint32(k))) {
			if g.keepSourceOrder {
				original := order[k]
				g.sourceOrder[original] = uint32(kept - 1)
				name := g.nameAt(spans[k])
				if name != g.nameAt(spans[kept-1]) {
					if g.alternateNames == nil {
						g.alternateNames = make(map[uint32]string)
					}
					g.alternateNames[original] = name
				}
			}
			continue // the earlier loaded name wins on a duplicate
		}
		if g.keepSourceOrder {
			g.sourceOrder[order[k]] = uint32(kept)
		}
		if kept != k {
			copy(rec(uint32(kept)), rec(uint32(k)))
			spans[kept] = spans[k]
		}
		nameBytes += g.nameEncodedSize(spans[kept])
		kept++
	}
	records := g.records[:kept*rw]
	// Heavy duplicate sets need compact storage. Nearly unique official
	// feeds retain the append buffer and avoid another peak-size copy.
	if cap(records) > len(records)+len(records)/4 {
		compact := make([]byte, len(records))
		copy(compact, records)
		records = compact
	}
	// Duplicate-heavy inputs can leave most of the compile-time offset
	// array unused. Compact those cases; a nearly unique source keeps the
	// original allocation even if append growth left some spare capacity.
	if kept < n-n/8 {
		spans = slices.Clone(spans[:kept])
	}
	g.count = kept
	// For the usual nearly unique feed, copying the entire name arena is a
	// large avoidable allocation while the old generation is still live.
	// Retain it only when duplicate names and unused capacity together cost
	// less than a modest fraction; duplicate-heavy inputs still compact.
	reuseArena := nameBytes >= len(g.arena)-len(g.arena)/8 && cap(g.arena) <= len(g.arena)+len(g.arena)/3
	if reuseArena {
		g.records, g.spans, g.offsets = records, spans[:kept:kept], nil
		return
	}
	arena := make([]byte, nameBytes)
	pos := 0
	for k, start := range spans[:kept] {
		n := g.nameEncodedSize(start)
		pos += copy(arena[pos:], g.arena[start:int(start)+n])
		spans[k] = uint32(pos - n)
	}
	g.records, g.spans, g.offsets, g.arena = records, spans[:kept:kept], nil, arena
}

func (g *hashGroup) nameEncodedSize(start uint32) int {
	length, prefix := binary.Uvarint(g.arena[start:])
	return prefix + int(length)
}

func (g *hashGroup) nameAt(start uint32) string {
	length, prefix := binary.Uvarint(g.arena[start:])
	return string(g.arena[int(start)+prefix : int(start)+prefix+int(length)])
}

func (g *hashGroup) name(index int) string { return g.nameAt(g.spans[index]) }

func (t *hashTable) len() int {
	if len(t.shards) > 0 {
		if t.frozen {
			return t.distinctCount
		}
		n := 0
		for _, shard := range t.shards {
			n += shard.len()
		}
		return n
	}
	n := 0
	for i := range t.groups {
		if t.frozen {
			n += t.groups[i].count
		} else {
			n += len(t.groups[i].offsets)
		}
	}
	return n
}

type hashCursor struct {
	group *hashGroup
	index int
}

func (c hashCursor) record() []byte {
	rw := c.group.width + 4
	return c.group.records[c.index*rw : (c.index+1)*rw]
}

// distinctShardRecords counts the union without allocating a map proportional
// to the full feed. Each source group is already sorted after freeze.
func distinctShardRecords(shards []*hashTable, duplicate ...func(hashCursor, hashCursor)) int {
	// Archive sources contain ordered member shards. Flatten only the small
	// shard list; records remain in their immutable, compact tables.
	var leaves []*hashTable
	var visit func(*hashTable)
	visit = func(shard *hashTable) {
		if len(shard.shards) == 0 {
			leaves = append(leaves, shard)
			return
		}
		for _, child := range shard.shards {
			visit(child)
		}
	}
	for _, shard := range shards {
		visit(shard)
	}
	total := 0
	for widthIndex := 0; widthIndex < 3; widthIndex++ {
		heap := make([]hashCursor, 0, len(leaves))
		push := func(cursor hashCursor) {
			heap = append(heap, cursor)
			for child := len(heap) - 1; child > 0; {
				parent := (child - 1) / 2
				if bytes.Compare(heap[parent].record(), heap[child].record()) <= 0 {
					break
				}
				heap[parent], heap[child] = heap[child], heap[parent]
				child = parent
			}
		}
		pop := func() hashCursor {
			first := heap[0]
			last := heap[len(heap)-1]
			heap = heap[:len(heap)-1]
			if len(heap) > 0 {
				heap[0] = last
				for parent := 0; ; {
					child := parent*2 + 1
					if child >= len(heap) {
						break
					}
					if child+1 < len(heap) && bytes.Compare(heap[child+1].record(), heap[child].record()) < 0 {
						child++
					}
					if bytes.Compare(heap[parent].record(), heap[child].record()) <= 0 {
						break
					}
					heap[parent], heap[child] = heap[child], heap[parent]
					parent = child
				}
			}
			return first
		}
		for _, shard := range leaves {
			group := &shard.groups[widthIndex]
			if group.count > 0 {
				push(hashCursor{group: group})
			}
		}
		var last []byte
		var previous hashCursor
		for len(heap) > 0 {
			cursor := pop()
			record := cursor.record()
			if last == nil || !bytes.Equal(last, record) {
				total++
				last = record
				previous = cursor
			} else if len(duplicate) != 0 {
				duplicate[0](previous, cursor)
			}
			cursor.index++
			if cursor.index < cursor.group.count {
				push(cursor)
			}
		}
	}
	return total
}

func (t *hashTable) lookup(digest string, size int64) string {
	if size >= 0 && size < int64(hashWildcard) {
		if n := t.findExact(digest, uint32(size)); n != "" {
			return n
		}
	}
	return t.findExact(digest, hashWildcard)
}

func (t *hashTable) findExact(digest string, size uint32) string {
	if t.referenceNames != nil {
		if name, ok := t.referenceNames[makeReferenceHashKey([]byte(digest), size)]; ok {
			return name
		}
	}
	if len(t.shards) > 0 {
		for _, shard := range t.shards {
			if name := shard.findExact(digest, size); name != "" {
				return name
			}
		}
		return ""
	}
	g := t.group(len(digest))
	if g == nil || len(g.records) == 0 {
		return ""
	}
	return g.find(digest, size)
}

func (g *hashGroup) find(digest string, size uint32) string {
	rw := g.width + 4
	var key [36]byte
	copy(key[:], digest)
	binary.BigEndian.PutUint32(key[g.width:], size)
	k := key[:rw]
	n := len(g.records) / rw
	i := sort.Search(n, func(i int) bool { return bytes.Compare(g.records[i*rw:(i+1)*rw], k) >= 0 })
	if i < n && bytes.Equal(g.records[i*rw:(i+1)*rw], k) {
		start := g.spans[i]
		length, prefix := binary.Uvarint(g.arena[start:])
		return string(g.arena[int(start)+prefix : int(start)+prefix+int(length)])
	}
	return ""
}

// sizes returns the distinct object sizes named by frozen rules; -1 stands
// for a wildcard size.
func (t *hashTable) sizes() map[int64]bool {
	out := make(map[int64]bool)
	if len(t.shards) > 0 {
		for _, shard := range t.shards {
			for size := range shard.sizes() {
				out[size] = true
			}
		}
		return out
	}
	for i := range t.groups {
		g := &t.groups[i]
		rw := g.width + 4
		for off := g.width; off+4 <= len(g.records); off += rw {
			if v := binary.BigEndian.Uint32(g.records[off:]); v == hashWildcard {
				out[-1] = true
			} else {
				out[int64(v)] = true
			}
		}
	}
	return out
}
