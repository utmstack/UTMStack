package signatures

import "math"

// iconLabDistance is the distance of an RGB pixel from the reference's
// mid-gray Lab point. The coefficients are the float constants in ClamAV
// 1.5.4's pe_icons.c, promoted to double for the calculation.
func iconLabDistance(p iconPixel) float64 {
	r, g, b := float64(p[0])/255, float64(p[1])/255, float64(p[2])/255
	linear := func(c float64) float64 {
		if c > float64(float32(0.04045)) {
			return math.Pow((c+float64(float32(0.055)))/float64(float32(1.055)), float64(float32(2.4)))
		}
		return c / float64(float32(12.92))
	}
	r, g, b = linear(r)*100, linear(g)*100, linear(b)*100
	x := (r*float64(float32(0.4124)) + g*float64(float32(0.3576)) + b*float64(float32(0.1805))) / float64(float32(95.047))
	y := (r*float64(float32(0.2126)) + g*float64(float32(0.7152)) + b*float64(float32(0.0722))) / 100
	z := (r*float64(float32(0.0193)) + g*float64(float32(0.1192)) + b*float64(float32(0.9505))) / float64(float32(108.883))
	f := func(v float64) float64 {
		if v > float64(float32(0.008856)) {
			return math.Pow(v, float64(float32(1.0/3.0)))
		}
		return float64(float32(7.787))*v + float64(float32(16.0/116.0))
	}
	x, y, z = f(x), f(y), f(z)
	l := float64(float32(116))*y - float64(float32(16))
	a := float64(float32(500)) * (x - y)
	bb := float64(float32(200)) * (y - z)
	const l0 = float64(float32(53.192777691077211))
	const a0 = float64(float32(0.0031420942181448197))
	const b0 = float64(float32(-0.0062075877844014471))
	return math.Sqrt((l0-l)*(l0-l) + (a0-a)*(a0-a) + (b0-bb)*(b0-bb))
}

// iconEdgeMap applies Sobel magnitude, max normalization and a separable
// 1:2:1 Gaussian blur to the scaled icon. Border pixels remain black.
func iconEdgeMap(px []iconPixel, w int) []int {
	lab := make([]float64, len(px))
	for i, p := range px {
		lab[i] = iconLabDistance(p)
	}
	magnitude := make([]int, len(px))
	peak := 0
	for y := 1; y < w-1; y++ {
		for x := 1; x < w-1; x++ {
			a, b, c := lab[(y-1)*w+x-1], lab[(y-1)*w+x], lab[(y-1)*w+x+1]
			d, f := lab[y*w+x-1], lab[y*w+x+1]
			g, h, i := lab[(y+1)*w+x-1], lab[(y+1)*w+x], lab[(y+1)*w+x+1]
			gx := a + 2*d + g - c - 2*f - i
			gy := a + 2*b + c - g - 2*h - i
			v := int(math.Sqrt(gx*gx + gy*gy))
			magnitude[y*w+x] = v
			peak = max(peak, v)
		}
	}
	normalized := make([]int, len(px))
	if peak > 0 {
		for y := 1; y < w-1; y++ {
			for x := 1; x < w-1; x++ {
				normalized[y*w+x] = magnitude[y*w+x] * 255 / peak
			}
		}
	} else {
		// The reference skips normalization when every Sobel magnitude is
		// zero, leaving the original blue channel for the blur pass.
		for y := 1; y < w-1; y++ {
			for x := 1; x < w-1; x++ {
				normalized[y*w+x] = px[y*w+x][2]
			}
		}
	}
	horizontal := make([]int, len(px))
	for y := 1; y < w-1; y++ {
		for x := 1; x < w-1; x++ {
			horizontal[y*w+x] = (normalized[y*w+x-1] + 2*normalized[y*w+x] + normalized[y*w+x+1]) / 4
		}
	}
	// The vertical pass updates the same image it reads. Its preceding row
	// therefore already contains the vertical result, while the next row
	// still contains the horizontal result.
	blurred := horizontal
	for y := 1; y < w-1; y++ {
		for x := 1; x < w-1; x++ {
			blurred[y*w+x] = (horizontal[(y-1)*w+x] + 2*horizontal[y*w+x] + horizontal[(y+1)*w+x]) / 4
		}
	}
	return blurred
}
