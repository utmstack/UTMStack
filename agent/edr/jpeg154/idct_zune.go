// This inverse DCT is a Go port of zune-jpeg 0.5.5's scalar idct_int,
// Copyright (c) 2023 the zune-jpeg contributors, used under the MIT license.
// See LICENSE.ZUNE-MIT and NOTICE.md in this package.
package jpeg154

const blockSize = 64

type block [blockSize]int32

// idct converts a dequantized JPEG coefficient block to centered samples.
// The caller adds 128 and clips each result when storing the image plane.
func idct(src *block) {
	allACZero := true
	for i := 1; i < len(src); i++ {
		if src[i] != 0 {
			allACZero = false
			break
		}
	}
	if allACZero {
		v := (src[0] + 4 + 1024) >> 3
		if v < 0 {
			v = 0
		} else if v > 255 {
			v = 255
		}
		for i := range src {
			src[i] = v - 128
		}
		return
	}

	for col := 0; col < 8; col++ {
		p2, p3 := src[col+16], src[col+48]
		p1 := (p2 + p3) * 2217
		t2, t3 := p1-p3*7567, p1+p2*3135
		p2, p3 = src[col], src[col+32]
		t0, t1 := (p2+p3)<<12, (p2-p3)<<12
		x0, x3 := t0+t3+512, t0-t3+512
		x1, x2 := t1+t2+512, t1-t2+512

		t0, t1, t2, t3 = src[col+56], src[col+40], src[col+24], src[col+8]
		p3, p4 := t0+t2, t1+t3
		p1, p2 = t0+t3, t1+t2
		p5 := (p3 + p4) * 4816
		t0 *= 1223
		t1 *= 8410
		t2 *= 12586
		t3 *= 6149
		p1 = p5 - p1*3685
		p2 = p5 - p2*10497
		p3 *= -8034
		p4 *= -1597
		t3 += p1 + p4
		t2 += p2 + p3
		t1 += p2 + p4
		t0 += p1 + p3

		src[col] = (x0 + t3) >> 10
		src[col+8] = (x1 + t2) >> 10
		src[col+16] = (x2 + t1) >> 10
		src[col+24] = (x3 + t0) >> 10
		src[col+32] = (x3 - t0) >> 10
		src[col+40] = (x2 - t1) >> 10
		src[col+48] = (x1 - t2) >> 10
		src[col+56] = (x0 - t3) >> 10
	}

	const scaleBits int32 = 512 + 65536 + (128 << 17)
	for row := 0; row < 8; row++ {
		i := row * 8
		p2, p3 := src[i+2], src[i+6]
		p1 := (p2 + p3) * 2217
		t2, t3 := p1-p3*7567, p1+p2*3135
		p2, p3 = src[i], src[i+4]
		t0, t1 := (p2+p3)<<12, (p2-p3)<<12
		x0, x3 := t0+t3+scaleBits, t0-t3+scaleBits
		x1, x2 := t1+t2+scaleBits, t1-t2+scaleBits

		t0, t1, t2, t3 = src[i+7], src[i+5], src[i+3], src[i+1]
		p3, p4 := t0+t2, t1+t3
		p1, p2 = t0+t3, t1+t2
		p5 := (p3 + p4) * 4816
		t0 *= 1223
		t1 *= 8410
		t2 *= 12586
		t3 *= 6149
		p1 = p5 - p1*3685
		p2 = p5 - p2*10497
		p3 *= -8034
		p4 *= -1597
		t3 += p1 + p4
		t2 += p2 + p3
		t1 += p2 + p4
		t0 += p1 + p3

		src[i] = centeredSample((x0 + t3) >> 17)
		src[i+1] = centeredSample((x1 + t2) >> 17)
		src[i+2] = centeredSample((x2 + t1) >> 17)
		src[i+3] = centeredSample((x3 + t0) >> 17)
		src[i+4] = centeredSample((x3 - t0) >> 17)
		src[i+5] = centeredSample((x2 - t1) >> 17)
		src[i+6] = centeredSample((x1 - t2) >> 17)
		src[i+7] = centeredSample((x0 - t3) >> 17)
	}
}

func centeredSample(v int32) int32 {
	if v < 0 {
		return -128
	}
	if v > 255 {
		return 127
	}
	return v - 128
}
