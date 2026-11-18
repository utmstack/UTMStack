package signatures

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"testing"
)

func TestResizeSineProfiles(t *testing.T) {
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(resizeMacCorrections))); got != "73fc133df744f8a607348a319ce61857c808654bc6cf1f0a19b5ec515b7525b3" {
		t.Fatalf("resize math table changed: %s", got)
	}
	var previous uint32
	for at := 0; at < len(resizeMacCorrections); at += 8 {
		key := binary.LittleEndian.Uint32([]byte(resizeMacCorrections[at : at+4]))
		value := binary.LittleEndian.Uint32([]byte(resizeMacCorrections[at+4 : at+8]))
		if key <= previous || key >= math.Float32bits(3*float32(math.Pi)) {
			t.Fatalf("unsorted or out-of-domain correction %08x", key)
		}
		previous = key
		arm := math.Float32bits(resizeSine32(math.Float32frombits(key)))
		if arm+1 != value && value+1 != arm {
			t.Fatalf("correction is not one ulp: %08x %08x %08x", key, arm, value)
		}
		for _, sign := range []uint32{0, 1 << 31} {
			if got := math.Float32bits(resizeMacSine32(math.Float32frombits(key | sign))); got != value^sign {
				t.Fatalf("correction lookup: %08x -> %08x, want %08x", key|sign, got, value^sign)
			}
		}
	}
	for _, row := range [][3]uint32{
		{0x3e16cbe4, 0x3e164083, 0x3e164084},
		{0x3ea78d36, 0x3ea493b7, 0x3ea493b7},
		{0x3f0e6b3b, 0x3f072f74, 0x3f072f74},
		{0x40e55616, 0x3f45e407, 0x3f45e407},
		{0x411539c5, 0x3dc8bcd9, 0x3dc8bcd9},
	} {
		x := math.Float32frombits(row[0])
		if a, m := math.Float32bits(resizeSine32(x)), math.Float32bits(resizeMacSine32(x)); a != row[1] || m != row[2] {
			t.Fatalf("native sine vector %08x: Linux %08x Mac %08x", row[0], a, m)
		}
	}
}

// The independent C/libm oracle writes a SHA-256 of (argument bits, sine bits)
// for every dimension, avoiding a multi-gigabyte checked-in golden stream.
// This optional exhaustive qualification checks 1,535,181,860 arguments per
// profile. Ordinary tests check the complete correction table and regressions.
func TestResizeSineExhaustive(t *testing.T) {
	paths := []string{os.Getenv("EDR_RESIZE_ARM_GRID"), os.Getenv("EDR_RESIZE_MAC_GRID")}
	if paths[0] == "" || paths[1] == "" {
		t.Skip("set both EDR_RESIZE_ARM_GRID and EDR_RESIZE_MAC_GRID to v263 gzip oracle records")
	}
	var gold [2][]byte
	for i, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		z, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			t.Fatal(err)
		}
		gold[i], err = io.ReadAll(io.LimitReader(z, 16384*40+1))
		closeErr := z.Close()
		f.Close()
		if err != nil || closeErr != nil || len(gold[i]) != 16384*40 {
			t.Fatalf("oracle record length/checksum: %d %v %v", len(gold[i]), err, closeErr)
		}
	}
	var total uint64
	for n := 1; n <= imageHashMaxDimension; n++ {
		streams := [2][]byte{make([]byte, 0, n*128+4096), make([]byte, 0, n*128+4096)}
		for o := 0; o < 32; o++ {
			ratio := float32(n) / 32
			scale := max(ratio, 1)
			support := float32(3 * scale)
			center := float32((float32(o) + 0.5) * ratio)
			left := min(max(int(math.Floor(float64(center-support))), 0), n-1)
			right := min(max(int(math.Ceil(float64(center+support))), left+1), n)
			center -= 0.5
			for i := left; i < right; i++ {
				x := (float32(i) - center) / scale
				if x == 0 || x <= -3 || x >= 3 {
					continue
				}
				for _, a := range []float32{x * float32(math.Pi), (x / 3) * float32(math.Pi)} {
					values := [2]float32{resizeSine32(a), resizeMacSine32(a)}
					for p, value := range values {
						streams[p] = binary.LittleEndian.AppendUint32(streams[p], math.Float32bits(a))
						streams[p] = binary.LittleEndian.AppendUint32(streams[p], math.Float32bits(value))
					}
					total++
				}
			}
		}
		for p, stream := range streams {
			record := gold[p][(n-1)*40 : n*40]
			sum := sha256.Sum256(stream)
			if int(binary.LittleEndian.Uint32(record)) != n || int(binary.LittleEndian.Uint32(record[4:])) != len(stream)/8 || string(sum[:]) != string(record[8:]) {
				t.Fatalf("native sine mismatch profile=%d dimension=%d count=%d sha256=%x", p, n, len(stream)/8, sum)
			}
		}
		if n%1024 == 0 {
			t.Logf("dimensions=%d evaluations_per_profile=%d", n, total)
		}
	}
	if total != 1535181860 {
		t.Fatalf("incomplete native grid: %d", total)
	}
}

func TestResizeWeightsExhaustive(t *testing.T) {
	paths := []string{os.Getenv("EDR_RESIZE_MAC_WEIGHTS"), os.Getenv("EDR_RESIZE_ARM_WEIGHTS")}
	if paths[0] == "" || paths[1] == "" {
		t.Skip("set both EDR_RESIZE_MAC_WEIGHTS and EDR_RESIZE_ARM_WEIGHTS to v263 gzip weight records")
	}
	var gold [2][]byte
	for p, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		z, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			t.Fatal(err)
		}
		gold[p], err = io.ReadAll(io.LimitReader(z, 16384*40+1))
		z.Close()
		f.Close()
		if err != nil || len(gold[p]) != 16384*40 {
			t.Fatalf("weight oracle: length=%d error=%v", len(gold[p]), err)
		}
	}
	var total uint64
	for n := 1; n <= imageHashMaxDimension; n++ {
		for p, sine := range []func(float32) float32{resizeMacSine32, resizeSine32} {
			stream := make([]byte, 0, n*64+4096)
			weights := make([]float32, 0, n)
			var count uint32
			for o := 0; o < 32; o++ {
				left, row := resizeWeights32(n, o, sine, weights[:0])
				weights = row
				stream = binary.LittleEndian.AppendUint32(stream, uint32(left))
				stream = binary.LittleEndian.AppendUint32(stream, uint32(left+len(row)))
				for _, weight := range row {
					stream = binary.LittleEndian.AppendUint32(stream, math.Float32bits(weight))
				}
				count += uint32(len(row))
			}
			record := gold[p][(n-1)*40 : n*40]
			sum := sha256.Sum256(stream)
			if int(binary.LittleEndian.Uint32(record)) != n || binary.LittleEndian.Uint32(record[4:]) != count || string(sum[:]) != string(record[8:]) {
				t.Fatalf("native filter mismatch profile=%d dimension=%d weights=%d sha256=%x", p, n, count, sum)
			}
			if p == 0 {
				total += uint64(count)
			}
		}
		if n%1024 == 0 {
			t.Logf("dimensions=%d normalized_weights_per_profile=%d", n, total)
		}
	}
}
