package engine

import (
	"bytes"
	"encoding/binary"
)

const apmEntrySize = 136

type apmEntry struct {
	first, count uint64
	kind         []byte
}

func apmEntryAt(data []byte, sector, index int) (apmEntry, bool) {
	var out apmEntry
	if sector <= 0 || index <= 0 || uint64(index)*uint64(sector) > uint64(len(data)) ||
		uint64(apmEntrySize) > uint64(len(data))-uint64(index)*uint64(sector) {
		return out, false
	}
	p := data[index*sector:]
	out.first = uint64(binary.BigEndian.Uint32(p[8:]))
	out.count = uint64(binary.BigEndian.Uint32(p[12:]))
	out.kind = p[48:80]
	return out, true
}

func apmKindIs(kind []byte, names ...string) bool {
	if end := bytes.IndexByte(kind, 0); end >= 0 {
		kind = kind[:end]
	}
	for _, name := range names {
		if bytes.Equal(kind, []byte(name)) {
			return true
		}
	}
	return false
}

func apmMapKind(kind []byte) bool {
	return apmKindIs(kind, "Apple_Partition_Map", "Apple_partition_map", "Apple_patition_map")
}

func apmDriverKind(kind []byte) bool {
	return apmKindIs(kind, "Apple_Driver", "Apple_Driver43", "Apple_Driver43_CD",
		"Apple_Driver_ATA", "Apple_Driver_ATAPI", "Apple_Patches")
}

// apm scans the partitions described by an Apple Partition Map. A 2048-byte
// driver description can use legacy 512-byte map entries; its aligned driver
// entries use a 2048-byte extent even though other entries use 512-byte blocks.
// Bounds are checked in uint64 before slicing any child window.
func (s *scanState) apm(data []byte, depth int) (string, error) {
	if len(data) < 82 || !bytes.Equal(data[:2], []byte("ER")) {
		return "", ErrMalformed
	}
	blockSize := int(binary.BigEndian.Uint16(data[2:]))
	blockCount := uint64(binary.BigEndian.Uint32(data[4:]))
	if blockSize == 0 || uint64(blockSize)*blockCount != uint64(len(data)) {
		return "", ErrMalformed
	}
	sector, oldSchool := blockSize, false
	if blockSize == 2048 && len(data) >= 512+apmEntrySize && bytes.Equal(data[512:514], []byte("PM")) {
		sector, oldSchool = 512, true
	}
	first, ok := apmEntryAt(data, sector, 1)
	if !ok || !bytes.Equal(data[sector:sector+2], []byte("PM")) || !apmMapKind(first.kind) {
		return "", ErrMalformed
	}
	count := min(int(binary.BigEndian.Uint32(data[sector+4:])), s.partitionLimit())
	if s.partitionIntersectionHeuristicOn {
		seen := make([]partitionRange, 0, count)
		for i := 1; i <= count; i++ {
			entry, ok := apmEntryAt(data, sector, i)
			if !ok {
				return "", ErrMalformed
			}
			blocks := entry.count
			if oldSchool && i%4 == 0 && apmDriverKind(entry.kind) {
				blocks *= 4
			}
			if partitionIntersects(seen, entry.first, entry.first+blocks) {
				return "Heuristics.APMPartitionIntersection", nil
			}
			seen = append(seen, partitionRange{entry.first, entry.first + blocks})
		}
	}
	var incomplete error
	for i := 2; i <= count; i++ {
		entry, ok := apmEntryAt(data, sector, i)
		if !ok {
			return "", ErrMalformed
		}
		if apmMapKind(entry.kind) {
			continue
		}
		length := entry.count * uint64(sector)
		if oldSchool && i%4 == 0 && apmDriverKind(entry.kind) {
			length = entry.count * 2048
		}
		start := entry.first * uint64(sector)
		if start == 0 || start > uint64(len(data)) || length > uint64(len(data))-start {
			continue
		}
		if sig, err := s.mbrPartition(data[int(start):int(start+length)], depth); sig != "" {
			return sig, nil
		} else if err != nil && incomplete == nil {
			incomplete = err
		}
	}
	return "", incomplete
}
