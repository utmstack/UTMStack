package bytecode

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"runtime"
	"testing"
)

func TestAPIPERawAddrAndSections(t *testing.T) {
	pe := &PEInfo{EntryPoint: 0x400, HeaderSize: 0x400, Sections: []Section{{RVA: 0x1000, VSize: 0x1000, Raw: 0x400, RSize: 0x200,
		Characteristics: 0x60000020, URVA: 0x1000, UVSize: 0xf00, URaw: 0x400, URSize: 0x1f0}}}
	state := &apiState{input: make([]byte, 0x600), pe: pe}
	apiTestCall(t, state, nil, 15, 0x410, 0x1010)
	apiTestCall(t, state, nil, 15, 0x100, 0x100)       // inside the headers
	apiTestCall(t, state, nil, 15, apiFailure, 0x1200) // past the section's raw data
	apiTestCall(t, state, nil, 15, apiFailure, 0x5000)
	mem := &memory{}
	buf := apiTestBytes(t, mem, make([]byte, 36))
	apiTestCall(t, state, mem, 20, 0, buf, 0)
	b, _ := mem.read(buf, 36)
	want := []uint32{0x1000, 0x1000, 0x400, 0x200, 0x60000020, 0x1000, 0xf00, 0x400, 0x1f0}
	for i, w := range want {
		if got := binary.LittleEndian.Uint32(b[i*4:]); got != w {
			t.Fatalf("section field %d = %#x, want %#x", i, got, w)
		}
	}
	apiTestCall(t, state, mem, 20, apiFailure, buf, 1)
	apiTestCall(t, &apiState{}, mem, 20, apiFailure, buf, 0) // not a PE
	apiTestCall(t, &apiState{input: []byte("x")}, nil, 15, apiFailure, 0)
	// The record buffer is checked before the call (36 bytes).
	if err := apiCallError(state, mem, 20, apiTestBytes(t, mem, make([]byte, 35)), 0); !errors.Is(err, ErrRuntime) {
		t.Fatalf("35-byte buffer: %v; want ErrRuntime", err)
	}
}

func TestAPIEnvironmentIsTheReferenceRecord(t *testing.T) {
	mem, state := &memory{}, &apiState{}
	buf := apiTestBytes(t, mem, make([]byte, 493))
	apiTestCall(t, state, mem, 72, 0, buf, 492)
	b, _ := mem.read(buf, 492)
	if b[484] != 8 || b[483] != 0 || string(b[28:33]) != "1.5.4" || binary.LittleEndian.Uint32(b[20:]) != EngineLevel {
		t.Fatalf("environment %v", b[480:])
	}
	if runtime.GOARCH == "amd64" && b[485] != 2 {
		t.Fatalf("arch %d, want 2 (x86_64)", b[485])
	}
	apiTestCall(t, state, mem, 72, apiFailure, buf, 493) // longer than the record
	small := apiTestBytes(t, mem, make([]byte, 8))
	apiTestCall(t, state, mem, 72, 0, small, 8)
	if err := apiCallError(state, mem, 72, small, 9); !errors.Is(err, ErrRuntime) {
		t.Fatalf("9 bytes into an 8-byte buffer: %v; want ErrRuntime", err)
	}
}

func TestEnvironmentMatchesPinnedDarwin154Profile(t *testing.T) {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Skip("the retained portable native reference is macOS ARM64")
	}
	b := environmentRecord()
	for offset, want := range map[int]uint32{
		0:  0x0403eaea, // OS, arch, compiler and both level bytes
		4:  0x08000000, // 64-bit little-endian
		8:  0x00040201, // pinned Clang build and no OS-feature bits
		12: 0x00040201,
		20: 234,
		24: 234,
	} {
		if got := binary.LittleEndian.Uint32(b[offset:]); got != want {
			t.Fatalf("environment field at %d = %#x, want %#x", offset, got, want)
		}
	}
	if b[484] != 8 || b[485] != 0 || b[486] != 4 || b[488] != 3 {
		t.Fatalf("platform bytes = %v", b[484:489])
	}
	for _, field := range []struct {
		start int
		want  string
	}{
		{28, "1.5.4"},
		{223, "Darwin"},
	} {
		end := bytes.IndexByte(b[field.start:field.start+65], 0)
		if end < 0 || string(b[field.start:field.start+end]) != field.want {
			t.Fatalf("environment string at %d = %q, want %q", field.start, b[field.start:field.start+65], field.want)
		}
	}
	_, release, version, machine := environmentUname()
	for _, field := range []struct {
		start int
		want  string
	}{
		{288, release}, {353, version}, {418, machine},
	} {
		end := bytes.IndexByte(b[field.start:field.start+65], 0)
		if end < 0 || string(b[field.start:field.start+end]) != field.want {
			t.Fatalf("uname field at %d = %q, want %q", field.start, b[field.start:field.start+65], field.want)
		}
	}
}

func TestStructPointerPrototypesMustMatchTheReference(t *testing.T) {
	// get_pe_section(struct cli_exe_section *, uint32): the pointee must be a
	// struct of nine i32, as the reference's prototype check requires.
	kind := 3 // struct
	section := func(elements ...int) memoryFixture {
		return memoryFixture{
			types:  []string{memoryType(kind, len(elements), elements...), memoryType(5, 1, 69), memoryType(1, 3, 32, 70, 32)},
			apiMax: 20, apis: []string{memoryAPI(20, 71, "get_pe_section")},
			functions: []compatFunction{{result: 32, blocks: [][]string{{compatRet(32, compatConstant(0, 32))}}}},
		}
	}
	if _, err := Parse(section(32, 32, 32, 32, 32, 32, 32, 32, 32).encode()); err != nil {
		t.Fatalf("matching prototype: %v", err)
	}
	for _, elements := range [][]int{{32, 32, 32, 32, 32, 32, 32, 32}, {32, 32, 32, 32, 32, 32, 32, 32, 16}} {
		if _, err := Parse(section(elements...).encode()); !errors.Is(err, ErrMalformed) {
			t.Fatalf("prototype %v: %v; want ErrMalformed", elements, err)
		}
	}
	kind = 2 // a packed struct is a different type
	if _, err := Parse(section(32, 32, 32, 32, 32, 32, 32, 32, 32).encode()); !errors.Is(err, ErrMalformed) {
		t.Fatalf("packed struct prototype: %v; want ErrMalformed", err)
	}
	_ = context.Background
}
