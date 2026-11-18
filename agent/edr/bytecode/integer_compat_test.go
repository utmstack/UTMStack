package bytecode

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// These fixtures independently encode the public ClamBC format, rather than
// inspecting parser internals to discover instruction boundaries. Every block
// explicitly lists its instructions, so calls and effect-only instructions need
// no opcode-dependent counting heuristic. They contain no upstream code or malware.
type compatFunction struct {
	args, locals []int
	result       int
	blocks       [][]string
}

func compatNum(v uint64) string {
	var digits []byte
	for v != 0 {
		digits = append(digits, 0x60|byte(v&15))
		v >>= 4
	}
	return string(append([]byte{0x60 + byte(len(digits))}, digits...))
}

func compatFixed(v uint64, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 0x60 | byte(v&15)
		v >>= 4
	}
	return string(b)
}

func compatConstant(value uint64, storageBits int) string {
	b := []byte(compatNum(value))
	b[0] &^= 0x20
	return string(b) + compatFixed(uint64(storageBits/8), 1)
}

func compatInst(resultType, dest, opcode int, fields ...string) string {
	return compatNum(uint64(resultType)) + compatNum(uint64(dest)) + compatFixed(uint64(opcode), 2) + strings.Join(fields, "")
}

func compatRet(width int, operand string) string {
	return "T" + compatFixed(19, 2) + compatNum(uint64(width)) + operand
}

func compatProgram(functions ...compatFunction) []byte {
	header := "ClamBC" + compatNum(7) + compatNum(0) + "|" + compatNum(0) + strings.Repeat(compatNum(0), 5) + "|" + compatNum(0) + compatNum(5) + compatNum(uint64(len(functions))) + compatNum(0x53e5493e9f3d1c30) + compatFixed(42, 2) + ":4096"
	lines := []string{header, "", "T" + compatFixed(69, 2), "E" + compatNum(0) + compatNum(0), "G" + compatNum(0) + compatNum(1) + compatNum(0) + "@`"}
	for _, f := range functions {
		declaration := "A" + compatFixed(uint64(len(f.args)), 1) + compatNum(uint64(f.result)) + "L" + compatNum(uint64(len(f.locals)))
		for _, types := range [][]int{f.args, f.locals} {
			for _, typ := range types {
				declaration += compatNum(uint64(typ)) + compatFixed(0, 1)
			}
		}
		count := 0
		for _, block := range f.blocks {
			count += len(block)
		}
		declaration += "F" + compatNum(uint64(count)) + compatNum(uint64(len(f.blocks)))
		lines = append(lines, declaration)
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

func compatRun(t *testing.T, want uint64, args []uint64, functions ...compatFunction) {
	t.Helper()
	p, err := Parse(compatProgram(functions...))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got, err := p.Run(context.Background(), 0, args, 100)
	if err != nil || got != want {
		t.Fatalf("Run = %#x, %v; want %#x, nil", got, err, want)
	}
}

func TestIntegerCompatCopyIgnoresDummyResult(t *testing.T) {
	// libclamav/bytecode.c parseBB OP_BC_COPY (lines 1369-1379) derives
	// the effective type from operand 1. bytecode_vm.c lines 1072-1100
	// writes that operand, not the serialized instruction destination.
	for _, dummyType := range []int{0, 8, 32} {
		for dest, want := range []uint64{99, 42} {
			t.Run(fmt.Sprintf("dummy-type-%d/return-%d", dummyType, dest), func(t *testing.T) {
				compatRun(t, want, nil, compatFunction{locals: []int{32, 32}, result: 32, blocks: [][]string{{
					compatInst(32, 0, 1, compatConstant(99, 32), compatConstant(0, 32)),
					compatInst(dummyType, 0, 34, compatConstant(42, 32), compatNum(1)),
					compatRet(32, compatNum(uint64(dest))),
				}}})
			})
		}
	}
}

func TestIntegerCompatBooleanConstants(t *testing.T) {
	// libclamav/bytecode.c readOperand (lines 494-526) encodes constant
	// STORAGE in bytes. A boolean therefore uses byte width 8 on the wire.
	// bytecode_vm.c READ1 (lines 329-331) masks the low bit: 2 is false.
	for _, value := range []uint64{0, 1, 2, 3, 255} {
		c := compatConstant(value, 8)
		want := value & 1
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			t.Run("return", func(t *testing.T) {
				compatRun(t, want, nil, compatFunction{result: 1, blocks: [][]string{{compatRet(1, c)}}})
			})
			t.Run("branch", func(t *testing.T) {
				compatRun(t, want, nil, compatFunction{result: 32, blocks: [][]string{
					{"T" + compatFixed(17, 2) + c + compatNum(1) + compatNum(2)},
					{compatRet(32, compatConstant(1, 32))},
					{compatRet(32, compatConstant(0, 32))},
				}})
			})
			t.Run("select", func(t *testing.T) {
				compatRun(t, want, nil, compatFunction{locals: []int{32}, result: 32, blocks: [][]string{{
					compatInst(32, 0, 31, c, compatConstant(1, 32), compatConstant(0, 32)),
					compatRet(32, compatNum(0)),
				}}})
			})
			t.Run("arithmetic", func(t *testing.T) {
				compatRun(t, want, nil, compatFunction{locals: []int{1}, result: 1, blocks: [][]string{{
					compatInst(1, 0, 11, c, compatConstant(3, 8)),
					compatRet(1, compatNum(0)),
				}}})
			})
			t.Run("comparison", func(t *testing.T) {
				compatRun(t, want, nil, compatFunction{locals: []int{1}, result: 1, blocks: [][]string{{
					compatInst(1, 0, 21, compatNum(1), c, compatConstant(1, 8)),
					compatRet(1, compatNum(0)),
				}}})
			})
			t.Run("copy", func(t *testing.T) {
				compatRun(t, want, nil, compatFunction{locals: []int{1}, result: 1, blocks: [][]string{{
					compatInst(0, 0, 34, c, compatNum(0)),
					compatRet(1, compatNum(0)),
				}}})
			})
			t.Run("call-argument", func(t *testing.T) {
				compatRun(t, want, nil,
					compatFunction{locals: []int{32}, result: 32, blocks: [][]string{{
						compatInst(32, 0, 32, compatFixed(1, 1), compatNum(2), c),
						compatRet(32, compatNum(0)),
					}}},
					compatFunction{args: []int{1}, locals: []int{32}, result: 32, blocks: [][]string{{
						compatInst(32, 1, 16, compatNum(0)),
						compatRet(32, compatNum(1)),
					}}})
			})
		})
	}
}

func TestIntegerCompatVoidCallPreservesDummyDestination(t *testing.T) {
	// bytecode_vm.c DEFINE_OP_BC_RET_VOID (lines 547-562) restores the
	// caller's frame without assigning a value to the call's dummy dest.
	compatRun(t, 99, nil,
		compatFunction{locals: []int{32}, result: 32, blocks: [][]string{{
			compatInst(32, 0, 1, compatConstant(99, 32), compatConstant(0, 32)),
			compatInst(0, 0, 32, compatFixed(0, 1), compatNum(2)),
			compatRet(32, compatNum(0)),
		}}},
		compatFunction{result: 0, blocks: [][]string{{"T" + compatFixed(20, 2)}}})
}

func TestIntegerCompatVoidAbortIsRuntimeFault(t *testing.T) {
	// clambc.h declares OP_BC_ABORT with zero operands. Upstream's
	// interpreter has no successful ABORT case; it falls to an error.
	// Accepting its dummy void type must never turn it into a return.
	p, err := Parse(compatProgram(compatFunction{result: 32, blocks: [][]string{{
		compatInst(0, 0, 45), compatRet(32, compatConstant(42, 32)),
	}}}))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if _, err := p.Run(context.Background(), 0, nil, 100); !errors.Is(err, ErrRuntime) {
		t.Fatalf("Run error = %v; want ErrRuntime", err)
	}
}

func TestIntegerCompatSignedI1UsesInterpreterArithmetic(t *testing.T) {
	// bytecode_vm.c DEFINE_BINOP_BC_HELPER (lines 383-395) READ1-masks
	// operands before assigning them to int8_t. Signed i1 arithmetic and
	// comparisons therefore see 0/+1, unlike LLVM's usual 0/-1 meaning.
	for _, tc := range []struct {
		name       string
		op         int
		a, b, want uint64
	}{
		{"sdiv-true", 5, 1, 1, 1}, {"sdiv-false", 5, 0, 1, 0},
		{"srem-true", 7, 1, 1, 0}, {"srem-false", 7, 0, 1, 0},
		{"sgt", 27, 1, 0, 1}, {"sge", 28, 1, 0, 1},
		{"sle", 29, 1, 0, 0}, {"slt", 30, 0, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fields := []string{compatNum(0), compatNum(1)}
			if tc.op >= 21 {
				fields = append([]string{compatNum(1)}, fields...)
			}
			compatRun(t, tc.want, []uint64{tc.a, tc.b}, compatFunction{args: []int{1, 1}, locals: []int{1, 32}, result: 32, blocks: [][]string{{
				compatInst(1, 2, tc.op, fields...),
				compatInst(32, 3, 16, compatNum(2)),
				compatRet(32, compatNum(3)),
			}}})
		})
	}
}

func TestIntegerCompatI1Casts(t *testing.T) {
	// Cast source widths come from the local type table, not the encoded
	// constant's byte width (bytecode.c lines 1314-1331). SEXT has its own
	// i1 rule: READ1 true becomes ~0 (bytecode_vm.c lines 793-798), even
	// though signed arithmetic above interprets true as positive one.
	for _, width := range []int{8, 16, 32, 64} {
		for _, op := range []int{15, 16} {
			for _, value := range []uint64{0, 1, 2, 3} {
				t.Run(fmt.Sprintf("op-%d/to-i%d/%d", op, width, value), func(t *testing.T) {
					want := value & 1
					if op == 15 && want == 1 {
						want = ^uint64(0) >> uint(64-width)
					}
					compatRun(t, want, []uint64{value}, compatFunction{args: []int{1}, locals: []int{width}, result: width, blocks: [][]string{{
						compatInst(width, 1, op, compatNum(0)), compatRet(width, compatNum(1)),
					}}})
				})
			}
		}
	}
	for _, value := range []uint64{0, 1, 2, 3, 0xffffffff} {
		t.Run(fmt.Sprintf("trunc/%d", value), func(t *testing.T) {
			compatRun(t, value&1, []uint64{value}, compatFunction{args: []int{32}, locals: []int{1, 32}, result: 32, blocks: [][]string{{
				compatInst(1, 1, 14, compatNum(0)),
				compatInst(32, 2, 16, compatNum(1)), compatRet(32, compatNum(2)),
			}}})
		})
	}
	// Wider signed inputs must retain normal sign extension.
	compatRun(t, 0xffffff80, []uint64{0x80}, compatFunction{args: []int{8}, locals: []int{32}, result: 32, blocks: [][]string{{
		compatInst(32, 1, 15, compatNum(0)), compatRet(32, compatNum(1)),
	}}})
}

func TestIntegerCompatRejectsAmbiguousConstantCasts(t *testing.T) {
	// A direct constant supplies storage width, not the source semantic
	// type consulted by upstream parseBB for casts. Reject rather than
	// silently treating an encoded byte as either an i1 or an i8.
	for _, op := range []int{14, 15, 16} {
		width, storage := 32, 8
		if op == 14 {
			width, storage = 1, 32
		}
		data := compatProgram(compatFunction{locals: []int{width}, result: width, blocks: [][]string{{
			compatInst(width, 0, op, compatConstant(1, storage)), compatRet(width, compatNum(0)),
		}}})
		if _, err := Parse(data); !errors.Is(err, ErrUnsupported) {
			t.Errorf("opcode %d: Parse error = %v; want ErrUnsupported", op, err)
		}
	}
}

func TestIntegerCompatDoesNotTreatByteRegisterAsBoolean(t *testing.T) {
	// The byte-storage exception applies to constants only. A typed i8
	// register is not an i1 condition and needs an explicit conversion.
	data := compatProgram(compatFunction{args: []int{8}, result: 32, blocks: [][]string{
		{"T" + compatFixed(17, 2) + compatNum(0) + compatNum(1) + compatNum(2)},
		{compatRet(32, compatConstant(1, 32))},
		{compatRet(32, compatConstant(0, 32))},
	}})
	if _, err := Parse(data); !errors.Is(err, ErrMalformed) {
		t.Fatalf("Parse error = %v; want ErrMalformed", err)
	}
}
