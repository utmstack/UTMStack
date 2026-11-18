package engine

// A partition table uses inclusive LBA endpoints for GPT and a start/count
// pair for MBR. Keep the comparison in sectors so a sector shared by two
// entries is an intersection while adjacent entries are not.
type partitionRange struct {
	first, end uint64 // half-open interval
}

func partitionIntersects(seen []partitionRange, first, end uint64) bool {
	if first >= end {
		return false
	}
	for _, other := range seen {
		if first < other.end && other.first < end {
			return true
		}
	}
	return false
}

func (s *scanState) partitionLimit() int {
	if s.maxPartitions <= 0 {
		return gptMaxScanPart
	}
	return min(s.maxPartitions, FixedMaxPartitions)
}
