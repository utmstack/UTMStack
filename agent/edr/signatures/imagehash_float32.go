package signatures

import "math"

// The pinned Rust build rounds every weighted product to f32 before adding.
// A fused multiply-add changes some output pixels close to .5.
//
//go:noinline
func mul32(a, b float32) float32 { return a * b }

// lanczos3f and resample32 reproduce the f32 filter path used for the
// reference's 32x32 image reduction. The vertical pass leaves pixels in f32;
// the horizontal pass is rounded to 8-bit only after both axes are sampled.
func lanczos3f(x float32) float32 {
	return lanczos3Profile(x, resizeMacSine32)
}

func lanczos3Profile(x float32, sine func(float32) float32) float32 {
	if x == 0 {
		return 1
	}
	if x <= -3 || x >= 3 {
		return 0
	}
	sinc := func(v float32) float32 {
		a := v * float32(math.Pi)
		return sine(a) / a
	}
	return mul32(sinc(x), sinc(x/3))
}

func resample32(src [][]float32, n int, vertical bool) [][]float32 {
	return resample32Profile(src, n, vertical, resizeMacSine32)
}

func resample32Profile(src [][]float32, n int, vertical bool, sine func(float32) float32) [][]float32 {
	h, w := len(src), len(src[0])
	const target = 32
	var out [][]float32
	if vertical {
		out = make([][]float32, target)
		for i := range out {
			out[i] = make([]float32, w)
		}
	} else {
		out = make([][]float32, h)
		for i := range out {
			out[i] = make([]float32, target)
		}
	}
	weights := make([]float32, 0, int(6*max(float32(n)/target, 1))+2)
	for o := 0; o < target; o++ {
		var left int
		left, weights = resizeWeights32(n, o, sine, weights[:0])
		right := left + len(weights)
		if vertical {
			for k, i := 0, left; i < right; i, k = i+1, k+1 {
				for x := range out[o] {
					out[o][x] += mul32(src[i][x], weights[k])
				}
			}
		} else {
			for y := 0; y < h; y++ {
				value := float32(0)
				for k, i := 0, left; i < right; i, k = i+1, k+1 {
					value += mul32(src[y][i], weights[k])
				}
				out[y][o] = value
			}
		}
	}
	return out
}

// Explicit f32 products preserve the native filter's rounding before bounds,
// normalization and accumulation. An assignment's type alone permits fusion.
func resizeWeights32(n, o int, sine func(float32) float32, weights []float32) (int, []float32) {
	ratio := float32(n) / 32
	scale := max(ratio, 1)
	support := float32(float32(3) * scale)
	center := float32((float32(o) + 0.5) * ratio)
	left := min(max(int(math.Floor(float64(center-support))), 0), n-1)
	right := min(max(int(math.Ceil(float64(center+support))), left+1), n)
	center -= 0.5
	sum := float32(0)
	for i := left; i < right; i++ {
		weight := lanczos3Profile((float32(i)-center)/scale, sine)
		weights = append(weights, weight)
		sum += weight
	}
	for i := range weights {
		weights[i] /= sum
	}
	return left, weights
}
