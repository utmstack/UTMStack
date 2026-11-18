package bytecode

import (
	"bytes"
	"context"
	"errors"
	"runtime"
	"testing"
)

func callAPI(t *testing.T, state *apiState, mem *memory, id int, args ...uint64) uint64 {
	t.Helper()
	got, err := state.invoke(context.Background(), id, args, mem, func(uint64) error { return nil })
	if err != nil {
		t.Fatalf("API %d(%v): %v", id, args, err)
	}
	return got
}

func apiCallError(state *apiState, mem *memory, id int, args ...uint64) error {
	_, err := state.invoke(context.Background(), id, args, mem, func(uint64) error { return nil })
	return err
}

func TestAPIMallocGivesFreshZeroMemory(t *testing.T) {
	mem, state := &memory{}, &apiState{}
	ptr := callAPI(t, state, mem, 18, 16)
	if b, err := mem.read(ptr, 16); err != nil || !bytes.Equal(b, make([]byte, 16)) {
		t.Fatalf("malloc memory %v %v", b, err)
	}
	if err := mem.write(ptr, []byte("abc")); err != nil {
		t.Fatalf("malloc memory is not writable: %v", err)
	}
	apiTestCall(t, state, mem, 18, 0, 0)       // zero bytes: NULL
	apiTestCall(t, state, mem, 18, 0, 1<<30+1) // over 1 GiB: NULL
	apiTestCall(t, state, mem, 18, 0, 0xffffffff)
	// The reference allocates up to 1 GiB; this interpreter holds less, so
	// the run fails instead of returning memory it does not have.
	if err := apiCallError(state, mem, 18, 64<<20); !errors.Is(err, ErrLimit) {
		t.Fatalf("64 MiB: %v; want ErrLimit", err)
	}
}

func TestAPIMemstrFollowsTheReferenceHelper(t *testing.T) {
	mem, state := &memory{}, &apiState{}
	h := apiTestBytes(t, mem, []byte("xxabcabc"))
	n := apiTestBytes(t, mem, []byte("abc"))
	apiTestCall(t, state, mem, 51, 2, h, 8, n, 3)
	apiTestCall(t, state, mem, 51, apiFailure, h, 8, n, 0) // empty needle: no match
	apiTestCall(t, state, mem, 51, apiFailure, h, 0, n, 3) // empty haystack
	apiTestCall(t, state, mem, 51, apiFailure, n, 3, h, 8) // needle longer
	apiTestCall(t, state, mem, 51, 0, h, 8, h, 3)          // same pointer
	aa := apiTestBytes(t, mem, []byte("aab"))
	apiTestCall(t, state, mem, 51, 5, apiTestBytes(t, mem, []byte("xaaxaaab")), 8, aa, 3)
	// The interpreter checks both pointers before the call: a negative
	// length, a NULL pointer or a pointer at the end of its memory fail
	// the run.
	for _, args := range [][]uint64{{h, 0xffffffff, n, 3}, {0, 8, n, 3}, {h, 8, 0, 3}, {h + 8, 0, n, 3}, {h, 9, n, 3}} {
		if err := apiCallError(state, mem, 51, args...); !errors.Is(err, ErrRuntime) {
			t.Errorf("memstr%v: %v; want ErrRuntime", args, err)
		}
	}
}

func TestAPIReadNumberFollowsTheReference(t *testing.T) {
	check := func(input string, radix, want uint64, offset int) {
		t.Helper()
		state := &apiState{input: []byte(input)}
		if got := callAPI(t, state, nil, 23, radix); got != want || state.offset != offset {
			t.Errorf("%q radix %d: %#x at %d; want %#x at %d", input, radix, got, state.offset, want, offset)
		}
	}
	check("zz 0x1fZZ and 42 then padding......", 16, 0x1f, 7)
	check("value 1234 and more padding bytes!", 10, 1234, 10)
	check("zz 0x1fZZ and 42 then padding......", 10, 0, 4) // radix 10 stops at the x
	check("ab 0x1fZZ and 42 then padding......", 16, 0xab, 2)
	check("................42 and more text......", 10, 42, 18)
	check("short 7", 10, apiFailure, 0)             // fewer than 16 readable bytes
	check("..................", 10, apiFailure, 16) // blocks without digits move the offset
	check("................12", 10, apiFailure, 16) // the digits are in a short last block
	check("abcdefghij 12345", 10, apiFailure, 0)    // 16 bytes are needed from the digit
	check("12345678901234567890", 8, apiFailure, 0) // only radix 10 and 16
	want := uint64(0x4876e7ff)                      // 99999999999 cut to 32 bits (64-bit unsigned long)
	if runtime.GOOS == "windows" {
		want = apiFailure // 32-bit unsigned long saturates
	}
	check("99999999999 and padding bytes....", 10, want, 11)
}

func TestProgramUsesMallocMemory(t *testing.T) {
	program := func(size uint64) memoryFixture {
		return memoryFixture{
			types:  []string{memoryType(1, 2, 65, 32)},
			apiMax: 18, apis: []string{memoryAPI(18, 69, "malloc")},
			functions: []compatFunction{{locals: []int{65, 8}, result: 8, blocks: [][]string{{
				compatInst(65, 0, 33, compatFixed(1, 1), compatNum(18), compatConstant(size, 32)),
				compatInst(8, 0, 38, compatConstant(0x5a, 8), compatNum(0)),
				compatInst(8, 1, 39, compatNum(0)),
				compatRet(8, compatNum(1)),
			}}}},
		}
	}
	got, err := memoryParse(t, program(8)).RunWithInput(context.Background(), 0, nil, nil, 1000)
	if err != nil || got.Value != 0x5a {
		t.Fatalf("store and load through malloc memory: %+v, %v", got, err)
	}
	if _, err := memoryParse(t, program(0)).RunWithInput(context.Background(), 0, nil, nil, 1000); !errors.Is(err, ErrRuntime) {
		t.Fatalf("malloc(0) is NULL, so the store must fail: %v", err)
	}
}
