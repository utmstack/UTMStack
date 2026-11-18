package signatures

import (
	"encoding/hex"
	"errors"
	"math"
)

// The reference's icon hash, checked against its debug output ("IDB
// SIGNATURE"), custom icon entries and ClamAV 1.5.4's public pe_icons.c.
// See docs/edr/native-engine/parity/icons.md for evidence and limits.
//
// An icon is composited onto white, scaled to a hash width of 16, 24 or 32
// (also 18, 21, 27 or 30 after fast scaling), and described by areas: square
// windows of a quarter of the width. Each measure keeps three windows: the
// three most and least colorful, brightest and darkest, and with the most
// and fewest edges. The color spread says how much of the icon is colored
// and in which hues.

// iconArea is one area of an icon hash: a window's mean value and its
// top-left corner.
type iconArea struct{ v, x, y int }

// iconHash is one icon's hash: its width, the color and gray areas (12-bit
// values), the bright, dark, edge and noedge areas (8-bit values) and the
// four color spread figures.
type iconHash struct {
	width                                   int
	color, gray, bright, dark, edge, noedge [3]iconArea
	spread                                  [4]int
}

// Limits of the icon database loader: a 12-bit value above 4072, a corner
// past seven eighths of the width, hue figures summing to more than 103, or
// a colored share above 100 is a malformed hash. 8-bit values are not
// checked.
const (
	iconMaxValue12 = 4072
	iconMaxSpread  = 103
	iconMaxColored = 100
)

// parseIconHash reads the 124 hexadecimal digits of an icon entry, rejecting
// what the reference's loader rejects.
func parseIconHash(s string) (iconHash, error) {
	h, err := decodeIconHash(s)
	if err != nil {
		return h, err
	}
	if h.width != 16 && h.width != 24 && h.width != 32 {
		return h, errors.New("invalid icon signature: bad size")
	}
	maxCorner := h.width - h.width/8
	for _, a := range h.color[:] {
		if a.v > iconMaxValue12 || a.x > maxCorner || a.y > maxCorner {
			return h, errors.New("invalid icon signature: bad color data")
		}
	}
	for _, a := range h.gray[:] {
		if a.v > iconMaxValue12 || a.x > maxCorner || a.y > maxCorner {
			return h, errors.New("invalid icon signature: bad color data")
		}
	}
	for _, areas := range [][3]iconArea{h.bright, h.dark, h.edge, h.noedge} {
		for _, a := range areas {
			if a.x > maxCorner || a.y > maxCorner {
				return h, errors.New("invalid icon signature: bad area data")
			}
		}
	}
	if h.spread[0]+h.spread[1]+h.spread[2] > iconMaxSpread || h.spread[3] > iconMaxColored {
		return h, errors.New("invalid icon signature: bad spread data")
	}
	return h, nil
}

// decodeIconHash reads an icon hash's fields without checking them.
func decodeIconHash(s string) (iconHash, error) {
	var h iconHash
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 62 {
		return h, errors.New("invalid icon signature fields")
	}
	h.width = int(b[0])
	nibble := func(i int) int { // the i-th hexadecimal digit after the width
		c := b[1+i/2]
		if i%2 == 0 {
			return int(c >> 4)
		}
		return int(c & 15)
	}
	for i := 0; i < 6; i++ {
		n := 7 * i
		a := iconArea{
			v: nibble(n)<<8 | nibble(n+1)<<4 | nibble(n+2),
			x: nibble(n+3)<<4 | nibble(n+4),
			y: nibble(n+5)<<4 | nibble(n+6),
		}
		if i < 3 {
			h.color[i] = a
		} else {
			h.gray[i-3] = a
		}
	}
	p := 22
	for _, dst := range []*[3]iconArea{&h.bright, &h.dark, &h.edge, &h.noedge} {
		for i := range dst {
			dst[i] = iconArea{int(b[p]), int(b[p+1]), int(b[p+2])}
			p += 3
		}
	}
	for i := range h.spread {
		h.spread[i] = int(b[p+i])
	}
	return h, nil
}

// iconPixel is a pixel composited onto white.
type iconPixel [3]int

// iconScaleRect follows the reference's slow scaler for an icon whose
// bitmap width and height differ. The chosen output size minimizes the
// Manhattan distance to 16, 24 or 32, with the reference's tie order.
func iconScaleRect(px []iconPixel, width, height int) ([]iconPixel, int) {
	distance := func(size int) int {
		return int(math.Abs(float64(width-size))) + int(math.Abs(float64(height-size)))
	}
	size := 16
	if distance(32) < distance(24) {
		size = 32
	} else if distance(24) < distance(16) {
		size = 24
	}
	scaleX := float64(width) / float64(size)
	scaleY := float64(height) / float64(size)
	out := make([]iconPixel, size*size)
	for y := 0; y < size; y++ {
		sy := int(float64(y) * scaleY)
		for x := 0; x < size; x++ {
			sx := int(float64(x)*scaleX + 0.5)
			out[y*size+x] = px[sy*width+sx]
		}
	}
	return out, size
}

// iconScale brings composited pixels of a w x w icon to its hash width. A
// width of 16, 24 or 32 is kept; a multiple of 24 or 32 is halved while it is
// above 32, each pixel the mean of its horizontal pairs' means (all integer
// divisions); any other width is resampled to the nearest of 16, 24 and 32
// (ties to the smaller, above 32 to 32) by taking, for each target pixel, the
// source column round(x*w/t) and row floor(y*w/t).
func iconScale(px []iconPixel, w int) ([]iconPixel, int) {
	switch {
	case w == 16 || w == 24 || w == 32:
	case w%24 == 0 || w%32 == 0:
		for w > 32 {
			nw := w / 2
			out := make([]iconPixel, nw*nw)
			for y := 0; y < nw; y++ {
				for x := 0; x < nw; x++ {
					p00, p01 := px[2*y*w+2*x], px[2*y*w+2*x+1]
					p10, p11 := px[(2*y+1)*w+2*x], px[(2*y+1)*w+2*x+1]
					for c := 0; c < 3; c++ {
						out[y*nw+x][c] = ((p00[c]+p01[c])/2 + (p10[c]+p11[c])/2) / 2
					}
				}
			}
			px, w = out, nw
		}
	default:
		t := 32
		switch {
		case w <= 20:
			t = 16
		case w <= 28:
			t = 24
		}
		s := float64(w) / float64(t)
		out := make([]iconPixel, t*t)
		for y := 0; y < t; y++ {
			sy := min(int(math.Floor(float64(y)*s)), w-1)
			for x := 0; x < t; x++ {
				sx := min(int(math.Round(float64(x)*s)), w-1)
				out[y*t+x] = px[sy*w+sx]
			}
		}
		px, w = out, t
	}
	return px, w
}

// iconMeasures computes the reference icon hash from its scaled pixels.
// Grayscale icons store a second set of edge and noedge areas in the color
// and gray fields, respectively.
func iconMeasures(px []iconPixel, w int) (h iconHash, grayMode bool) {
	h.width = w
	value := make([]int, len(px))
	color := make([]int, len(px))
	for i, p := range px {
		v := max(p[0], p[1], p[2])
		value[i] = v
		if v > 0 {
			s := 255 * (v - min(p[0], p[1], p[2])) / v
			color[i] = int(math.Sqrt(float64(v)) * float64(s))
		}
	}
	h.bright = iconAreas(value, w, true)
	h.dark = iconAreas(value, w, false)
	h.spread = iconSpread(px)
	edges := iconEdgeMap(px, w)
	if h.spread == [4]int{} {
		high, low := iconAreasN(edges, w, true, 6), iconAreasN(edges, w, false, 6)
		copy(h.edge[:], high[:3])
		copy(h.color[:], high[3:])
		copy(h.noedge[:], low[:3])
		copy(h.gray[:], low[3:])
		// The reference stores the second noedge values at the second
		// edge set's coordinates.
		for i := range h.gray {
			h.gray[i].x, h.gray[i].y = high[i+3].x, high[i+3].y
		}
		return h, true
	}
	h.edge = iconAreas(edges, w, true)
	h.noedge = iconAreas(edges, w, false)
	h.color = iconAreas(color, w, true)
	h.gray = iconAreas(color, w, false)
	return h, false
}

// iconAreas picks three windows of a quarter of the width, greedily, each
// not overlapping the ones already picked: the largest (high) or smallest
// window sum, the first in row order on ties. Corners run over x up to
// w-k-2 and y up to w-k-1. A high search starts from an empty area at (0,0)
// that only a positive sum replaces. Values are the windows' integer means.
func iconAreas(vals []int, w int, high bool) [3]iconArea {
	a := iconAreasN(vals, w, high, 3)
	return [3]iconArea{a[0], a[1], a[2]}
}

func iconAreasN(vals []int, w int, high bool, count int) []iconArea {
	k := w / 4
	out := make([]iconArea, count)
	for n := 0; n < count; n++ {
		found := high
		var best iconArea
		bestSum := 0
		for y := 0; y <= w-k-1; y++ {
			for x := 0; x <= w-k-2; x++ {
				clash := false
				for _, c := range out[:n] {
					if abs(c.x-x) < k && abs(c.y-y) < k {
						clash = true
						break
					}
				}
				if clash {
					continue
				}
				sum := 0
				for j := y; j < y+k; j++ {
					for _, v := range vals[j*w+x : j*w+x+k] {
						sum += v
					}
				}
				if !found || high && sum > bestSum || !high && sum < bestSum {
					best, bestSum, found = iconArea{sum / (k * k), x, y}, sum, true
				}
			}
		}
		out[n] = best
	}
	return out
}

// iconSpread is the color spread. A pixel is colored when its saturation
// 255*(max-min)/max and its max are both at least 86. Its chroma is split
// into three hue parts by the smallest channel: with the blue channel
// smallest, the parts are red and green beyond their common yellow, and the
// yellow; with red smallest, the cyan and the green and blue beyond it; with
// green smallest, the red beyond the magenta, the magenta and the blue
// beyond it. Each part is ceil(100*part/(max-min)). The first three figures
// are the parts' sums over the colored pixels divided by their number, the
// fourth the colored share in percent (floored); with no colored pixel or a
// share of 5 or less the spread is zero.
func iconSpread(px []iconPixel) [4]int {
	var sums [3]int
	colored := 0
	for _, p := range px {
		hi, lo := max(p[0], p[1], p[2]), min(p[0], p[1], p[2])
		if hi < 86 || 255*(hi-lo)/hi < 86 {
			continue
		}
		colored++
		r, g, b, c := p[0]-lo, p[1]-lo, p[2]-lo, hi-lo
		var parts [3]int
		switch lo {
		case p[2]:
			y := min(r, g)
			parts = [3]int{r - y, g - y, y}
		case p[0]:
			cy := min(g, b)
			parts = [3]int{cy, g - cy, b - cy}
		default:
			m := min(r, b)
			parts = [3]int{r - m, m, b - m}
		}
		for i, part := range parts {
			sums[i] += (100*part + c - 1) / c
		}
	}
	if colored == 0 {
		return [4]int{}
	}
	share := colored * 100 / len(px)
	if share <= 5 {
		return [4]int{}
	}
	return [4]int{sums[0] / colored, sums[1] / colored, sums[2] / colored, share}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
