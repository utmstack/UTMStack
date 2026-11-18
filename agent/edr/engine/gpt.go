package engine

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
)

const (
	gptHeaderSize  = 92
	gptEntrySize   = 128
	gptMaxScanPart = 50 // ClamAV's default MaxPartitions
)

type gptHeader struct {
	firstUsable, lastUsable uint64
	tableStart              uint64
	tableCount              uint32
	tableCRC                uint32
	table                   []byte
}

func gptSectorSize(data []byte) int {
	for _, sector := range [...]int{512, 1024, 2048, 4096} {
		if len(data) >= sector+8 && bytes.Equal(data[sector:sector+8], []byte("EFI PART")) {
			return sector
		}
	}
	return 0
}

// gptHeaderAt validates both GPT checksums and the partition-table bounds
// before any partition offsets are converted to host ints.
func gptHeaderAt(data []byte, sector int, at int) (gptHeader, bool) {
	var out gptHeader
	if at < 0 || at > len(data)-gptHeaderSize {
		return out, false
	}
	h := data[at : at+gptHeaderSize]
	if !bytes.Equal(h[:8], []byte("EFI PART")) || binary.LittleEndian.Uint32(h[12:]) != gptHeaderSize || binary.LittleEndian.Uint32(h[20:]) != 0 {
		return out, false
	}
	checksum := binary.LittleEndian.Uint32(h[16:])
	var checked [gptHeaderSize]byte
	copy(checked[:], h)
	clear(checked[16:20])
	if crc32.ChecksumIEEE(checked[:]) != checksum {
		return out, false
	}
	last := uint64(len(data)/sector - 1)
	current, backup := binary.LittleEndian.Uint64(h[24:]), binary.LittleEndian.Uint64(h[32:])
	if !(current == 1 && backup == last || current == last && backup == 1) {
		return out, false
	}
	out.firstUsable, out.lastUsable = binary.LittleEndian.Uint64(h[40:]), binary.LittleEndian.Uint64(h[48:])
	if out.firstUsable <= 1 || out.firstUsable > out.lastUsable || out.lastUsable >= last {
		return out, false
	}
	out.tableStart = binary.LittleEndian.Uint64(h[72:])
	out.tableCount = binary.LittleEndian.Uint32(h[80:])
	if binary.LittleEndian.Uint32(h[84:]) != gptEntrySize {
		return out, false
	}
	out.tableCRC = binary.LittleEndian.Uint32(h[88:])
	tableBytes := uint64(out.tableCount) * gptEntrySize
	if out.tableStart <= 1 || out.tableStart >= last || out.tableStart > uint64(len(data)/sector) ||
		tableBytes > uint64(len(data))-out.tableStart*uint64(sector) {
		return out, false
	}
	tableEnd := out.tableStart
	if tableBytes != 0 {
		tableEnd += (tableBytes - 1) / uint64(sector)
	}
	if tableEnd >= last || out.tableStart <= out.lastUsable && tableEnd >= out.firstUsable {
		return out, false
	}
	start := int(out.tableStart * uint64(sector))
	out.table = data[start : start+int(tableBytes)]
	if crc32.ChecksumIEEE(out.table) != out.tableCRC {
		return out, false
	}
	return out, true
}

// gpt scans valid primary/secondary GUID partition tables. If both valid
// tables disagree by checksum, the reference scans both; a broken primary
// falls back to the valid secondary.
func (s *scanState) gpt(data []byte, depth int) (string, error) {
	sector := gptSectorSize(data)
	if sector == 0 || len(data)%sector != 0 || len(data) < 3*sector {
		return "", ErrMalformed
	}
	primary, primaryOK := gptHeaderAt(data, sector, sector)
	secondary, secondaryOK := gptHeaderAt(data, sector, len(data)-sector)
	if !primaryOK && !secondaryOK {
		return "", ErrMalformed
	}
	if s.partitionIntersectionHeuristicOn {
		check := func(h gptHeader) bool {
			var seen []partitionRange
			for i := 0; i < min(int(h.tableCount), s.partitionLimit()); i++ {
				entry := h.table[i*gptEntrySize : (i+1)*gptEntrySize]
				first, last := binary.LittleEndian.Uint64(entry[32:]), binary.LittleEndian.Uint64(entry[40:])
				if first == 0 || first > last || first < h.firstUsable || last > h.lastUsable || last >= uint64(len(data)/sector) {
					continue
				}
				if partitionIntersects(seen, first, last+1) {
					return true
				}
				seen = append(seen, partitionRange{first, last + 1})
			}
			return false
		}
		if primaryOK && check(primary) || secondaryOK && (!primaryOK || primary.tableCRC != secondary.tableCRC) && check(secondary) {
			return "Heuristics.GPTPartitionIntersection", nil
		}
	}
	var incomplete error
	scan := func(h gptHeader) (string, bool) {
		count := min(int(h.tableCount), s.partitionLimit())
		for i := 0; i < count; i++ {
			entry := h.table[i*gptEntrySize : (i+1)*gptEntrySize]
			first, last := binary.LittleEndian.Uint64(entry[32:]), binary.LittleEndian.Uint64(entry[40:])
			if first == 0 || first > last || first < h.firstUsable || last > h.lastUsable || last >= uint64(len(data)/sector) {
				continue
			}
			start, end := int(first*uint64(sector)), int((last+1)*uint64(sector))
			if sig, err := s.mbrPartition(data[start:end], depth); sig != "" {
				return sig, true
			} else if err != nil && incomplete == nil {
				incomplete = err
			}
		}
		return "", false
	}
	if primaryOK {
		if sig, found := scan(primary); found {
			return sig, nil
		}
	}
	if secondaryOK && (!primaryOK || primary.tableCRC != secondary.tableCRC) {
		if sig, found := scan(secondary); found {
			return sig, nil
		}
	}
	return "", incomplete
}
