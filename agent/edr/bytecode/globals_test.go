package bytecode

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"testing"
)

func globalInitializer(values ...uint64) string {
	s := ""
	for _, value := range values {
		encoded := []byte(num(value))
		encoded[0] &^= 0x20
		s += string(encoded)
	}
	return s + "`"
}

func globalFixture(entries ...string) []byte {
	s := "G" + num(0) + num(uint64(len(entries)+1)) + num(0) + "@`"
	for _, entry := range entries {
		s += entry
	}
	return []byte(s)
}

func TestGlobalScalarInitializersStayZero(t *testing.T) {
	// cli_bytecode_prepare_interpreter skips globals whose type ID is below
	// 65: their initializers are parsed and validated, but the values the
	// interpreter reads remain zero.
	table := parseTestTypes(t)
	var entries []string
	widths := []int{1, 8, 9, 16, 17, 32, 33, 64}
	for _, width := range widths {
		entries = append(entries, num(uint64(width))+globalInitializer(math.MaxUint64))
	}
	globals, err := parseGlobals(globalFixture(entries...), table)
	if err != nil {
		t.Fatal(err)
	}
	if len(globals[0].data) != 0 || globals[0].typeID != 0 || len(globals[0].pointers) != 0 || globals[0].unresolved {
		t.Fatal("reserved null object", globals[0])
	}
	for i, width := range widths {
		g := globals[i+1]
		if !bytes.Equal(g.data, make([]byte, integerSize(width))) || g.typeID != width || len(g.pointers) != 0 || g.unresolved {
			t.Fatalf("width %d initializer %x; want zero", width, g.data)
		}
	}
	zero, err := parseGlobals(globalFixture(num(64)+"@`"), table)
	if err != nil || !bytes.Equal(zero[1].data, make([]byte, 8)) {
		t.Fatal(zero, err)
	}
	if globals, err := parseGlobals([]byte("G"+num(0)+num(0)), table); err != nil || len(globals) != 0 {
		t.Fatal("empty globals table", globals, err)
	}
}

func TestGlobalAggregateInitializersAndForwardPointers(t *testing.T) {
	for _, kind := range []int{packedStructType, structType} {
		table := parseTestTypes(t,
			typeEncoding(kind, 3, 8, 65, 16),
			typeEncoding(arrayType, 2, 69),
			typeEncoding(arrayType, 4, 8),
			typeEncoding(arrayType, 2, 65),
		)
		// The interpreter leaves structs and arrays of 11-byte structs zero,
		// including their nested pointers. It copies raw components into
		// byte arrays and pointer arrays, and composes top-level pointers.
		encoded := globalFixture(
			num(70)+globalInitializer(0x12, 1, 2, 0x3456, 0x78, 4, 2, 0x9abc),
			num(71)+globalInitializer(0x1a1, 0xb2, 0xc3, 0xd4),
			num(65)+globalInitializer(11, 1), // pointer into the array's second struct
			num(65)+globalInitializer(0, 4),  // self reference
			num(65)+"@`",                     // compact zero pointer
			num(65)+globalInitializer(0, 0),  // explicit (0, 0) pointer
			num(72)+globalInitializer(3, 2, 0, 0),
			num(69)+globalInitializer(0x12, 3, 7, 0x3456),
		)
		globals, err := parseGlobals(encoded, table)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(globals[1].data, make([]byte, 22)) || len(globals[1].pointers) != 0 {
			t.Fatalf("array of structs: %x %+v", globals[1].data, globals[1].pointers)
		}
		if !bytes.Equal(globals[2].data, []byte{0xa1, 0xb2, 0xc3, 0xd4}) {
			t.Fatalf("byte array: %x", globals[2].data)
		}
		if len(globals[3].pointers) != 1 || globals[3].pointers[0] != (globalPointer{target: 1, targetOffset: 11}) {
			t.Fatal(globals[3].pointers)
		}
		if len(globals[4].pointers) != 1 || globals[4].pointers[0].target != 4 || globals[4].unresolved {
			t.Fatal("self pointer", globals[4])
		}
		for _, g := range globals[5:7] {
			if len(g.pointers) != 0 || !g.unresolved || !bytes.Equal(g.data, make([]byte, 8)) {
				t.Fatal("zero pointer global must stay unreadable", g)
			}
		}
		pointerArray := make([]byte, 16)
		binary.LittleEndian.PutUint64(pointerArray, 3)
		binary.LittleEndian.PutUint64(pointerArray[8:], 2)
		if !bytes.Equal(globals[7].data, pointerArray) || len(globals[7].pointers) != 0 {
			t.Fatalf("pointer array: %x %+v", globals[7].data, globals[7].pointers)
		}
		if !bytes.Equal(globals[8].data, make([]byte, 11)) || len(globals[8].pointers) != 0 || globals[8].unresolved {
			t.Fatalf("struct: %x %+v", globals[8].data, globals[8].pointers)
		}
	}
}

func TestGlobalsMalformedAndUnsupported(t *testing.T) {
	table := parseTestTypes(t,
		typeEncoding(arrayType, 2, 8),
		typeEncoding(functionType, 1, 32),
	)
	for _, tc := range []struct {
		line []byte
		want error
	}{
		{nil, ErrMalformed},
		{[]byte("G" + num(1) + num(0)), ErrUnsupported},
		{[]byte("G" + num(0) + num(maxGlobals+1)), ErrLimit},
		{[]byte("G" + num(0) + num(1) + num(32) + "@`"), ErrMalformed},
		{[]byte("G" + num(0) + num(1) + num(0) + globalInitializer(1)), ErrMalformed},
		{globalFixture(num(0) + "@`"), ErrMalformed},
		{globalFixture(num(71) + "@`"), ErrMalformed},
		{globalFixture(num(0x8008) + "@`"), ErrMalformed},
		{globalFixture(num(70) + "@`"), ErrUnsupported},
		{globalFixture(num(69) + globalInitializer(1)), ErrMalformed},
		{globalFixture(num(69) + globalInitializer(1, 2, 3)), ErrMalformed},
		{globalFixture(num(8) + num(1) + "`"), ErrMalformed}, // numeric references aren't constants
		{globalFixture(num(65) + globalInitializer(0, 2)), ErrMalformed},
		{globalFixture(num(65) + globalInitializer(1, 0)), ErrMalformed},
		{globalFixture(num(65) + globalInitializer(9, 1)), ErrMalformed},
		{globalFixture(num(65) + globalInitializer(math.MaxUint64, 1)), ErrMalformed},
		{globalFixture(num(65) + globalInitializer(0, 0x8000)), ErrMalformed},
		{globalFixture(num(65) + globalInitializer(0, math.MaxUint64)), ErrUnsupported},
	} {
		if _, err := parseGlobals(tc.line, table); !errors.Is(err, tc.want) {
			t.Fatalf("line %q: error %v, want %v", tc.line, err, tc.want)
		}
	}
	valid := globalFixture(num(69)+globalInitializer(1, 2), num(65)+globalInitializer(2, 1))
	for i := 0; i < len(valid); i++ {
		if _, err := parseGlobals(valid[:i], table); err == nil {
			t.Fatalf("accepted truncation %d", i)
		}
	}
	if _, err := parseGlobals(append(valid, '`'), table); !errors.Is(err, ErrMalformed) {
		t.Fatal("accepted trailing garbage", err)
	}
	if _, err := parseGlobals(valid, nil); !errors.Is(err, ErrMalformed) {
		t.Fatal("accepted missing types", err)
	}
}

func TestGlobalComponentBudgetsAndEmptyAggregates(t *testing.T) {
	table := parseTestTypes(t,
		typeEncoding(arrayType, maxGlobalComponents, 8),
		typeEncoding(structType, 0),
		typeEncoding(arrayType, maxTypeBytes, 70),
		typeEncoding(arrayType, maxGlobalComponents/2, 8),
	)
	// Even compact zero initializers must respect the total component budget;
	// the reserved null global itself uses one component.
	if _, err := parseGlobals(globalFixture(num(69)+"@`"), table); !errors.Is(err, ErrLimit) {
		t.Fatal("unbounded zero initializer", err)
	}
	if _, err := parseGlobals(globalFixture(num(72)+"@`", num(72)+"@`"), table); !errors.Is(err, ErrLimit) {
		t.Fatal("unbounded total components", err)
	}
	globals, err := parseGlobals(globalFixture(num(71)+"`", num(70)+"@`"), table)
	if err != nil || len(globals[1].data) != 0 || len(globals[2].data) != 0 {
		t.Fatal("empty aggregates", globals, err)
	}
	// A compact DAG of empty structs can describe exponentially many empty
	// values. Its non-zero-init spelling must remain constant in materialized
	// size rather than expanding those paths recursively.
	var encoded []string
	for i := 0; i < 64; i++ {
		if i == 63 {
			encoded = append(encoded, typeEncoding(structType, 0))
		} else {
			encoded = append(encoded, typeEncoding(structType, 2, 70+i, 70+i))
		}
	}
	empty := parseTestTypes(t, encoded...)
	if _, err := parseGlobals(globalFixture(num(69)+"`"), empty); err != nil {
		t.Fatal("empty DAG", err)
	}
	// Empty siblings around nonempty fields do not multiply initialization
	// work: arrays copy at most one raw component per element and structs
	// are not walked, so the cost is bounded by the component budget.
	fields := make([]int, 257)
	for i := range fields {
		fields[i] = 69
	}
	fields[len(fields)-1] = 8
	wide := parseTestTypes(t, typeEncoding(structType, 0), typeEncoding(structType, len(fields), fields...), typeEncoding(arrayType, 5000, 70))
	values := make([]uint64, 5000)
	for i := range values {
		values[i] = 1
	}
	globals, err = parseGlobals(globalFixture(num(71)+globalInitializer(values...)), wide)
	if err != nil || !bytes.Equal(globals[1].data, bytes.Repeat([]byte{1}, 5000)) {
		t.Fatal("array of one-byte structs", err)
	}
}

func TestGlobalFileSizeDeclaredSubset(t *testing.T) {
	table := parseTestTypes(t)
	declare := func(max uint64, entries ...string) []byte {
		base := globalFixture(entries...)
		return append([]byte("G"+num(max)), base[2:]...)
	}
	for _, offset := range []uint64{0, 1, 3, 4} {
		globals, err := parseGlobals(declare(globalFileSize, num(67)+globalInitializer(offset, globalFileSize)), table)
		if err != nil || len(globals) != 2 || len(globals[1].pointers) != 1 {
			t.Fatal(offset, globals, err)
		}
		if got := globals[1].pointers[0]; got != (globalPointer{target: globalFileSize, targetOffset: offset}) {
			t.Fatalf("file-size fixup: %+v", got)
		}
	}
	// Declaring the capability does not require a reference to it.
	if _, err := parseGlobals(declare(globalFileSize), table); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		max, target, offset uint64
		want                error
	}{
		{0, globalFileSize, 0, ErrMalformed},
		{globalFileSize, globalFileSize, 5, ErrMalformed},
		{globalFileSize, globalFileSize, math.MaxUint64, ErrMalformed},
		// The detection-name table and the PE header record are declarable
		// (the reference refuses only globals above the highest it knows).
		{0x8002, 0, 0, nil},
		{0x8003, 0, 0, nil},
		{globalFileSize, 0x8002, 0, nil},
		{globalFileSize, 0x8003, 0, nil},
		{globalFileSize, 0x8005, 0, ErrMalformed},
		{globalFileSize, math.MaxUint64, 0, ErrUnsupported},
		{globalFileSize, 2, 0, ErrMalformed},
	} {
		if _, err := parseGlobals(declare(tc.max, num(67)+globalInitializer(tc.offset, tc.target)), table); !errors.Is(err, tc.want) {
			t.Fatalf("declared %#x target %#x offset %d: got %v, want %v", tc.max, tc.target, tc.offset, err, tc.want)
		}
	}
}

func FuzzGlobalInitializers(f *testing.F) {
	table, err := parseTypeTable(typeFixture(typeEncoding(structType, 3, 8, 65, 64), typeEncoding(arrayType, 4, 8)), 7)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(globalFixture())
	f.Add(globalFixture(num(69)+globalInitializer(1, 0, 2, math.MaxUint64), num(70)+globalInitializer(1, 2, 3, 4)))
	fileSizeGlobal := globalFixture(num(67) + globalInitializer(0, globalFileSize))
	f.Add(append([]byte("G"+num(globalFileSize)), fileSizeGlobal[2:]...))
	for _, id := range []uint64{globalMatchCounts, globalKind, globalMatchOffsets} {
		g := globalFixture(num(67) + globalInitializer(0, id))
		f.Add(append([]byte("G"+num(id)), g[2:]...))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 65536 {
			return
		}
		globals, err := parseGlobals(data, table)
		if err != nil {
			return
		}
		total := 0
		for _, g := range globals {
			total += len(g.data)
			for _, p := range g.pointers {
				if p.offset < 0 || p.offset > len(g.data)-8 || p.target < 0 {
					t.Fatalf("invalid pointer %+v", p)
				}
				if size, known := specialGlobalSize(p.target); known {
					if p.targetOffset > uint64(size) {
						t.Fatalf("invalid special-global pointer %+v", p)
					}
					continue
				}
				if p.target >= len(globals) || p.targetOffset > uint64(len(globals[p.target].data)) {
					t.Fatalf("invalid pointer %+v", p)
				}
			}
		}
		if total > maxGlobalBytes {
			t.Fatal("unbounded allocation", total)
		}
	})
}
