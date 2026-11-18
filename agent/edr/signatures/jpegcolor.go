package signatures

import "image"

// The interpolation and color coefficients below follow zune-jpeg 0.5.5
// (zune-image developers), used under its MIT license. See
// ../jpeg154/LICENSE.ZUNE-MIT and ../jpeg154/NOTICE.md.

// referenceJPEGImage retains Go's decoded JPEG component planes, but uses the
// chroma interpolation and fixed-point color conversion of the image decoder
// used by the pinned ClamAV release. Floating-point hash-stage boundaries and
// unsupported JPEG layouts can still require an uncertain rule decision.
type referenceJPEGImage struct{ *image.YCbCr }

// referenceCMYKImage uses the reference's Adobe CMYK-to-RGB multiplication
// after the decoder has interpolated and inverted the four components.
type referenceCMYKImage struct{ *image.CMYK }

func referenceCMYKProduct(a, b uint8) uint8 {
	t := int(a)*int(b) + 128
	return uint8((t + (t >> 8)) >> 8)
}

// referenceJPEGChroma uses the JPEG decoder's 3:1 neighbor interpolation when
// chroma is subsampled by two. The vertical pass precedes the horizontal pass.
// The visible edge is replicated; malformed and unusual layouts stay covered
// by the image hash's conservative uncertainty handling.
func referenceJPEGChroma(src *image.YCbCr, x, y int, cb bool) uint8 {
	b := src.Bounds()
	hFactor, vFactor := 1, 1
	switch src.SubsampleRatio {
	case image.YCbCrSubsampleRatio422:
		hFactor = 2
	case image.YCbCrSubsampleRatio420:
		hFactor, vFactor = 2, 2
	case image.YCbCrSubsampleRatio440:
		vFactor = 2
	case image.YCbCrSubsampleRatio411:
		hFactor = 4
	case image.YCbCrSubsampleRatio410:
		hFactor, vFactor = 4, 2
	}
	// The decoder keeps full MCU planes behind the cropped visible bounds.
	// The reference upsamples these planes before cropping the output image.
	maxX := src.CStride*hFactor - 1
	maxY := len(src.Cb)/src.CStride*vFactor - 1
	sample := func(x, y int) int {
		if x < b.Min.X {
			x = b.Min.X
		} else if x > maxX {
			x = maxX
		}
		if y < b.Min.Y {
			y = b.Min.Y
		} else if y > maxY {
			y = maxY
		}
		i := src.COffset(x, y)
		if cb {
			return int(src.Cb[i])
		}
		return int(src.Cr[i])
	}
	v := sample(x, y)
	horizontal, vertical := false, false
	switch src.SubsampleRatio {
	case image.YCbCrSubsampleRatio422:
		horizontal = true
	case image.YCbCrSubsampleRatio420:
		horizontal, vertical = true, true
	case image.YCbCrSubsampleRatio440:
		vertical = true
	}
	neighbor := func(n int) int {
		if n&1 == 0 {
			return n - 1
		}
		return n + 1
	}
	if vertical {
		v = (3*v + sample(x, neighbor(y)) + 2) >> 2
	}
	if horizontal {
		other := sample(neighbor(x), y)
		if vertical {
			other = (3*other + sample(neighbor(x), neighbor(y)) + 2) >> 2
		}
		v = (3*v + other + 2) >> 2
	}
	return uint8(v)
}

// referenceJPEGToRGB uses the reference decoder's 14-bit BT.601 full-range
// coefficients and its 0.4999 rounding offset.
func referenceJPEGToRGB(y, cb, cr uint8) (uint8, uint8, uint8) {
	cbi, cri := int(cb)-128, int(cr)-128
	y0 := int(y)*16384 + 8191
	clamp := func(v int) uint8 {
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return uint8(v)
	}
	return clamp((y0 + cri*22970) >> 14),
		clamp((y0 - cri*11700 - cbi*5638) >> 14),
		clamp((y0 + cbi*29032) >> 14)
}
