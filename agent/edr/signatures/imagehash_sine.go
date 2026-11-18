package signatures

import (
	_ "embed"
	"encoding/binary"
	"math"
)

// resizeSine32 is the bounded, scalar Arm sinf polynomial used by the Linux
// reference. Lanczos3 calls it only for finite arguments with |x| < 3*pi.
// Constants and polynomial organization are from Arm optimized-routines,
// commit 20859d3fd9e797dde592c489c8ef7d383f778620. Copyright (c) 2018-2026
// Arm Limited. Used under the MIT license; see LICENSE.resize-math.
// Explicit float64 conversions prohibit fused operations in this path.
func resizeSine32(y float32) float32 {
	x := float64(y)
	n := int(math.Round(float64(x * 0x1.45f306dc9c883p-1)))
	x -= float64(float64(n) * 0x1.921fb54442d18p0)
	z := float64(x * x)
	if n&1 == 0 {
		if n&2 != 0 {
			x = -x
		}
		x3 := float64(x * z)
		s1 := 0x1.1107605230bc4p-7 + float64(z*-0x1.994eb3774cf24p-13)
		x5 := float64(x3 * z)
		s := x + float64(x3*-0x1.555545995a603p-3)
		return float32(s + float64(x5*s1))
	}
	x4 := float64(z * z)
	c2 := -0x1.6c087e89a359dp-10 + float64(z*0x1.99343027bf8c3p-16)
	c1 := 1 + float64(z*-0x1.ffffffd0c621cp-2)
	c := c1 + float64(x4*0x1.55553e1068f19p-5)
	r := c + float64(float64(x4*z)*c2)
	if n&2 != 0 {
		r = -r
	}
	return float32(r)
}

// The Mac reference has one-ulp differences at a finite subset of resize
// arguments. This sorted table contains positive argument/result bit pairs;
// sine's odd symmetry supplies the negative half. It was generated from every
// Lanczos3 argument for every admitted dimension 1..16384, independently of
// image contents. See the v263 image-resize audit for generation, native
// comparisons and per-dimension SHA-256 conformance records. It is math data,
// not malware signatures or a dependency on a native library at runtime.
//
//go:embed resize_math_macos.bin
var resizeMacCorrections string

func resizeMacSine32(x float32) float32 {
	u := math.Float32bits(x)
	key := u & 0x7fffffff
	lo, hi := 0, len(resizeMacCorrections)/8
	for lo < hi {
		mid := int(uint(lo+hi) >> 1)
		v := binary.LittleEndian.Uint32([]byte(resizeMacCorrections[mid*8 : mid*8+4]))
		if v < key {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo < len(resizeMacCorrections)/8 && binary.LittleEndian.Uint32([]byte(resizeMacCorrections[lo*8:lo*8+4])) == key {
		bits := binary.LittleEndian.Uint32([]byte(resizeMacCorrections[lo*8+4 : lo*8+8]))
		return math.Float32frombits(bits ^ (u & 0x80000000))
	}
	return resizeSine32(x)
}
