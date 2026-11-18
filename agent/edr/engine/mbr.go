package engine

import (
	"encoding/binary"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

const mbrSectorSize = 512

// mbr scans a 512-byte-sector disk image's boot record, primary partitions
// and conventional extended-partition chains. Protective GPT entries route
// to the GPT reader; its EBR reader also handles alternate first-two-entry
// orderings while malformed and cyclic chains remain incomplete.
// The reference also scans the boot sector as ANY and partitions as PART_ANY
// (or the matching partition-magic type), so their exact hashes and raw rules
// are matched on separate object layers.
func (s *scanState) mbr(data []byte, depth int) (string, error) {
	if len(data) < mbrSectorSize {
		return "", ErrMalformed
	}
	// The reference dispatches a protective/hybrid first entry to GPT
	// before applying the ordinary MBR partition-status and range checks.
	if data[446+4] == 0xee || data[446+4] == 0xed {
		return s.gpt(data, depth)
	}
	if len(data) < 2*mbrSectorSize || len(data)%mbrSectorSize != 0 || data[510] != 0x55 || data[511] != 0xaa {
		return "", ErrMalformed
	}
	type entry struct {
		kind       byte
		start, end int
	}
	var entries [4]entry
	for i := range entries {
		p := data[446+16*i:]
		if p[0] != 0 && p[0] != 0x80 {
			return "", ErrMalformed
		}
		first, length := uint64(binary.LittleEndian.Uint32(p[8:])), uint64(binary.LittleEndian.Uint32(p[12:]))
		if first+length > uint64(len(data)/mbrSectorSize) {
			return "", ErrMalformed
		}
		entries[i] = entry{kind: p[4], start: int(first) * mbrSectorSize, end: int(first+length) * mbrSectorSize}
	}
	any := signatures.TypeAny
	var incomplete error
	if sig, err := s.visitObject(data[:mbrSectorSize], depth+1, objectOptions{forcedType: &any}); sig != "" {
		return sig, nil
	} else if err != nil {
		incomplete = err
	}
	// ClamAV checks the entire primary table before scanning any partition.
	// The opt-in alert includes an extended partition's outer bounds.
	if s.partitionIntersectionHeuristicOn {
		var seen []partitionRange
		count := 0
		for _, e := range entries {
			if count >= s.partitionLimit() {
				break
			}
			if e.kind == 0 {
				count++
				continue
			}
			if e.start == e.end {
				count++
				continue
			}
			first, end := uint64(e.start/mbrSectorSize), uint64(e.end/mbrSectorSize)
			if partitionIntersects(seen, first, end) {
				return "Heuristics.MBRPartitionnIntersect", nil
			}
			seen = append(seen, partitionRange{first, end})
			if e.kind == 0x05 {
				if mbrExtendedIntersects(data, e.start, e.end, &count, s.partitionLimit()) {
					return "Heuristics.MBRPartitionnIntersect", nil
				}
			} else {
				count++
			}
		}
	}
	partitionCount := 0
	for _, e := range entries {
		if partitionCount >= s.partitionLimit() {
			break
		}
		switch e.kind {
		case 0:
			partitionCount++
			continue
		case 0x05:
			sig, err := s.mbrExtended(data, depth, e.start, e.end, &partitionCount)
			if sig != "" {
				return sig, nil
			}
			if err != nil && incomplete == nil {
				incomplete = err
			}
			continue
		}
		partitionCount++
		if e.start == e.end {
			continue
		}
		sig, err := s.mbrPartition(data[e.start:e.end], depth)
		if sig != "" {
			return sig, nil
		}
		if err != nil && incomplete == nil {
			incomplete = err
		}
	}
	return "", incomplete
}

// The reference checks logical EBR ranges separately from primary ranges.
// Its logical check uses the EBR's offset from the extended partition base
// and entry zero's sector count; entry one is the chain link.
func mbrExtendedIntersects(data []byte, base, end int, count *int, limit int) bool {
	if base < mbrSectorSize || end > len(data) || end-base < mbrSectorSize {
		return false
	}
	var seen []partitionRange
	visited := map[int]bool{}
	current := base
	for *count < limit {
		if current < base || current+mbrSectorSize > end || visited[current] {
			return false
		}
		visited[current] = true
		*count++
		ebr := data[current : current+mbrSectorSize]
		first := uint64((current - base) / mbrSectorSize)
		n := uint64(binary.LittleEndian.Uint32(ebr[446+12:]))
		if partitionIntersects(seen, first, first+n) {
			return true
		}
		if n > 0 {
			seen = append(seen, partitionRange{first, first + n})
		}
		if ebr[462+4] != 0x05 {
			return false
		}
		link := uint64(binary.LittleEndian.Uint32(ebr[462+8:]))
		if link == 0 || link > uint64((end-base)/mbrSectorSize-1) {
			return false
		}
		current = base + int(link)*mbrSectorSize
	}
	return false
}

func (s *scanState) mbrPartition(part []byte, depth int) (string, error) {
	typ := s.db.PartitionType(part)
	return s.visitObject(part, depth+1, objectOptions{forcedType: &typ})
}

// mbrExtended follows the reference's two-entry EBR state transitions. Most
// images put a logical partition first and an extended link second, but the
// reference also scans a partition in entry 1 and follows a link in entry 0.
func (s *scanState) mbrExtended(data []byte, depth, base, end int, count *int) (string, error) {
	if base < mbrSectorSize || end > len(data) || end-base < mbrSectorSize {
		return "", ErrMalformed
	}
	current := base
	seen := map[int]bool{}
	var incomplete error
	var linkLBA uint64
	for *count < s.partitionLimit() {
		if current < base || current+mbrSectorSize > end || seen[current] {
			return "", ErrMalformed
		}
		seen[current] = true
		*count++
		ebr := data[current : current+mbrSectorSize]
		if ebr[510] != 0x55 || ebr[511] != 0xaa {
			return "", ErrMalformed
		}
		for i := 0; i < 2; i++ {
			if status := ebr[446+16*i]; status != 0 && status != 0x80 {
				return "", ErrMalformed
			}
		}
		if ebr[446+2*16+4] != 0 || ebr[446+3*16+4] != 0 {
			return "", ErrMalformed
		}
		const (
			seenNothing = iota
			seenPartition
			seenExtended
			seenEmpty
		)
		state := seenNothing
		for i := 0; i < 2; i++ {
			entry := ebr[446+i*16 : 446+(i+1)*16]
			rel, n := uint64(binary.LittleEndian.Uint32(entry[8:])), uint64(binary.LittleEndian.Uint32(entry[12:]))
			switch entry[4] {
			case 0:
				switch state {
				case seenNothing:
					state = seenEmpty
				case seenPartition, seenEmpty:
					linkLBA = 0
				}
			case 0x05:
				if state == seenExtended {
					return "", ErrMalformed
				}
				if state == seenNothing {
					state = seenExtended
				}
				linkLBA = rel
			default:
				switch state {
				case seenNothing:
					state = seenPartition
				case seenPartition, seenEmpty:
					linkLBA = 0
				}
				start := uint64(base) + (linkLBA+rel)*mbrSectorSize
				finish := start + n*mbrSectorSize
				if finish > uint64(end) || start < uint64(base) {
					return "", ErrMalformed
				}
				if n != 0 {
					if sig, err := s.mbrPartition(data[int(start):int(finish)], depth); sig != "" {
						return sig, nil
					} else if err != nil && incomplete == nil {
						incomplete = err
					}
				}
			}
		}
		if linkLBA == 0 {
			return "", incomplete
		}
		next := uint64(base) + linkLBA*mbrSectorSize
		if next > uint64(end-mbrSectorSize) {
			return "", ErrMalformed
		}
		current = int(next)
	}
	return "", incomplete
}
