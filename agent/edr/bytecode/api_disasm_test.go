package bytecode

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"
)

// Records the reference decoder writes for "55" (push ebp) and "8b ec"
// (mov ebp, esp), captured with the oracle.
const (
	disasmPushEBP = "7c000000000305000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	disasmMovEBP  = "6d000000000305000000000000000003040000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
)

func TestAPIDisasmX86(t *testing.T) {
	mem := &memory{}
	marked := bytes.Repeat([]byte{0xaa}, 64)
	buf := apiTestBytes(t, mem, marked)
	record := func() []byte {
		b, err := mem.read(buf, 64)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	state := &apiState{input: []byte{0x55, 0x8b, 0xec}}
	apiTestCall(t, state, mem, 8, 1, buf, 64)
	if got := hex.EncodeToString(record()); got != disasmPushEBP {
		t.Fatalf("record at 0:\n got %s\nwant %s", got, disasmPushEBP)
	}
	if state.offset != 0 {
		t.Fatalf("offset moved to %d", state.offset)
	}
	state.offset = 1
	apiTestCall(t, state, mem, 8, 3, buf, 64) // offset + bytes consumed
	if got := hex.EncodeToString(record()); got != disasmMovEBP {
		t.Fatalf("record at 1:\n got %s\nwant %s", got, disasmMovEBP)
	}

	// At the end of the input nothing is decoded or written.
	if err := mem.write(buf, marked); err != nil {
		t.Fatal(err)
	}
	state.offset = 3
	apiTestCall(t, state, mem, 8, apiFailure, buf, 64)
	if !bytes.Equal(record(), marked) {
		t.Fatalf("record written at the end of input: %x", record())
	}

	// A failed decode (the call cut short by the end of input) zeroes only
	// the trailing 29 bytes.
	cut := &apiState{input: []byte{0xe8, 0x00, 0x00}}
	apiTestCall(t, cut, mem, 8, apiFailure, buf, 64)
	got := record()
	if !bytes.Equal(got[:35], marked[:35]) || !bytes.Equal(got[35:], make([]byte, 29)) {
		t.Fatalf("record after a failed decode: %x", got)
	}

	// At most 32 bytes are decoded: 32 prefixes before a nop are prefixes alone.
	long := &apiState{input: append(bytes.Repeat([]byte{0x2e}, 32), 0x90)}
	apiTestCall(t, long, mem, 8, apiFailure, buf, 64)

	// The record pointer is checked for the whole 64-byte record before the call.
	short := apiTestBytes(t, mem, make([]byte, 63))
	if err := apiCallError(state, mem, 8, short, 64); !errors.Is(err, ErrRuntime) {
		t.Fatalf("63-byte buffer: %v; want ErrRuntime", err)
	}
	if err := apiCallError(state, mem, 8, 0, 64); !errors.Is(err, ErrRuntime) {
		t.Fatalf("NULL buffer: %v; want ErrRuntime", err)
	}
}

func TestDisasmPrototypeMustMatchTheReference(t *testing.T) {
	// disasm_x86(struct DISASM_RESULT *, uint32): the pointee is
	// {i16, i8, i8, i8, [3 x [10 x i8]], [29 x i8]}.
	fixture := func(last int) memoryFixture {
		return memoryFixture{
			types: []string{
				memoryType(4, 10, 8),                  // 69
				memoryType(4, 3, 69),                  // 70
				memoryType(4, last, 8),                // 71
				memoryType(3, 6, 16, 8, 8, 8, 70, 71), // 72
				memoryType(5, 1, 72),                  // 73
				memoryType(1, 3, 32, 73, 32),          // 74
			},
			apiMax: 8, apis: []string{memoryAPI(8, 74, "disasm_x86")},
			functions: []compatFunction{{result: 32, blocks: [][]string{{compatRet(32, compatConstant(0, 32))}}}},
		}
	}
	if _, err := Parse(fixture(29).encode()); err != nil {
		t.Fatalf("matching prototype: %v", err)
	}
	if _, err := Parse(fixture(28).encode()); !errors.Is(err, ErrMalformed) {
		t.Fatalf("28-byte tail: %v; want ErrMalformed", err)
	}
}
