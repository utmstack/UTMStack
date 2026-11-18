package bytecode

import (
	"bytes"
	"context"
	"math/rand"
	"testing"
	"time"
)

// apiWork calls an API and returns the work it charged.
func apiWork(t *testing.T, state *apiState, mem *memory, id int, args ...uint64) uint64 {
	t.Helper()
	var total uint64
	if _, err := state.invoke(context.Background(), id, args, mem, func(n uint64) error { total += n; return nil }); err != nil {
		t.Fatalf("API %d: %v", id, err)
	}
	return total
}

func TestAPIByteWorkIsChargedPer64Bytes(t *testing.T) {
	// A runtime call is charged one step plus one per 64 bytes it moves or
	// scans, as the interpreter's own bulk memory operations are.
	const size = 1 << 20
	const bound = 2 + size/64 + size/4096 + 64
	mem := &memory{}
	needle := apiTestBytes(t, mem, []byte("NEEDLE"))
	buffer := apiTestBytes(t, mem, make([]byte, size))
	state := &apiState{input: make([]byte, size)}
	for name, w := range map[string]uint64{
		"file_find": apiWork(t, state, mem, 16, needle, 6),
		"read":      apiWork(t, &apiState{input: make([]byte, size)}, mem, 2, buffer, size),
		"write":     apiWork(t, &apiState{}, mem, 3, buffer, size),
		"malloc":    apiWork(t, &apiState{}, mem, 18, size),
		"memstr":    apiWork(t, &apiState{}, mem, 51, buffer, size, needle, 6),
	} {
		if w > bound {
			t.Errorf("%s over %d bytes charged %d steps, want at most %d", name, size, w, bound)
		}
		if w < 2 {
			t.Errorf("%s charged %d steps", name, w)
		}
	}
}

func TestLinearIndexMatchesBytesIndex(t *testing.T) {
	rng := rand.New(rand.NewSource(20260926))
	for i := 0; i < 20000; i++ {
		alphabet := byte(1 + rng.Intn(3))
		h := make([]byte, rng.Intn(400))
		for j := range h {
			h[j] = 'a' + byte(rng.Intn(int(alphabet)))
		}
		n := make([]byte, 1+rng.Intn(130))
		for j := range n {
			n[j] = 'a' + byte(rng.Intn(int(alphabet)))
		}
		if rng.Intn(3) == 0 && len(h) > len(n) {
			copy(h[rng.Intn(len(h)-len(n)+1):], n)
		}
		if got, want := linearIndex(h, n), bytes.Index(h, n); got != want {
			t.Fatalf("linearIndex(%q, %q) = %d, want %d", h, n, got, want)
		}
	}
	// A needle of repeated bytes ending in another byte, in a haystack of
	// the repeated byte: linear, not quadratic.
	h := bytes.Repeat([]byte{'a'}, 750_000)
	n := append(bytes.Repeat([]byte{'a'}, 250_000), 'b')
	if got := linearIndex(h, n); got != -1 {
		t.Fatalf("pathological search: %d", got)
	}
}

func TestMemstrLongNeedleIsLinear(t *testing.T) {
	// A repetitive haystack and a needle that almost matches everywhere make
	// the standard search slow (about 1 s here); memstr must use the linear
	// search alone for long needles.
	pattern := []byte("0123456789abcdefg")
	h := bytes.Repeat(pattern, (2<<20)/len(pattern))
	n := append(bytes.Repeat(pattern, (1<<20)/len(pattern)), 'Z')
	mem := &memory{}
	hp, np := apiTestBytes(t, mem, h), apiTestBytes(t, mem, n)
	start := time.Now()
	apiTestCall(t, &apiState{}, mem, 51, apiFailure, hp, uint64(len(h)), np, uint64(len(n)))
	if elapsed := time.Since(start); elapsed > 250*time.Millisecond {
		t.Fatalf("memstr took %v on a long needle", elapsed)
	}
}
