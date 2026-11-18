package bytecode

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// Extend the independent wire encoder in integer_compat_test.go with compound
// types, global initializers, APIs and the addressable-local flag. No private
// parsed instruction or memory representation is used to construct programs.
type memoryFixture struct {
	types, globals, apis []string
	apiMax, globalMax    int
	kind                 int
	logical              string
	functions            []compatFunction
}

func memoryType(kind, count int, elements ...int) string {
	s := compatFixed(uint64(kind), 1)
	if kind != 5 {
		s += compatNum(uint64(count))
	}
	for _, id := range elements {
		s += compatNum(uint64(id))
	}
	return s
}

func memoryGlobal(typ int, values ...uint64) string {
	s := compatNum(uint64(typ))
	for _, value := range values {
		b := []byte(compatNum(value))
		b[0] &^= 0x20
		s += string(b)
	}
	return s + "`"
}

func memoryGlobalRef(id int) string { return compatConstant(uint64(id), 0) }

func memoryAPI(id, typ int, name string) string {
	s := compatNum(uint64(id)) + compatNum(uint64(typ)) + "|" + compatNum(uint64(len(name)+1))
	for _, b := range append([]byte(name), 0) {
		s += compatFixed(uint64(b), 2)
	}
	return s
}

func (m memoryFixture) encode() []byte {
	header := "ClamBC" + compatNum(7) + compatNum(0) + "|" + compatNum(0) + compatNum(0) + compatNum(uint64(m.kind)) + strings.Repeat(compatNum(0), 3) + "|" + compatNum(0) + compatNum(uint64(5+len(m.types))) + compatNum(uint64(len(m.functions))) + compatNum(0x53e5493e9f3d1c30) + compatFixed(42, 2) + ":65536"
	globals := "G" + compatNum(uint64(m.globalMax)) + compatNum(uint64(len(m.globals)+1)) + compatNum(0) + "@`" + strings.Join(m.globals, "")
	lines := []string{header, m.logical, "T" + compatFixed(69, 2) + strings.Join(m.types, ""), "E" + compatNum(uint64(m.apiMax)) + compatNum(uint64(len(m.apis))) + strings.Join(m.apis, ""), globals}
	for _, f := range m.functions {
		declaration := "A" + compatFixed(uint64(len(f.args)), 1) + compatNum(uint64(f.result)) + "L" + compatNum(uint64(len(f.locals)))
		for _, types := range [][]int{f.args, f.locals} {
			for _, typ := range types {
				declaration += compatNum(uint64(typ&0x7fff)) + compatFixed(uint64(typ>>15), 1)
			}
		}
		count := 0
		for _, block := range f.blocks {
			count += len(block)
		}
		lines = append(lines, declaration+"F"+compatNum(uint64(count))+compatNum(uint64(len(f.blocks))))
		for i, block := range f.blocks {
			line := "B" + strings.Join(block, "")
			if i == len(f.blocks)-1 {
				line += "E"
			}
			lines = append(lines, line)
		}
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

func memoryParse(t *testing.T, m memoryFixture) *Program {
	t.Helper()
	p, err := Parse(m.encode())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return p
}

func memoryRun(t *testing.T, m memoryFixture, want uint64, wantErr error) {
	t.Helper()
	got, err := memoryParse(t, m).Run(context.Background(), 0, nil, 1000)
	if !errors.Is(err, wantErr) || err == nil && got != want {
		t.Fatalf("Run = %#x, %v; want %#x, %v", got, err, want, wantErr)
	}
}

func memoryArrayTypes(n, element int) []string {
	return []string{memoryType(4, n, element), memoryType(5, 1, 69)}
}

func TestMemoryVMAddressableScalarAndBooleanLoads(t *testing.T) {
	for _, width := range []int{1, 8, 16, 32, 64} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			storage := width
			if width == 1 {
				storage = 8
			}
			memoryRun(t, memoryFixture{functions: []compatFunction{{locals: []int{width | 0x8000, width}, result: width, blocks: [][]string{{
				compatInst(0, 0, 34, compatConstant(1, storage), compatNum(0)),
				compatInst(width, 1, 39, compatNum(0)), compatRet(width, compatNum(1)),
			}}}}}, 1, nil)
		})
	}
	// READ1 must mask memory contents as well as literals. A byte of value 2
	// is false, even when that byte was initialized by a memory operation.
	for _, value := range []uint64{0, 1, 2, 3, 255} {
		t.Run(fmt.Sprintf("boolean-byte-%d", value), func(t *testing.T) {
			memoryRun(t, memoryFixture{functions: []compatFunction{{locals: []int{0x8001, 1}, result: 1, blocks: [][]string{{
				compatInst(0, 0, 40, compatNum(0), compatConstant(value, 32), compatConstant(1, 64)),
				compatInst(1, 1, 39, compatNum(0)), compatRet(1, compatNum(1)),
			}}}}}, value&1, nil)
		})
	}
}

func TestMemoryVMGEPScalingAndArrayByteOffsets(t *testing.T) {
	// Upstream GEP1 scales by pointee size; GEPZ's array case instead uses
	// the supplied raw byte offset (bytecode_vm.c GEP1/GEPZ and bytecode.c
	// calc_gepz). Distances ensure a wrong implementation cannot pass by
	// applying the same wrong offset to both STORE and LOAD. GEP1 scales by
	// its i32* operand type; its result uses the interpreter's i8* width.
	for _, index := range []int{0, 1, 2, -1} {
		t.Run(fmt.Sprintf("scaled-%d", index), func(t *testing.T) {
			memoryRun(t, memoryFixture{types: memoryArrayTypes(4, 32), functions: []compatFunction{{locals: []int{0x8045, 67, 65, 32}, result: 32, blocks: [][]string{{
				compatInst(67, 1, 36, compatNum(70), compatNum(0), compatConstant(8, 32)),
				compatInst(65, 2, 35, compatNum(67), compatNum(1), compatConstant(uint64(uint32(int32(index))), 32)),
				compatInst(32, 3, 49, compatNum(2), compatNum(1)), compatRet(32, compatNum(3)),
			}}}}}, uint64(uint32(int32(index*4))), nil)
		})
	}
	memoryRun(t, memoryFixture{types: memoryArrayTypes(4, 32), functions: []compatFunction{{locals: []int{0x8045, 67, 65, 8}, result: 8, blocks: [][]string{{
		compatInst(67, 1, 36, compatNum(70), compatNum(0), compatConstant(0, 32)),
		compatInst(32, 0, 38, compatConstant(0x11223344, 32), compatNum(1)),
		compatInst(65, 2, 36, compatNum(70), compatNum(0), compatConstant(1, 32)),
		compatInst(8, 3, 39, compatNum(2)), compatRet(8, compatNum(3)),
	}}}}}, 0x33, nil)
}

func TestMemoryVMPackedStructFieldOffsets(t *testing.T) {
	// The interpreter packs both ClamBC struct kinds by summing field sizes
	// (bytecode.c typesize/calc_gepz); it does not use Go/C struct padding.
	for _, kind := range []int{2, 3} {
		for field, want := range []uint64{0, 1, 5} {
			t.Run(fmt.Sprintf("kind-%d/field-%d", kind, field), func(t *testing.T) {
				memoryRun(t, memoryFixture{types: []string{memoryType(kind, 3, 8, 32, 16), memoryType(5, 1, 69)}, functions: []compatFunction{{locals: []int{0x8045, 65, 32}, result: 32, blocks: [][]string{{
					compatInst(65, 1, 36, compatNum(70), compatNum(0), compatConstant(uint64(field), 32)),
					compatInst(32, 2, 49, compatNum(1), compatNum(0)), compatRet(32, compatNum(2)),
				}}}}}, want, nil)
			})
		}
	}
}

func TestMemoryVMStoreThroughDirectCallPointer(t *testing.T) {
	// CALL_DIRECT passes a GEPZ-derived i32* into the caller's addressable
	// array; the interpreter cannot copy an addressed local itself as an
	// argument. The callee's STORE derives its width from the typed source,
	// not dummy 0.
	memoryRun(t, memoryFixture{types: memoryArrayTypes(1, 32), functions: []compatFunction{
		{locals: []int{0x8045, 67, 32}, result: 32, blocks: [][]string{{
			compatInst(67, 1, 36, compatNum(70), compatNum(0), compatConstant(0, 32)),
			compatInst(32, 0, 38, compatConstant(7, 32), compatNum(1)),
			compatInst(32, 2, 32, compatFixed(1, 1), compatNum(2), compatNum(1)),
			compatInst(32, 2, 39, compatNum(1)),
			compatRet(32, compatNum(2)),
		}}},
		{args: []int{67}, locals: []int{32}, result: 32, blocks: [][]string{{
			compatInst(32, 1, 1, compatConstant(41, 32), compatConstant(1, 32)),
			compatInst(0, 0, 38, compatNum(1), compatNum(0)),
			compatInst(32, 1, 39, compatNum(0)), compatRet(32, compatNum(1)),
		}}},
	}}, 42, nil)
}

func TestMemoryVMMemsetMemcpyAndMemcmp(t *testing.T) {
	for _, same := range []bool{true, false} {
		t.Run(fmt.Sprintf("equal-%t", same), func(t *testing.T) {
			body := []string{
				compatInst(0, 0, 40, compatNum(0), compatConstant(0x141, 32), compatConstant(8, 64)),
				compatInst(0, 0, 41, compatNum(1), compatNum(0), compatConstant(8, 32)),
			}
			want := uint64(1)
			if !same {
				body = append(body, compatInst(0, 0, 40, compatNum(1), compatConstant(0x42, 32), compatConstant(1, 64)))
				want = 0
			}
			body = append(body,
				compatInst(32, 2, 43, compatNum(0), compatNum(1), compatConstant(8, 32)),
				compatInst(1, 3, 21, compatNum(32), compatNum(2), compatConstant(0, 32)), compatRet(1, compatNum(3)))
			memoryRun(t, memoryFixture{types: memoryArrayTypes(8, 8), functions: []compatFunction{{locals: []int{0x8045, 0x8045, 32, 1}, result: 1, blocks: [][]string{body}}}}, want, nil)
		})
	}
	// Byte truncation of memset's i32 fill argument and copied bytes are
	// observable independently of memcmp's implementation.
	memoryRun(t, memoryFixture{types: memoryArrayTypes(8, 8), functions: []compatFunction{{locals: []int{0x8045, 0x8045, 64}, result: 64, blocks: [][]string{{
		compatInst(0, 0, 40, compatNum(0), compatConstant(0x141, 32), compatConstant(8, 64)),
		compatInst(0, 0, 41, compatNum(1), compatNum(0), compatConstant(8, 32)),
		compatInst(64, 2, 39, compatNum(1)), compatRet(64, compatNum(2)),
	}}}}}, 0x4141414141414141, nil)
}

func TestMemoryVMMemsetAndMemmoveTakeI32Lengths(t *testing.T) {
	// The reference reads memset's and memmove's length as a 64-bit slot and
	// uses its low 32 bits for both the pointer check and the operation, so an
	// i32 length variable (as official programs emit) behaves as its value. A
	// negative one fails the run.
	for _, op := range []int{40, 42} {
		for _, length := range []uint64{8, 0xffffffff} {
			t.Run(fmt.Sprintf("op-%d/length-%#x", op, length), func(t *testing.T) {
				fill := compatInst(0, 0, 40, compatNum(0), compatConstant(0x41, 32), compatNum(3))
				if op == 42 {
					fill = compatInst(0, 0, 42, compatNum(0), memoryGlobalRef(2), compatNum(3))
				}
				want, wantErr := uint64(0x4141414141414141), error(nil)
				if length != 8 {
					wantErr = ErrLimit
				}
				memoryRun(t, memoryFixture{types: memoryArrayTypes(8, 8), globals: []string{
					memoryGlobal(69, 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A'), memoryGlobal(65, 0, 1),
				}, functions: []compatFunction{{locals: []int{0x8045, 64, 64, 32}, result: 64, blocks: [][]string{{
					compatInst(32, 3, 1, compatConstant(length, 32), compatConstant(0, 32)),
					fill,
					compatInst(64, 1, 39, compatNum(0)), compatRet(64, compatNum(1)),
				}}}}}, want, wantErr)
			})
		}
	}
}

func TestMemoryVMMemmoveOverlapAndMemcpyRejection(t *testing.T) {
	for _, forward := range []bool{true, false} {
		for _, op := range []int{41, 42} {
			t.Run(fmt.Sprintf("op-%d/forward-%t", op, forward), func(t *testing.T) {
				dst, src := compatNum(1), compatNum(0)
				wantBytes := []byte("ababcdef")
				if !forward {
					dst, src = src, dst
					wantBytes = []byte("cdefghgh")
				}
				lengthWidth := 64
				var wantErr error
				if op == 41 {
					lengthWidth, wantErr = 32, ErrRuntime // overlapping memcpy is undefined
				}
				memoryRun(t, memoryFixture{types: memoryArrayTypes(8, 8), globals: []string{
					memoryGlobal(69, 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'), memoryGlobal(65, 0, 1),
				}, functions: []compatFunction{{locals: []int{0x8045, 65, 64}, result: 64, blocks: [][]string{{
					compatInst(0, 0, 41, compatNum(0), memoryGlobalRef(2), compatConstant(8, 32)),
					compatInst(65, 1, 36, compatNum(70), compatNum(0), compatConstant(2, 32)),
					compatInst(0, 0, op, dst, src, compatConstant(6, lengthWidth)),
					compatInst(64, 2, 39, compatNum(0)), compatRet(64, compatNum(2)),
				}}}}}, binary.LittleEndian.Uint64(wantBytes), wantErr)
			})
		}
	}
}

func TestMemoryVMGlobalPointersAndRunIsolation(t *testing.T) {
	memoryRun(t, memoryFixture{types: memoryArrayTypes(4, 8), globals: []string{
		memoryGlobal(69, 0x11, 0x22, 0x33, 0x44), memoryGlobal(65, 1, 1),
	}, functions: []compatFunction{{locals: []int{8}, result: 8, blocks: [][]string{{
		compatInst(8, 0, 39, memoryGlobalRef(2)), compatRet(8, compatNum(0)),
	}}}}}, 0x22, nil)
	// The counter is a one-element array: the reference interpreter leaves
	// scalar integer globals zero but copies array initializers.
	p := memoryParse(t, memoryFixture{types: []string{memoryType(4, 1, 32)}, globals: []string{memoryGlobal(69, 7), memoryGlobal(67, 0, 1)}, functions: []compatFunction{{locals: []int{32}, result: 32, blocks: [][]string{{
		compatInst(32, 0, 39, memoryGlobalRef(2)),
		compatInst(32, 0, 1, compatNum(0), compatConstant(1, 32)),
		compatInst(0, 0, 38, compatNum(0), memoryGlobalRef(2)), compatRet(32, compatNum(0)),
	}}}}})
	for run := 0; run < 2; run++ {
		if got, err := p.Run(context.Background(), 0, nil, 100); err != nil || got != 8 {
			t.Fatalf("run %d = %d, %v; global state leaked between executions", run, got, err)
		}
	}
}

func TestMemoryVMStaleStackPointerFault(t *testing.T) {
	memoryRun(t, memoryFixture{types: memoryArrayTypes(8, 8), functions: []compatFunction{
		{locals: []int{65, 8}, result: 8, blocks: [][]string{{
			compatInst(65, 0, 32, compatFixed(0, 1), compatNum(2)),
			compatInst(8, 1, 39, compatNum(0)), compatRet(8, compatNum(1)),
		}}},
		{locals: []int{0x8045, 65}, result: 65, blocks: [][]string{{
			compatInst(0, 0, 40, compatNum(0), compatConstant(0x41, 32), compatConstant(8, 64)),
			compatInst(65, 1, 36, compatNum(70), compatNum(0), compatConstant(0, 32)),
			compatRet(65, compatNum(1)),
		}}},
	}}, 0, ErrRuntime)
}

func TestMemoryVMInvalidDereferences(t *testing.T) {
	for _, tc := range []struct {
		name string
		body []string
	}{
		{"null", []string{compatInst(64, 2, 39, memoryGlobalRef(0))}},
		{"uninitialized", []string{compatInst(64, 2, 39, compatNum(0))}},
		{"partial-initialization", []string{
			compatInst(0, 0, 40, compatNum(0), compatConstant(0, 32), compatConstant(4, 64)),
			compatInst(64, 2, 39, compatNum(0)),
		}},
		{"out-of-bounds", []string{
			compatInst(65, 1, 36, compatNum(70), compatNum(0), compatConstant(512, 32)),
			compatInst(64, 2, 39, compatNum(1)),
		}},
		{"negative-offset", []string{
			compatInst(65, 1, 36, compatNum(70), compatNum(0), compatConstant(0xffffffff, 32)),
			compatInst(64, 2, 39, compatNum(1)),
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := append(append([]string(nil), tc.body...), compatRet(64, compatNum(2)))
			memoryRun(t, memoryFixture{types: memoryArrayTypes(8, 8), functions: []compatFunction{{locals: []int{0x8045, 65, 64}, result: 64, blocks: [][]string{body}}}}, 0, ErrRuntime)
		})
	}
	for _, offset := range []uint64{0, 4} {
		t.Run(fmt.Sprintf("global-pointer-offset-%d", offset), func(t *testing.T) {
			target := uint64(1)
			if offset == 0 {
				target = 0 // explicit null pointer initializer
			}
			memoryRun(t, memoryFixture{types: memoryArrayTypes(4, 8), globals: []string{memoryGlobal(69, 1, 2, 3, 4), memoryGlobal(65, offset, target)}, functions: []compatFunction{{locals: []int{8}, result: 8, blocks: [][]string{{
				compatInst(8, 0, 39, memoryGlobalRef(2)), compatRet(8, compatNum(0)),
			}}}}}, 0, ErrRuntime)
		})
	}
}

func TestMemoryVMContextAndBulkBudget(t *testing.T) {
	p := memoryParse(t, memoryFixture{types: memoryArrayTypes(4096, 8), functions: []compatFunction{{locals: []int{0x8045}, result: 32, blocks: [][]string{{
		compatInst(0, 0, 40, compatNum(0), compatConstant(0, 32), compatConstant(4096, 64)),
		compatRet(32, compatConstant(42, 32)),
	}}}}})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := p.Run(ctx, 0, nil, 1000); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled Run: %v", err)
	}
	if _, err := p.Run(context.Background(), 0, nil, 100); !errors.Is(err, ErrLimit) {
		t.Fatalf("bulk operation escaped work budget: %v", err)
	}
	if got, err := p.Run(context.Background(), 0, nil, 200); err != nil || got != 42 {
		t.Fatalf("sufficient bulk budget: %d, %v", got, err)
	}
}

func TestMemoryVMAllocationBounds(t *testing.T) {
	// An acyclic chain remains below the call-depth bound, so this must hit
	// live memory's bound even when instruction fuel is sufficient.
	functions := make([]compatFunction, 33)
	for i := range functions {
		body := []string{compatRet(32, compatConstant(42, 32))}
		if i+1 < len(functions) {
			body = []string{compatInst(32, 1, 32, compatFixed(0, 1), compatNum(uint64(i+2))), compatRet(32, compatNum(1))}
		}
		functions[i] = compatFunction{locals: []int{0x8045, 32}, result: 32, blocks: [][]string{body}}
	}
	p := memoryParse(t, memoryFixture{types: memoryArrayTypes(1<<20, 8), functions: functions})
	if _, err := p.Run(context.Background(), 0, nil, 1_000_000); !errors.Is(err, ErrLimit) {
		t.Fatalf("live allocation bound: %v", err)
	}
	// Retired frames do not accumulate against live memory. Allocator churn
	// is bounded instead by the work budget, which charges every frame byte.
	body := make([]string, 65, 66)
	for i := range body {
		body[i] = compatInst(32, 0, 32, compatFixed(0, 1), compatNum(2))
	}
	body = append(body, compatRet(32, compatNum(0)))
	p = memoryParse(t, memoryFixture{types: memoryArrayTypes(1<<20, 8), functions: []compatFunction{
		{locals: []int{32}, result: 32, blocks: [][]string{body}},
		{locals: []int{0x8045}, result: 32, blocks: [][]string{{compatRet(32, compatConstant(0, 32))}}},
	}})
	if got, err := p.Run(context.Background(), 0, nil, 2_000_000); err != nil || got != 0 {
		t.Fatalf("sequential retired frames: %d, %v", got, err)
	}
	if _, err := p.Run(context.Background(), 0, nil, 1_000_000); !errors.Is(err, ErrLimit) {
		t.Fatalf("frame churn escaped the work budget: %v", err)
	}
}

func TestMemoryVMAPIDeclarationsValidatedWhenUnused(t *testing.T) {
	for _, tc := range []struct {
		name, declaration, prototype string
		want                         error
	}{
		{"valid", memoryAPI(2, 69, "read"), memoryType(1, 3, 32, 65, 32), nil},
		{"unknown-id", memoryAPI(4095, 69, "unknown"), memoryType(1, 3, 32, 65, 32), ErrUnsupported},
		{"wrong-name", memoryAPI(2, 69, "write"), memoryType(1, 3, 32, 65, 32), ErrMalformed},
		{"wrong-result", memoryAPI(2, 69, "read"), memoryType(1, 3, 64, 65, 32), ErrMalformed},
		{"wrong-pointer", memoryAPI(2, 69, "read"), memoryType(1, 3, 32, 66, 32), ErrMalformed},
		{"wrong-arity", memoryAPI(2, 69, "read"), memoryType(1, 2, 32, 65), ErrMalformed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := memoryFixture{types: []string{tc.prototype}, apis: []string{tc.declaration}, apiMax: 4095, functions: []compatFunction{{result: 32, blocks: [][]string{{compatRet(32, compatConstant(0, 32))}}}}}
			if _, err := Parse(m.encode()); !errors.Is(err, tc.want) {
				t.Fatalf("Parse error = %v; want %v", err, tc.want)
			}
		})
	}
}

func memoryAPICall(dest, id int, args ...string) string {
	fields := []string{compatFixed(uint64(len(args)), 1), compatNum(uint64(id))}
	return compatInst(32, dest, 33, append(fields, args...)...)
}

// This program exercises the wire declarations, operand validation, VM and API
// adapter together. Input begins with a NUL-terminated label and has at least
// five bytes. It extracts overlapping input slices, updates a private global,
// captures a diagnostic, and overwrites the buffer after copying its label.
func memoryIOProgram(tail ...string) memoryFixture {
	body := []string{
		memoryAPICall(1, 2, compatNum(0), compatConstant(4, 32)),
		memoryAPICall(1, 3, compatNum(0), compatNum(1)),
		memoryAPICall(1, 4, compatConstant(0xffffffff, 32), compatConstant(1, 32)),
		memoryAPICall(1, 2, compatNum(0), compatConstant(2, 32)),
		memoryAPICall(1, 3, compatNum(0), compatNum(1)),
		memoryAPICall(1, 4, compatConstant(0xfffffffe, 32), compatConstant(2, 32)),
		memoryAPICall(1, 2, compatNum(0), compatConstant(2, 32)),
		memoryAPICall(1, 3, compatNum(0), compatNum(1)),
		memoryAPICall(1, 4, compatConstant(0, 32), compatConstant(0, 32)),
		memoryAPICall(1, 2, compatNum(0), compatConstant(32, 32)),
		memoryAPICall(1, 3, compatNum(0), compatNum(1)),
		memoryAPICall(1, 5, compatNum(0), compatConstant(0, 32)),
		compatInst(32, 2, 39, memoryGlobalRef(2)),
		compatInst(32, 2, 1, compatNum(2), compatConstant(1, 32)),
		compatInst(0, 0, 38, compatNum(2), memoryGlobalRef(2)),
		memoryAPICall(1, 7, compatNum(2)),
		compatInst(0, 0, 40, compatNum(0), compatConstant('Z', 32), compatConstant(32, 64)),
	}
	body = append(body, tail...)
	body = append(body, compatRet(32, compatNum(2)))
	return memoryFixture{
		types: append(memoryArrayTypes(32, 8),
			memoryType(1, 3, 32, 65, 32), memoryType(1, 3, 32, 32, 32), memoryType(1, 2, 32, 32), memoryType(4, 1, 32)),
		globals: []string{memoryGlobal(74, 7), memoryGlobal(67, 0, 1)},
		apiMax:  7, apis: []string{
			memoryAPI(2, 71, "read"), memoryAPI(3, 71, "write"),
			memoryAPI(4, 72, "seek"), memoryAPI(5, 71, "setvirusname"), memoryAPI(7, 73, "debug_print_uint"),
		},
		functions: []compatFunction{{locals: []int{0x8045, 32, 32}, result: 32, blocks: [][]string{body}}},
	}
}

func memoryIOExpected(input []byte) Result {
	output := append([]byte(nil), input[:4]...)
	output = append(output, input[3:5]...)
	output = append(output, input[len(input)-2:]...)
	output = append(output, input...)
	return Result{Value: 8, Output: output, VirusName: string(input[:bytes.IndexByte(input, 0)]), VirusNameSet: true, Diagnostics: []string{"8"}}
}

func TestMemoryVMAPIInputOutputSnapshots(t *testing.T) {
	// bytecode_api.c read/seek/write keep separate input and extraction
	// positions. setvirusname ignores its length argument and consumes a
	// NUL-terminated name; the native adapter must copy it before retirement.
	p := memoryParse(t, memoryIOProgram())
	input := []byte("UTMStack.Test\x00abcdefghi")
	original := bytes.Clone(input)
	want := memoryIOExpected(input)
	got, err := p.RunWithInput(context.Background(), 0, nil, input, 1000)
	want.WorkUsed = got.WorkUsed
	if err != nil || got.WorkUsed == 0 || !reflect.DeepEqual(got, want) {
		t.Fatalf("RunWithInput = %+v, %v; want %+v", got, err, want)
	}
	if !bytes.Equal(input, original) {
		t.Fatal("input buffer was modified")
	}
	input[0] = '!'
	if !reflect.DeepEqual(got, want) {
		t.Fatal("result aliases caller's input")
	}
	got.Output[0] = '?'
	got.Diagnostics[0] = "modified"
	again, err := p.RunWithInput(context.Background(), 0, nil, original, 1000)
	want.WorkUsed = again.WorkUsed
	if err != nil || again.WorkUsed == 0 || !reflect.DeepEqual(again, want) {
		t.Fatalf("result mutation leaked into subsequent execution: %+v, %v", again, err)
	}
}

func TestMemoryVMAPIErrorDiscardsPartialSnapshots(t *testing.T) {
	// Every tail executes after output, label and diagnostic production.
	// A resource error is no more a successful detection than a VM fault.
	for _, tc := range []struct {
		name string
		tail string
		want error
	}{
		{"abort", compatInst(0, 0, 45), ErrRuntime},
		{"division-by-zero", compatInst(32, 1, 4, compatConstant(1, 32), compatConstant(0, 32)), ErrRuntime},
		{"bulk-limit", compatInst(0, 0, 40, compatNum(0), compatConstant(0, 32), compatConstant(33<<20, 64)), ErrLimit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := memoryParse(t, memoryIOProgram(tc.tail))
			got, err := p.RunWithInput(context.Background(), 0, nil, []byte("UTMStack.Test\x00abcdefghi"), 1000)
			if !errors.Is(err, tc.want) || !failedEmpty(got, MaxRunSteps) {
				t.Fatalf("failed execution retained partial result: %+v, %v; want zero Result, %v", got, err, tc.want)
			}
		})
	}
}

func TestMemoryVMFileSizeGlobal(t *testing.T) {
	// Generated upstream declarations identify __clambc_filesize as a
	// read-only uint32_t[1], external target 0x8004. The offset is in bytes.
	m := memoryFixture{globalMax: 0x8004, globals: []string{memoryGlobal(67, 0, 0x8004)}, functions: []compatFunction{{locals: []int{32}, result: 32, blocks: [][]string{{
		compatInst(32, 0, 39, memoryGlobalRef(1)), compatRet(32, compatNum(0)),
	}}}}}
	p := memoryParse(t, m)
	for _, n := range []int{0, 1, 257, 4096} {
		got, err := p.RunWithInput(context.Background(), 0, nil, make([]byte, n), 100)
		if err != nil || got.Value != uint64(n) {
			t.Fatalf("filesize with %d input bytes: %+v, %v", n, got, err)
		}
	}
	for _, mutation := range []string{
		compatInst(32, 0, 38, compatConstant(0, 32), memoryGlobalRef(1)),
		compatInst(0, 0, 40, memoryGlobalRef(1), compatConstant(0, 32), compatConstant(4, 64)),
	} {
		m.functions[0].blocks = [][]string{{mutation, compatRet(32, compatConstant(42, 32))}}
		got, err := memoryParse(t, m).RunWithInput(context.Background(), 0, nil, []byte("input"), 100)
		if !errors.Is(err, ErrRuntime) || !failedEmpty(got, MaxRunSteps) {
			t.Fatalf("write to external filesize global: %+v, %v; want zero Result, ErrRuntime", got, err)
		}
	}
}

func TestMemoryVMConcurrentInputIsolation(t *testing.T) {
	p := memoryParse(t, memoryIOProgram())
	const workers = 24
	start := make(chan struct{})
	failures := make(chan error, workers)
	var wg sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-start
			for run := 0; run < 5; run++ {
				ctx, cancel := context.WithCancel(context.Background())
				input := []byte(fmt.Sprintf("Rule.%02d.%d\x00%s", id, run, strings.Repeat(string(byte('a'+id)), 5+id%8)))
				want := memoryIOExpected(input)
				canceled := id%4 == 0 && run%2 == 0
				if canceled {
					cancel()
				}
				got, err := p.RunWithInput(ctx, 0, nil, input, 1000)
				cancel()
				want.WorkUsed = got.WorkUsed
				if canceled {
					if !errors.Is(err, context.Canceled) || !failedEmpty(got, MaxRunSteps) {
						failures <- fmt.Errorf("worker %d run %d canceled result: %+v, %v", id, run, got, err)
						return
					}
				} else if err != nil || got.WorkUsed == 0 || !reflect.DeepEqual(got, want) {
					failures <- fmt.Errorf("worker %d run %d: %+v, %v; want %+v", id, run, got, err, want)
					return
				}
				// Owned snapshots can be edited while other runs continue.
				if len(got.Output) != 0 {
					got.Output[0] ^= 0xff
					got.Diagnostics[0] = "owned by caller"
				}
			}
		}(worker)
	}
	close(start)
	wg.Wait()
	close(failures)
	for err := range failures {
		t.Error(err)
	}
}

// memoryDebugLoop prints the loop counter iterations times through
// debug_print_uint, then sets a label and returns zero.
func memoryDebugLoop(label string, iterations uint64) memoryFixture {
	components := make([]uint64, 8)
	for i := 0; i < len(label); i++ {
		components[i] = uint64(label[i])
	}
	return memoryFixture{
		types:   append(memoryArrayTypes(8, 8), memoryType(1, 2, 32, 32), memoryType(1, 3, 32, 65, 32)),
		globals: []string{memoryGlobal(69, components...), memoryGlobal(65, 0, 1)},
		apiMax:  7, apis: []string{memoryAPI(5, 72, "setvirusname"), memoryAPI(7, 71, "debug_print_uint")},
		functions: []compatFunction{{locals: []int{32, 32, 1}, result: 32, blocks: [][]string{
			{compatInst(0, 0, 34, compatConstant(0, 32), compatNum(0)), "T" + compatFixed(18, 2) + compatNum(1)},
			{
				memoryAPICall(1, 7, compatNum(0)),
				compatInst(32, 0, 1, compatNum(0), compatConstant(1, 32)),
				compatInst(1, 2, 25, compatNum(32), compatNum(0), compatConstant(iterations, 32)),
				"T" + compatFixed(17, 2) + compatNum(2) + compatNum(1) + compatNum(2),
			},
			{memoryAPICall(1, 5, memoryGlobalRef(2), compatConstant(uint64(len(label)), 32)), compatRet(32, compatConstant(0, 32))},
		}}},
	}
}

func TestDebugOutputCapDoesNotFailDetection(t *testing.T) {
	// Upstream debug_print_uint returns zero and has no effect when debug
	// output is disabled. Filling the bounded diagnostic capture inside the
	// work budget must not turn a completed labelled run into a failure.
	const label = "Test.Db"
	for _, iterations := range []uint64{1000, 20000} {
		t.Run(fmt.Sprint(iterations), func(t *testing.T) {
			got, err := memoryParse(t, memoryDebugLoop(label, iterations)).RunWithInput(context.Background(), 0, nil, nil, 1_000_000)
			if err != nil || got.Value != 0 || !got.VirusNameSet || got.VirusName != label {
				t.Fatalf("RunWithInput = %+v, %v; want label %q", got.VirusName, err, label)
			}
			size := 0
			for i, message := range got.Diagnostics {
				if message != fmt.Sprint(i) {
					t.Fatalf("diagnostic %d = %q; retained diagnostics must be an ordered prefix", i, message)
				}
				size += len(message) + 1
			}
			if size > maxAPIDebug || got.DiagnosticsTruncated != (len(got.Diagnostics) < int(iterations)) {
				t.Fatalf("retained %d diagnostics (%d bytes), truncated %t", len(got.Diagnostics), size, got.DiagnosticsTruncated)
			}
		})
	}
}

func TestManySequentialCallsWithinBudget(t *testing.T) {
	// Upstream allocates call frames on a reusable stack and bounds only its
	// depth and time. More than 65,535 sequential calls with two live frames
	// must complete inside the work budget instead of exhausting region IDs.
	const calls = 70000
	p := memoryParse(t, memoryFixture{functions: []compatFunction{
		{locals: []int{32, 32, 1}, result: 32, blocks: [][]string{
			{compatInst(0, 0, 34, compatConstant(0, 32), compatNum(0)), "T" + compatFixed(18, 2) + compatNum(1)},
			{
				compatInst(32, 1, 32, compatFixed(1, 1), compatNum(2), compatNum(0)),
				compatInst(32, 0, 1, compatNum(0), compatConstant(1, 32)),
				compatInst(1, 2, 25, compatNum(32), compatNum(0), compatConstant(calls, 32)),
				"T" + compatFixed(17, 2) + compatNum(2) + compatNum(1) + compatNum(2),
			},
			{compatRet(32, compatNum(1))},
		}},
		{args: []int{32}, result: 32, blocks: [][]string{{compatRet(32, compatNum(0))}}},
	}})
	got, err := p.RunWithInput(context.Background(), 0, nil, nil, 1_000_000)
	if err != nil || got.Value != calls-1 {
		t.Fatalf("RunWithInput = %d, %v; want %d", got.Value, err, calls-1)
	}
}

func TestStaleFramePointerFaultsAfterSlotReuse(t *testing.T) {
	// Function 1 returns a pointer into its own frame. Function 2's frame
	// then reuses that retired region slot, and it receives the stale
	// pointer. Without the slot generation the load would read function 2's
	// own argument bytes; it must fault instead.
	memoryRun(t, memoryFixture{types: memoryArrayTypes(1, 32), functions: []compatFunction{
		{locals: []int{65, 32}, result: 32, blocks: [][]string{{
			compatInst(65, 0, 32, compatFixed(0, 1), compatNum(2)),
			compatInst(32, 1, 32, compatFixed(1, 1), compatNum(3), compatNum(0)),
			compatRet(32, compatNum(1)),
		}}},
		{locals: []int{0x8045, 65}, result: 65, blocks: [][]string{{
			compatInst(65, 1, 36, compatNum(70), compatNum(0), compatConstant(0, 32)),
			compatInst(32, 0, 38, compatConstant(7, 32), compatNum(1)),
			compatRet(65, compatNum(1)),
		}}},
		{args: []int{65}, locals: []int{32}, result: 32, blocks: [][]string{{
			compatInst(32, 1, 39, compatNum(0)), compatRet(32, compatNum(1)),
		}}},
	}}, 0, ErrRuntime)
}

func TestIntegerGlobalInitializerMatchesReference(t *testing.T) {
	// The pinned cli_bytecode_prepare_interpreter skips globals whose type ID
	// is below 65, so an i32 global initialized to 3 reads as zero upstream,
	// both as a direct operand and through a pointer to it.
	for name, body := range map[string][]string{
		"direct-operand":  {compatInst(32, 0, 1, memoryGlobalRef(1), compatConstant(0, 32))},
		"through-pointer": {compatInst(32, 0, 39, memoryGlobalRef(2))},
	} {
		t.Run(name, func(t *testing.T) {
			memoryRun(t, memoryFixture{globals: []string{memoryGlobal(32, 3), memoryGlobal(67, 0, 1)}, functions: []compatFunction{{locals: []int{32}, result: 32, blocks: [][]string{
				append(body, compatRet(32, compatNum(0))),
			}}}}, 0, nil)
		})
	}
}

func TestAggregateGlobalInitializersMatchReference(t *testing.T) {
	// Upstream leaves struct globals zero and copies the first count raw
	// initializer components into arrays whose elements are 1, 2, 4 or 8
	// bytes, truncated to the element size. Arrays of pointers therefore hold
	// raw offset/ID numbers, and arrays of other element sizes stay zero.
	types := []string{
		memoryType(3, 2, 8, 32),  // 69: struct {i8, i32}
		memoryType(4, 2, 65),     // 70: [2 x i8*]
		memoryType(4, 2, 1),      // 71: [2 x i1]
		memoryType(4, 2, 73),     // 72: [2 x struct {i32, i32}]
		memoryType(3, 2, 32, 32), // 73: struct {i32, i32}
		memoryType(4, 2, 75),     // 74: [2 x [2 x i8]]
		memoryType(4, 2, 8),      // 75: [2 x i8]
		memoryType(4, 2, 77),     // 76: [2 x packed struct {i8, i16}]
		memoryType(2, 2, 8, 16),  // 77: packed struct {i8, i16}, 3 bytes
		memoryType(4, 2, 16),     // 78: [2 x i16]
		memoryType(4, 1, 64),     // 79: [1 x i64]
	}
	globals := []string{
		memoryGlobal(69, 0x11, 0x22334455),
		memoryGlobal(70, 4, 1, 0, 1),
		memoryGlobal(71, 3, 2),
		memoryGlobal(72, 0x11, 0x22, 0x33, 0x44),
		memoryGlobal(74, 0x11, 0x22, 0x33, 0x44),
		memoryGlobal(76, 0x11, 0x2233, 0x44, 0x5566),
		memoryGlobal(78, 0x11234, 0x5678),
		memoryGlobal(79, 0x1122334455667788),
	}
	for _, tc := range []struct {
		name                  string
		global, offset, width int
		want                  uint64
	}{
		{"struct", 1, 0, 8, 0}, {"struct-field", 1, 1, 32, 0},
		{"pointer-array", 2, 0, 64, 4}, {"pointer-array-second", 2, 8, 64, 1},
		{"i1-array", 3, 0, 16, 0x0203},
		{"struct-array", 4, 0, 64, 0x11}, {"struct-array-second", 4, 8, 64, 0x22},
		{"nested-array", 5, 0, 32, 0x00220011},
		{"three-byte-elements", 6, 0, 16, 0},
		{"i16-array", 7, 0, 32, 0x56781234},
		{"i64-array", 8, 0, 64, 0x1122334455667788},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Global 9 is an i8* to the tested object at the tested offset.
			all := append(append([]string(nil), globals...), memoryGlobal(65, uint64(tc.offset), uint64(tc.global)))
			memoryRun(t, memoryFixture{types: types, globals: all, functions: []compatFunction{{locals: []int{tc.width}, result: tc.width, blocks: [][]string{{
				compatInst(tc.width, 0, 39, memoryGlobalRef(len(all))), compatRet(tc.width, compatNum(0)),
			}}}}}, tc.want, nil)
		})
	}
}

func TestNullPointerGlobalValueIsIncomplete(t *testing.T) {
	// Upstream composes a pointer global initialized to (0, 0), including the
	// compact zero initializer, into a pointer to offset zero of its shared
	// global block. That value has no native equivalent, so reading it must
	// fail rather than behave as null; storing a pointer first is well defined.
	for _, null := range []string{memoryGlobal(65, 0, 0), compatNum(65) + "@`"} {
		globals := []string{null, memoryGlobal(65, 0, 3), memoryGlobal(69, 7, 8, 9, 10), memoryGlobal(65, 0, 1)}
		read := compatInst(8, 0, 39, memoryGlobalRef(1))
		for name, tc := range map[string]struct {
			body []string
			want error
		}{
			"load-through":  {[]string{read}, ErrRuntime},
			"compare-null":  {[]string{compatInst(1, 1, 21, compatNum(65), memoryGlobalRef(1), memoryGlobalRef(0)), compatInst(0, 0, 34, compatConstant(7, 8), compatNum(0))}, ErrRuntime},
			"after-a-store": {[]string{compatInst(65, 0, 38, memoryGlobalRef(2), memoryGlobalRef(4)), read}, nil},
		} {
			t.Run(name, func(t *testing.T) {
				memoryRun(t, memoryFixture{types: memoryArrayTypes(4, 8), globals: globals, functions: []compatFunction{{locals: []int{8, 1}, result: 8, blocks: [][]string{
					append(append([]string(nil), tc.body...), compatRet(8, compatNum(0))),
				}}}}, 7, tc.want)
			})
		}
	}
}
