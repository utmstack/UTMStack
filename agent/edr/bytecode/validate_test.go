package bytecode

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestNonBytePointerCopyMatchesReference(t *testing.T) {
	// The pinned parseBB derives interp_op = opcode*5 + size class and only
	// assigns a size class to types up to 65. A COPY of an i32* (type 67)
	// value therefore runs the one-bit handler upstream and the following
	// LOAD faults. The native engine must not accept it as a 64-bit copy.
	m := memoryFixture{globalMax: 0x8004, globals: []string{memoryGlobal(67, 0, 0x8004)}, functions: []compatFunction{{locals: []int{67, 32, 1}, result: 1, blocks: [][]string{{
		compatInst(0, 0, 34, memoryGlobalRef(1), compatNum(0)),
		compatInst(32, 1, 39, compatNum(0)),
		compatInst(1, 2, 21, compatNum(32), compatNum(1), compatConstant(6, 32)),
		compatRet(1, compatNum(2)),
	}}}}}
	p, err := Parse(m.encode())
	if !errors.Is(err, ErrUnsupported) {
		got, runErr := Result{}, error(nil)
		if err == nil {
			got, runErr = p.RunWithInput(context.Background(), 0, nil, []byte("xxABxx"), 1000)
		}
		t.Fatalf("Parse = %v (run %d, %v); want unsupported type-67 COPY", err, got.Value, runErr)
	}
}

func TestPointerInstructionWidthsFollowInterpreterSizeClasses(t *testing.T) {
	// Types 66-68 are i16*/i32*/i64*; 70 and 72 point to an array and a
	// struct. Only type 65 has the interpreter's 64-bit size class.
	types := append(memoryArrayTypes(4, 32), memoryType(3, 1, 32), memoryType(5, 1, 71))
	for _, typ := range []int{65, 66, 67, 68, 70, 72} {
		for name, body := range map[string][]string{
			"copy":         {compatInst(0, 0, 34, compatNum(1), compatNum(2))},
			"copy-global":  {compatInst(0, 0, 34, memoryGlobalRef(1), compatNum(2))},
			"load":         {compatInst(typ, 1, 39, compatNum(0))},
			"store":        {compatInst(0, 0, 38, compatNum(1), compatNum(0))},
			"store-global": {compatInst(typ, 0, 38, memoryGlobalRef(1), compatNum(0))},
			"select":       {compatInst(typ, 2, 31, compatNum(3), compatNum(1), compatNum(1))},
			"icmp-eq":      {compatInst(1, 3, 21, compatNum(uint64(typ)), compatNum(1), compatNum(2))},
			"icmp-ne":      {compatInst(1, 3, 22, compatNum(uint64(typ)), compatNum(1), compatNum(2))},
			"gep1":         {compatInst(typ, 1, 35, compatNum(67), compatNum(0), compatConstant(1, 32))},
		} {
			t.Run(fmt.Sprintf("%s-%d", name, typ), func(t *testing.T) {
				m := memoryFixture{types: types, globals: []string{memoryGlobal(typ, 0, 0)}, functions: []compatFunction{{
					locals: []int{65, typ, typ, 1}, result: 32,
					blocks: [][]string{append(append([]string(nil), body...), compatRet(32, compatConstant(0, 32)))},
				}}}
				var want error
				if typ != 65 {
					want = ErrUnsupported
				}
				if _, err := Parse(m.encode()); !errors.Is(err, want) {
					t.Fatalf("Parse = %v; want %v", err, want)
				}
			})
		}
		t.Run(fmt.Sprintf("result-%d", typ), func(t *testing.T) {
			m := memoryFixture{types: types, functions: []compatFunction{
				{locals: []int{typ}, result: 32, blocks: [][]string{{compatInst(typ, 0, 32, compatFixed(0, 1), compatNum(2)), compatRet(32, compatConstant(0, 32))}}},
				{locals: []int{typ}, result: typ, blocks: [][]string{{compatRet(typ, compatNum(0))}}},
			}}
			var want error
			if typ != 65 {
				want = ErrUnsupported
			}
			if _, err := Parse(m.encode()); !errors.Is(err, want) {
				t.Fatalf("Parse = %v; want %v", err, want)
			}
		})
	}
	// Width-independent pointer uses keep accepting other pointer types:
	// GEPZ results, arguments, and LOAD/STORE addresses read 64 bits.
	m := memoryFixture{types: types, functions: []compatFunction{
		{locals: []int{0x8045, 67, 32}, result: 32, blocks: [][]string{{
			compatInst(67, 1, 36, compatNum(70), compatNum(0), compatConstant(4, 32)),
			compatInst(32, 0, 38, compatConstant(41, 32), compatNum(1)),
			compatInst(32, 2, 32, compatFixed(1, 1), compatNum(2), compatNum(1)),
			compatRet(32, compatNum(2)),
		}}},
		{args: []int{67}, locals: []int{32}, result: 32, blocks: [][]string{{
			compatInst(32, 1, 39, compatNum(0)),
			compatInst(32, 1, 1, compatNum(1), compatConstant(1, 32)),
			compatRet(32, compatNum(1)),
		}}},
	}}
	if got, err := memoryParse(t, m).Run(context.Background(), 0, nil, 100); err != nil || got != 42 {
		t.Fatalf("typed pointer through GEPZ and argument: %d, %v", got, err)
	}
}

func TestDirectCallGlobalAndAddressedLocalArgumentsMatchReference(t *testing.T) {
	// Upstream maps direct-call operands with MAPPTR, which yields
	// 0x80000000|offset for a global and 0x40000000|offset for an addressed
	// local. OP_BC_CALL_DIRECT copies arguments with READOLD8/16/32/64, whose
	// CHECK_GT(func->numBytes, p) then fails with CL_EBYTECODE, so the
	// reference never runs such a call. Only constants and plain values work.
	types := memoryArrayTypes(4, 32)
	globals := []string{memoryGlobal(69, 1, 2, 3, 4), memoryGlobal(65, 0, 1), memoryGlobal(32, 5)}
	callee := func(arg int) compatFunction {
		return compatFunction{args: []int{arg}, locals: []int{32}, result: 32, blocks: [][]string{{compatRet(32, compatConstant(42, 32))}}}
	}
	for _, tc := range []struct {
		name     string
		param    int
		locals   []int
		setup    []string
		argument string
		want     error
	}{
		{"pointer-global", 65, nil, nil, memoryGlobalRef(2), ErrUnsupported},
		{"integer-global", 32, nil, nil, memoryGlobalRef(3), ErrUnsupported},
		{"null-global", 65, nil, nil, memoryGlobalRef(0), ErrUnsupported},
		{"addressed-local", 67, []int{0x8020}, []string{compatInst(0, 0, 34, compatConstant(7, 32), compatNum(0))}, compatNum(0), ErrUnsupported},
		{"addressed-array", 70, []int{0x8045}, nil, compatNum(0), ErrUnsupported},
		{"constant", 32, nil, nil, compatConstant(7, 32), nil},
		{"plain-integer", 32, []int{32}, []string{compatInst(0, 0, 34, compatConstant(7, 32), compatNum(0))}, compatNum(0), nil},
		{"plain-pointer", 67, []int{0x8045, 67}, []string{compatInst(67, 1, 36, compatNum(70), compatNum(0), compatConstant(0, 32))}, compatNum(1), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			locals := append(append([]int(nil), tc.locals...), 32)
			body := append(append([]string(nil), tc.setup...),
				compatInst(32, len(locals)-1, 32, compatFixed(1, 1), compatNum(2), tc.argument),
				compatRet(32, compatNum(uint64(len(locals)-1))))
			m := memoryFixture{types: types, globals: globals, functions: []compatFunction{
				{locals: locals, result: 32, blocks: [][]string{body}}, callee(tc.param),
			}}
			p, err := Parse(m.encode())
			if !errors.Is(err, tc.want) {
				t.Fatalf("Parse = %v; want %v", err, tc.want)
			}
			if err == nil {
				if got, err := p.Run(context.Background(), 0, nil, 100); err != nil || got != 42 {
					t.Fatalf("Run = %d, %v", got, err)
				}
			}
		})
	}
}
