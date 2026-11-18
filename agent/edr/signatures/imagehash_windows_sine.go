package signatures

import (
	_ "embed"
	"encoding/binary"
	"math"
)

// The pinned Windows reference's UCRT sinf differs from portable Go double
// sine at a finite subset of the admitted Lanczos3 arguments. This sorted
// table stores positive argument/result bit pairs; odd symmetry supplies the
// negative half. Native oracles cover every admitted dimension 1..16384,
// independently of image contents. Runtime analysis stays entirely in Go.
//
//go:embed resize_math_windows.bin
var resizeWindowsCorrections string

func resizeWindowsSine32(x float32) float32 {
	bits := math.Float32bits(x)
	key := bits & 0x7fffffff
	lo, hi := 0, len(resizeWindowsCorrections)/8
	for lo < hi {
		mid := int(uint(lo+hi) >> 1)
		value := binary.LittleEndian.Uint32([]byte(resizeWindowsCorrections[mid*8 : mid*8+4]))
		if value < key {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo < len(resizeWindowsCorrections)/8 && binary.LittleEndian.Uint32([]byte(resizeWindowsCorrections[lo*8:lo*8+4])) == key {
		value := binary.LittleEndian.Uint32([]byte(resizeWindowsCorrections[lo*8+4 : lo*8+8]))
		return math.Float32frombits(value ^ (bits & 0x80000000))
	}
	return float32(math.Sin(float64(x)))
}
