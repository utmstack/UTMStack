package signatures

import (
	"image"
	"image/color"
)

// webpFancyRGB converts VP8's 4:2:0 planes with the reference's two-row,
// two-column chroma interpolation and integer BT.601 coefficients. The
// interpolation and coefficients follow libwebp's BSD-licensed
// src/dsp/upsampling.c and src/dsp/yuv.h; VP8 entropy decoding remains Go.
func webpFancyRGB(src *image.YCbCr) *image.NRGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	if w == 0 || h == 0 {
		return dst
	}
	mulhi := func(v, c int) int { return v * c >> 8 }
	clip := func(v int) uint8 {
		v >>= 6
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return uint8(v)
	}
	put := func(x, y int, uv uint32) {
		yv := int(src.Y[src.YOffset(b.Min.X+x, b.Min.Y+y)])
		u, v := int(uv&0xff), int(uv>>16)
		yc := mulhi(yv, 19077)
		dst.SetNRGBA(x, y, color.NRGBA{
			R: clip(yc + mulhi(v, 26149) - 14234),
			G: clip(yc - mulhi(u, 6419) - mulhi(v, 13320) + 8708),
			B: clip(yc + mulhi(u, 33050) - 17685), A: 255,
		})
	}
	sample := func(x, y int) uint32 {
		off := src.COffset(b.Min.X+2*x, b.Min.Y+2*y)
		return uint32(src.Cb[off]) | uint32(src.Cr[off])<<16
	}
	pair := func(topY, bottomY, topC, bottomC int) {
		a, c := sample(0, topC), sample(0, bottomC)
		put(0, topY, (3*a+c+0x00020002)>>2)
		if bottomY >= 0 {
			put(0, bottomY, (3*c+a+0x00020002)>>2)
		}
		for x := 1; x <= (w-1)/2; x++ {
			b, d := sample(x, topC), sample(x, bottomC)
			avg := a + b + c + d + 0x00080008
			diag12 := (avg + 2*(b+c)) >> 3
			diag03 := (avg + 2*(a+d)) >> 3
			put(2*x-1, topY, (diag12+a)>>1)
			put(2*x, topY, (diag03+b)>>1)
			if bottomY >= 0 {
				put(2*x-1, bottomY, (diag03+c)>>1)
				put(2*x, bottomY, (diag12+d)>>1)
			}
			a, c = b, d
		}
		if w&1 == 0 {
			put(w-1, topY, (3*a+c+0x00020002)>>2)
			if bottomY >= 0 {
				put(w-1, bottomY, (3*c+a+0x00020002)>>2)
			}
		}
	}
	pair(0, -1, 0, 0)
	for y := 1; y+1 < h; y += 2 {
		pair(y, y+1, y/2, y/2+1)
	}
	if h&1 == 0 {
		pair(h-1, -1, (h-1)/2, (h-1)/2)
	}
	return dst
}
