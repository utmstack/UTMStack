// Package dct32 implements the reference image hash's fixed f32 transform.
// Adapted from RustDCT 0.7.1, Copyright (c) 2015 The RustFFT Developers.
// Distributed under the MIT license in LICENSE.MIT. See NOTICE.md.
package dct32

import "math"

//go:noinline
func product(a, b float32) float32 { return a * b }

type twiddle struct{ c, s float32 }

// Fixed IEEE-754 rotations keep transform setup independent of host libm.
var rotations = [4][8]twiddle{
	{{math.Float32frombits(0x3f6c835e), math.Float32frombits(0x3ec3ef15)}},
	{{math.Float32frombits(0x3f7b14be), math.Float32frombits(0x3e47c5c2)}, {math.Float32frombits(0x3f54db31), math.Float32frombits(0x3f0e39da)}},
	{{math.Float32frombits(0x3f7ec46d), math.Float32frombits(0x3dc8bd36)}, {math.Float32frombits(0x3f74fa0b), math.Float32frombits(0x3e94a031)}, {math.Float32frombits(0x3f61c598), math.Float32frombits(0x3ef15aea)}, {math.Float32frombits(0x3f45e403), math.Float32frombits(0x3f226799)}},
	{{math.Float32frombits(0x3f7fb10f), math.Float32frombits(0x3d48fb30)}, {math.Float32frombits(0x3f7d3aac), math.Float32frombits(0x3e164083)}, {math.Float32frombits(0x3f7853f8), math.Float32frombits(0x3e78cfcc)}, {math.Float32frombits(0x3f710908), math.Float32frombits(0x3eac7cd4)}, {math.Float32frombits(0x3f676bd8), math.Float32frombits(0x3edae880)}, {math.Float32frombits(0x3f5b941a), math.Float32frombits(0x3f039c3d)}, {math.Float32frombits(0x3f4d9f02), math.Float32frombits(0x3f187fc0)}, {math.Float32frombits(0x3f3daef9), math.Float32frombits(0x3f2beb4a)}},
}

const rootHalf = float32(0.707106781186547524400844362104849039)

func two(b []float32) {
	sum := b[0] + b[1]
	b[1] = product(b[0]-b[1], rootHalf)
	b[0] = sum
}

func sineTwo(b []float32) {
	sum := b[0] - b[1]
	b[0] = product(b[0]+b[1], rootHalf)
	b[1] = sum
}

func four(b []float32) {
	lower, upper := b[0]-b[3], b[2]-b[1]
	zero, two := b[0]+b[3], b[2]+b[1]
	b[0], b[2] = zero+two, product(zero-two, rootHalf)
	t := rotations[0][0]
	b[1] = product(lower, t.c) - product(upper, t.s)
	b[3] = product(upper, t.c) + product(lower, t.s)
}

func sineFour(b []float32) {
	lower, upper := b[0]+b[3], b[2]+b[1]
	three, one := b[0]-b[3], b[2]-b[1]
	b[3], b[1] = three+one, product(three-one, rootHalf)
	t := rotations[0][0]
	b[2] = product(lower, t.c) - product(upper, t.s)
	b[0] = product(upper, t.c) + product(lower, t.s)
}

func eight(b []float32) {
	even := [4]float32{b[0] + b[7], b[1] + b[6], b[2] + b[5], b[3] + b[4]}
	four(even[:])
	d := [4]float32{b[0] - b[7], b[3] - b[4], b[1] - b[6], b[2] - b[5]}
	t0, t1 := rotations[1][0], rotations[1][1]
	cos := [2]float32{product(d[0], t0.c) + product(d[1], t0.s), product(d[2], t1.c) + product(d[3], t1.s)}
	sin := [2]float32{product(d[3], t1.c) - product(d[2], t1.s), product(d[1], t0.c) - product(d[0], t0.s)}
	two(cos[:])
	sineTwo(sin[:])
	b[0], b[1], b[2], b[3] = even[0], cos[0], even[1], cos[1]-sin[0]
	b[4], b[5], b[6], b[7] = even[2], cos[1]+sin[0], even[3], sin[1]
}

func sixteen(b []float32) {
	var even [8]float32
	for i := range even {
		even[i] = b[i] + b[15-i]
	}
	eight(even[:])
	var cos, sin [4]float32
	for i := 0; i < 4; i++ {
		lower, upper := b[i]-b[15-i], b[7-i]-b[8+i]
		t := rotations[2][i]
		cos[i] = product(lower, t.c) + product(upper, t.s)
		sin[3-i] = product(upper, t.c) - product(lower, t.s)
	}
	four(cos[:])
	sineFour(sin[:])
	b[0], b[1], b[2], b[3] = even[0], cos[0], even[1], cos[1]-sin[0]
	b[4], b[5], b[6], b[7] = even[2], cos[1]+sin[0], even[3], cos[2]+sin[1]
	b[8], b[9], b[10], b[11] = even[4], cos[2]-sin[1], even[5], cos[3]-sin[2]
	b[12], b[13], b[14], b[15] = even[6], cos[3]+sin[2], even[7], sin[3]
}

func thirtyTwo(b *[32]float32) {
	var even [16]float32
	var cos, sin [8]float32
	for i := 0; i < 8; i++ {
		bottom, top, halfBottom, halfTop := b[i], b[31-i], b[15-i], b[16+i]
		even[i], even[15-i] = top+bottom, halfBottom+halfTop
		lower, upper := bottom-top, halfBottom-halfTop
		t := rotations[3][i]
		cos[i] = product(lower, t.c) + product(upper, t.s)
		v := product(upper, t.c) - product(lower, t.s)
		if i&1 != 0 {
			v = -v
		}
		sin[7-i] = v
	}
	sixteen(even[:])
	eight(cos[:])
	eight(sin[:])
	b[0], b[1], b[2] = even[0], cos[0], even[1]
	for i := 1; i < 8; i++ {
		v := sin[8-i]
		if i&1 == 0 {
			v = -v
		}
		b[4*i-1], b[4*i], b[4*i+1], b[4*i+2] = cos[i]+v, even[2*i], cos[i]-v, even[2*i+1]
	}
	b[31] = -sin[0]
}

// Coefficients returns the top-left 8x8 of the normalized, column-first
// transform. Products are rounded before addition, and both passes double
// their output, matching the reference's fixed transform plan.
func Coefficients(pixels *[32][32]uint8) (out [64]float32) {
	var intermediate [32][32]float32
	for x := 0; x < 32; x++ {
		var column [32]float32
		for y := range column {
			column[y] = float32(pixels[y][x]) / 255
		}
		thirtyTwo(&column)
		for y, v := range column {
			intermediate[y][x] = v * 2
		}
	}
	for y := 0; y < 8; y++ {
		thirtyTwo(&intermediate[y])
		for x := 0; x < 8; x++ {
			out[y*8+x] = intermediate[y][x] * 2
		}
	}
	return
}
