package signatures

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

var errHashOrderBudget = fmt.Errorf("%w: signature hash candidate ordering work limit exceeded", ErrMatchLimit)

// referenceHashCompare describes the little-endian ClamAV 1.5.4 hash profile.
// The first word sorts in descending unsigned order; the remaining digest
// bytes sort lexically. All supported endpoint architectures are little-endian.
func referenceHashCompare(a, b []byte) int {
	x, y := binary.LittleEndian.Uint32(a), binary.LittleEndian.Uint32(b)
	if x < y {
		return 1
	}
	if x > y {
		return -1
	}
	return bytes.Compare(a[4:], b[4:])
}

// referenceHashPermutation preserves the pinned profile's candidate order,
// including equal digests. A general stable sort cannot select its duplicate
// name. Work is bounded and no published index is modified on a failed build.
// Equal-only ranges use the equivalent linear rotation instead of quadratic
// repeated partitions. The caller owns the index slice and immutable digests.
func referenceHashPermutation(order []uint32, digest func(uint32) []byte, budget int64) error {
	return referenceHashPermutationBudget(order, digest, &budget)
}

func referenceHashPermutationBudget(order []uint32, digest func(uint32) []byte, budget *int64) error {
	type interval struct{ start, end int }
	stack := []interval{{0, len(order)}}
	compare := func(a, b uint32) (int, error) {
		if *budget <= 0 {
			return 0, errHashOrderBudget
		}
		*budget--
		return referenceHashCompare(digest(a), digest(b)), nil
	}
	for len(stack) > 0 {
		block := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		start, end := block.start, block.end
		if end-start < 2 {
			continue
		}
		pivot := order[start]
		equal := true
		for _, index := range order[start+1 : end] {
			c, err := compare(index, pivot)
			if err != nil {
				return err
			}
			if c != 0 {
				equal = false
				break
			}
		}
		if equal {
			copy(order[start:end-1], order[start+1:end])
			order[end-1] = pivot
			continue
		}
		front, back := start+1, end
		for front < back {
			c, err := compare(order[front], pivot)
			if err != nil {
				return err
			}
			if c > 0 {
				back--
				order[front], order[back] = order[back], order[front]
			} else {
				front++
			}
		}
		split := front - 1
		order[start], order[split] = order[split], order[start]
		left, right := interval{start, split}, interval{back, end}
		// Process the smaller range first to bound the explicit stack even for
		// adversarial source ordering; partition order within each range is fixed.
		if left.end-left.start < right.end-right.start {
			stack = append(stack, right, left)
		} else {
			stack = append(stack, left, right)
		}
	}
	return nil
}

// Only ambiguous digest/size pairs need an override in the compact index.
// Source order is retained separately so incremental generations can reuse
// immutable shards without changing the whole-generation candidate decision.
type referenceHashKey struct {
	Digest [32]byte
	Size   uint32
	Width  uint8
}

type referenceHashBucket struct {
	Width int
	Size  uint32
}

type referenceHashEntry struct {
	group          *hashGroup
	slot, original uint32
}

func makeReferenceHashKey(digest []byte, size uint32) referenceHashKey {
	k := referenceHashKey{Size: size, Width: uint8(len(digest))}
	copy(k.Digest[:], digest)
	return k
}

func (e referenceHashEntry) digest() []byte {
	width := e.group.width
	start := int(e.slot) * (width + 4)
	return e.group.records[start : start+width]
}

func (e referenceHashEntry) name() string {
	if name, ok := e.group.alternateNames[e.original]; ok {
		return name
	}
	return e.group.name(int(e.slot))
}

func (t *hashTable) prepareReferenceNames() error {
	var leaves []*hashTable
	var walk func(*hashTable)
	walk = func(table *hashTable) {
		if len(table.shards) == 0 {
			leaves = append(leaves, table)
			return
		}
		for _, child := range table.shards {
			walk(child)
		}
	}
	walk(t)
	ambiguous := make(map[referenceHashKey]struct{})
	needed := make(map[referenceHashBucket][]referenceHashKey)
	mark := func(group *hashGroup, slot int) {
		width := group.width
		record := group.records[slot*(width+4) : (slot+1)*(width+4)]
		size := binary.BigEndian.Uint32(record[width:])
		key := makeReferenceHashKey(record[:width], size)
		if _, exists := ambiguous[key]; exists {
			return
		}
		ambiguous[key] = struct{}{}
		bucket := referenceHashBucket{width, size}
		needed[bucket] = append(needed[bucket], key)
	}
	for _, leaf := range leaves {
		for i := range leaf.groups {
			group := &leaf.groups[i]
			for original, alternate := range group.alternateNames {
				slot := int(group.sourceOrder[original])
				if t.referenceNameDiffers(group.name(slot), alternate) {
					mark(group, slot)
				}
			}
		}
	}
	if len(leaves) > 1 {
		distinctShardRecords(leaves, func(a, b hashCursor) {
			if t.referenceNameDiffers(a.group.name(a.index), b.group.name(b.index)) {
				mark(a.group, a.index)
			}
		})
	}
	if len(ambiguous) == 0 {
		return nil
	}
	buckets := make(map[referenceHashBucket][]referenceHashEntry)
	for _, leaf := range leaves {
		for i := range leaf.groups {
			group := &leaf.groups[i]
			if group.count > 0 && len(group.sourceOrder) == 0 {
				return errors.New("signature hash source order is unavailable")
			}
			width := group.width
			for original, slot := range group.sourceOrder {
				start := int(slot) * (width + 4)
				size := binary.BigEndian.Uint32(group.records[start+width : start+width+4])
				id := referenceHashBucket{width, size}
				if len(needed[id]) != 0 {
					buckets[id] = append(buckets[id], referenceHashEntry{group, slot, uint32(original)})
				}
			}
		}
	}
	names := make(map[referenceHashKey]string, len(ambiguous))
	// One ceiling covers every ambiguous size group in this table, including
	// adversarially ordered sources. A fixed allowance avoids rejecting small
	// legitimate sorted inputs merely because their partitions are quadratic.
	budget := int64(8_000_000)
	for _, entries := range buckets {
		budget += int64(len(entries)) * 128
	}
	for id, entries := range buckets {
		if uint64(len(entries)) > uint64(^uint32(0)) {
			return errHashOrderBudget
		}
		order := make([]uint32, len(entries))
		for i := range order {
			order[i] = uint32(i)
		}
		if err := referenceHashPermutationBudget(order, func(i uint32) []byte { return entries[i].digest() }, &budget); err != nil {
			return err
		}
		for _, key := range needed[id] {
			low, high := 0, len(order)-1
			for low <= high {
				middle := low + (high-low)/2
				entry := entries[order[middle]]
				c := referenceHashCompare(key.Digest[:key.Width], entry.digest())
				if c < 0 {
					high = middle - 1
				} else if c > 0 {
					low = middle + 1
				} else {
					names[key] = entry.name()
					break
				}
			}
			if _, ok := names[key]; !ok {
				return errors.New("signature hash candidate index is inconsistent")
			}
		}
	}
	t.referenceNames = names
	return nil
}

// Classification-sensitive aliases cannot use the ordinary source-name
// preference. These are the exact, case-sensitive prefixes in the pinned
// scanner's evidence classifier; source selection is separate from alerting.
func referenceHashClass(name string) uint8 {
	if strings.HasPrefix(name, "Weak.") {
		return 2
	}
	for _, prefix := range []string{"PUA.", "Heuristics.", "BC.Heuristics."} {
		if strings.HasPrefix(name, prefix) {
			return 1
		}
	}
	return 0
}
func (t *hashTable) referenceNameDiffers(a, b string) bool {
	if t.sourcePriority && len(t.shards) > 0 {
		return referenceHashClass(a) != referenceHashClass(b)
	}
	return a != b
}
