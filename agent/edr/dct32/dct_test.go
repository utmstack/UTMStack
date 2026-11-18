package dct32

import (
	"math"
	"testing"
)

// Compare the optimized fixed transform with its independent mathematical
// definition. Exact reference hash bits are separately checked on image
// corpora, including transforms with coefficients on the median.
func TestTransformDefinition(t *testing.T) {
	state := uint32(0x11235813)
	for trial := 0; trial < 64; trial++ {
		var input [32]float32
		for i := range input {
			state = state*1664525 + 1013904223
			input[i] = float32(state&65535)/65535 - 0.5
		}
		got := input
		thirtyTwo(&got)
		for k, actual := range got {
			want := 0.0
			for x, v := range input {
				want += float64(v) * math.Cos(math.Pi/32*(float64(x)+0.5)*float64(k))
			}
			if math.Abs(float64(actual)-want) > 3e-6 {
				t.Fatalf("trial %d coefficient %d: got %.9g want %.9g", trial, k, actual, want)
			}
		}
	}
}

func TestCoefficientsAxes(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		var pixels [32][32]uint8
		for y := range pixels {
			for x := range pixels[y] {
				v := x
				if vertical {
					v = y
				}
				pixels[y][x] = uint8(v * 7)
			}
		}
		got := Coefficients(&pixels)
		for ky := 0; ky < 8; ky++ {
			for kx := 0; kx < 8; kx++ {
				want := 0.0
				for y := range pixels {
					for x, v := range pixels[y] {
						want += float64(float32(v)/255) * 4 *
							math.Cos(math.Pi/32*(float64(x)+0.5)*float64(kx)) *
							math.Cos(math.Pi/32*(float64(y)+0.5)*float64(ky))
					}
				}
				if math.Abs(float64(got[ky*8+kx])-want) > 0.002 {
					t.Fatalf("vertical %t (%d,%d): got %.9g want %.9g", vertical, kx, ky, got[ky*8+kx], want)
				}
			}
		}
	}
}
