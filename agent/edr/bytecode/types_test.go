package bytecode

import (
	"errors"
	"math"
	"testing"
)

func typeEncoding(kind int, count int, elements ...int) string {
	s := fixed(uint64(kind), 1)
	if kind != pointerType {
		s += num(uint64(count))
	}
	for _, id := range elements {
		s += num(uint64(id))
	}
	return s
}

func typeFixture(encoded ...string) []byte {
	s := "T" + fixed(firstDynamicType, 2)
	for _, ty := range encoded {
		s += ty
	}
	return []byte(s)
}

func parseTestTypes(t *testing.T, encoded ...string) *typeTable {
	t.Helper()
	table, err := parseTypeTable(typeFixture(encoded...), len(encoded)+5)
	if err != nil {
		t.Fatal(err)
	}
	return table
}

func TestTypeTablePrimitiveAndStaticPointerLayout(t *testing.T) {
	table := parseTestTypes(t)
	for id := 0; id <= 64; id++ {
		want := 8
		switch {
		case id == 0:
			want = 0
		case id <= 8:
			want = 1
		case id <= 16:
			want = 2
		case id <= 32:
			want = 4
		}
		if got, err := table.size(id); err != nil || got != want {
			t.Fatalf("type %d: size %d, error %v; want %d", id, got, err, want)
		}
		if table.pointer(id) {
			t.Fatalf("integer type %d identified as pointer", id)
		}
		if want == 0 {
			want = 1
		}
		if got, err := table.align(id); err != nil || got != want {
			t.Fatalf("type %d: alignment %d, error %v; want %d", id, got, err, want)
		}
	}
	for i, child := range []int{8, 16, 32, 64} {
		id := firstCompoundType + i
		if !table.pointer(id) || !table.pointer(id|0x8000) {
			t.Fatalf("type %d is not a pointer", id)
		}
		if size, err := table.size(id | 0x8000); err != nil || size != 8 {
			t.Fatalf("pointer size %d, error %v", size, err)
		}
		next, offset, err := table.element(id, -3)
		if err != nil || next != child || offset != -(3<<i) {
			t.Fatalf("pointer %d element %d offset %d error %v", id, next, offset, err)
		}
	}
	for _, invalid := range []int{-1, 69, 70, 4096, 65536, math.MaxInt} {
		if _, err := table.size(invalid); !errors.Is(err, ErrMalformed) || table.pointer(invalid) {
			t.Fatalf("accepted undefined type %d: %v", invalid, err)
		}
	}
}

func TestTypeTableInterpreterStructLayoutAndGEP(t *testing.T) {
	for _, kind := range []int{packedStructType, structType} {
		table := parseTestTypes(t,
			typeEncoding(kind, 4, 8, 32, 16, 68),
			typeEncoding(arrayType, 3, 69),
			typeEncoding(pointerType, 1, 70),
		)
		if size, err := table.size(69); err != nil || size != 15 {
			t.Fatal("struct must use interpreter's unpadded ABI", size, err)
		}
		if align, err := table.align(69); err != nil || align != 8 {
			t.Fatal(align, err)
		}
		if size, err := table.size(70); err != nil || size != 45 {
			t.Fatal(size, err)
		}
		for i, want := range []int{0, 1, 5, 7} {
			next, offset, err := table.element(69, i)
			if err != nil || next != table.types[4].elements[i] || offset != want {
				t.Fatal(next, offset, err)
			}
		}
		for _, index := range []int{-2, 0, 2, 3, 4} {
			next, offset, err := table.element(70, index)
			if err != nil || next != 69 || offset != index*15 {
				t.Fatal("array GEP", next, offset, err)
			}
		}
		next, offset, err := table.element(71, 2)
		if err != nil || next != 70 || offset != 90 {
			t.Fatal("pointer GEP", next, offset, err)
		}
		for _, index := range []int{-1, 4} {
			if _, _, err := table.element(69, index); !errors.Is(err, ErrMalformed) {
				t.Fatal("out of range struct GEP", index, err)
			}
		}
		for _, index := range []int{math.MinInt, math.MaxInt, maxTypeBytes} {
			if _, _, err := table.element(71, index); !errors.Is(err, ErrLimit) {
				t.Fatal("unbounded GEP", index, err)
			}
		}
	}
}

func TestTypeTableUpstreamAPIShapes(t *testing.T) {
	// These are independently encoded layouts/prototypes from the public
	// bytecode_api_decl.c table: cli_exe_section is nine i32 fields; the read
	// API returns i32 and takes i8*, i32. DISASM_RESULT has three byte/word
	// fields followed by byte arrays, totaling 64 bytes in the interpreter.
	table := parseTestTypes(t,
		typeEncoding(structType, 9, 32, 32, 32, 32, 32, 32, 32, 32, 32), // 69
		typeEncoding(functionType, 3, 32, 65, 32),                       // 70
		typeEncoding(pointerType, 1, 70),                                // 71
		typeEncoding(arrayType, 29, 8),                                  // 72
		typeEncoding(arrayType, 10, 8),                                  // 73
		typeEncoding(arrayType, 3, 73),                                  // 74
		typeEncoding(structType, 6, 16, 8, 8, 8, 72, 74),                // 75
	)
	for id, want := range map[int]int{69: 36, 70: 8, 71: 8, 72: 29, 73: 10, 74: 30, 75: 64} {
		if got, err := table.size(id); err != nil || got != want {
			t.Fatalf("API type %d: size %d, error %v; want %d", id, got, err, want)
		}
	}
	prototype := table.types[70-firstCompoundType]
	if prototype.kind != functionType || prototype.count != 3 || prototype.elements[0] != 32 || prototype.elements[1] != 65 || prototype.elements[2] != 32 {
		t.Fatalf("incorrect API prototype: %+v", prototype)
	}
	if _, _, err := table.element(71, 1); !errors.Is(err, ErrUnsupported) {
		t.Fatal("GEP over function is unsupported", err)
	}
}

func TestTypeTableForwardReferencesAndPointerCycles(t *testing.T) {
	table := parseTestTypes(t,
		typeEncoding(arrayType, 4, 71),      // forward aggregate
		typeEncoding(pointerType, 1, 71),    // pointer back to containing struct
		typeEncoding(structType, 2, 32, 70), // pointer ends layout recursion
		typeEncoding(pointerType, 1, 72),    // recursive pointer is finite-sized
		typeEncoding(structType, 0),         // empty structs are valid
		typeEncoding(arrayType, 0, 32),      // zero arrays are valid
		typeEncoding(functionType, 1, 0),    // void(void)
		typeEncoding(pointerType, 1, 0),     // opaque void pointer
	)
	for id, want := range map[int]int{69: 48, 70: 8, 71: 12, 72: 8, 73: 0, 74: 0, 75: 8, 76: 8} {
		if got, err := table.size(id); err != nil || got != want {
			t.Fatalf("type %d: size %d, error %v; want %d", id, got, err, want)
		}
	}
	if _, _, err := table.element(76, 1); !errors.Is(err, ErrUnsupported) {
		t.Fatal("opaque pointer GEP must fail", err)
	}
	for _, types := range [][]string{
		{typeEncoding(structType, 1, 69)},
		{typeEncoding(arrayType, 0, 69)},
		{typeEncoding(arrayType, 2, 70), typeEncoding(structType, 1, 69)},
		{typeEncoding(structType, 1, 0)},
		{typeEncoding(arrayType, 1, 0)},
		{typeEncoding(functionType, 2, 32, 0)},
		{typeEncoding(functionType, 1, 69)},
		{typeEncoding(structType, 1, 70), typeEncoding(functionType, 1, 0)},
	} {
		if _, err := parseTypeTable(typeFixture(types...), len(types)+5); !errors.Is(err, ErrMalformed) {
			t.Fatal("accepted invalid graph", types, err)
		}
	}
}

func TestTypeTableMalformedAndResourceBounds(t *testing.T) {
	for _, tc := range []struct {
		line  []byte
		count int
		want  error
	}{
		{nil, 5, ErrMalformed},
		{typeFixture(), 4, ErrMalformed},
		{typeFixture(), maxTypes + 1, ErrLimit},
		{[]byte("T" + fixed(68, 2)), 5, ErrUnsupported},
		{typeFixture(), 6, ErrMalformed},
		{typeFixture(typeEncoding(0, 0)), 6, ErrMalformed},
		{typeFixture(typeEncoding(functionType, 0)), 6, ErrMalformed},
		{typeFixture(typeEncoding(pointerType, 1, 70)), 6, ErrMalformed},
		{typeFixture(typeEncoding(pointerType, 1, 0x8008)), 6, ErrMalformed},
		{typeFixture(typeEncoding(arrayType, maxTypeBytes, 64)), 6, ErrLimit},
		{typeFixture(typeEncoding(arrayType, maxTypeBytes+1, 8)), 6, ErrLimit},
		{typeFixture(typeEncoding(structType, maxTypeMembers+1)), 6, ErrLimit},
	} {
		if _, err := parseTypeTable(tc.line, tc.count); !errors.Is(err, tc.want) {
			t.Fatalf("line %q count %d: error %v, want %v", tc.line, tc.count, err, tc.want)
		}
	}
	valid := typeFixture(typeEncoding(structType, 3, 8, 32, 64))
	for i := 0; i < len(valid); i++ {
		if _, err := parseTypeTable(valid[:i], 6); err == nil {
			t.Fatalf("accepted truncation %d", i)
		}
	}
	if _, err := parseTypeTable(append(valid, '`'), 6); !errors.Is(err, ErrMalformed) {
		t.Fatal("accepted trailing byte", err)
	}
	var nested []string
	for i := 0; i < maxTypeDepth+1; i++ {
		child := firstDynamicType + i + 1
		if i == maxTypeDepth {
			child = 8
		}
		nested = append(nested, typeEncoding(arrayType, 1, child))
	}
	if _, err := parseTypeTable(typeFixture(nested...), len(nested)+5); !errors.Is(err, ErrLimit) {
		t.Fatal("unbounded nesting", err)
	}
	var wide []string
	fields := make([]int, maxTypeMembers)
	for i := range fields {
		fields[i] = 8
	}
	for i := 0; i < maxTypeReferences/maxTypeMembers+1; i++ {
		wide = append(wide, typeEncoding(structType, len(fields), fields...))
	}
	if _, err := parseTypeTable(typeFixture(wide...), len(wide)+5); !errors.Is(err, ErrLimit) {
		t.Fatal("unbounded total references", err)
	}
	// Adding two individually permitted arrays must not overflow the total.
	oversized := []string{typeEncoding(arrayType, maxTypeBytes, 8), typeEncoding(structType, 2, 69, 8)}
	if _, err := parseTypeTable(typeFixture(oversized...), 7); !errors.Is(err, ErrLimit) {
		t.Fatal("unbounded structure size", err)
	}
}

func FuzzTypeTable(f *testing.F) {
	f.Add(typeFixture(), 5)
	f.Add(typeFixture(typeEncoding(structType, 3, 8, 32, 64)), 6)
	f.Add(typeFixture(typeEncoding(pointerType, 1, 70), typeEncoding(structType, 1, 69)), 7)
	f.Fuzz(func(t *testing.T, data []byte, count int) {
		if len(data) > 65536 {
			return
		}
		table, err := parseTypeTable(data, count)
		if err != nil {
			return
		}
		for id := 0; id < firstCompoundType+len(table.types); id++ {
			if size, err := table.size(id); err != nil || size < 0 || size > maxTypeBytes {
				t.Fatalf("invalid computed size: %d %v", size, err)
			}
			_, _ = table.align(id)
			_ = table.pointer(id)
			_, _, _ = table.element(id, 0)
			_, _, _ = table.element(id, math.MaxInt)
			_, _, _ = table.element(id, math.MinInt)
		}
	})
}
