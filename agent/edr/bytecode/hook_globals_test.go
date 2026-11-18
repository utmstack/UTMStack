package bytecode

import (
	"errors"
	"math"
	"testing"
)

func hookGlobalFixture(declared uint64, entries ...string) []byte {
	base := globalFixture(entries...)
	return append([]byte("G"+num(declared)), base[2:]...)
}

func TestSpecialGlobalFixedLayouts(t *testing.T) {
	// Pin the public bytecode_api.h ABI: uint32[64] counts and offsets,
	// uint16 kind, uint32[1] file size and the 648-byte PE header record.
	// These sizes never depend on input.
	for id, want := range map[int]int{0x8000: 256, 0x8001: 2, 0x8003: 648, 0x8004: 4, 0x8005: 256} {
		if got, known := specialGlobalSize(id); !known || got != want {
			t.Fatalf("global %#x: size %d, known %v; want %d", id, got, known, want)
		}
	}
	// The detection-name table has no size: the reference never registers it.
	for _, id := range []int{-1, 0, 0x7fff, 0x8002, 0x8006, 0xffff, math.MaxInt} {
		if size, known := specialGlobalSize(id); known || size != 0 {
			t.Fatalf("accepted unknown global %#x: %d %v", id, size, known)
		}
	}
}

func TestSpecialGlobalReferencesAndDeclarations(t *testing.T) {
	table := parseTestTypes(t)
	ids := []int{globalMatchCounts, globalKind, globalPEData, globalFileSize, globalMatchOffsets}
	for _, declared := range ids {
		if _, err := parseGlobals(hookGlobalFixture(uint64(declared)), table); err != nil {
			t.Fatalf("unused declaration %#x: %v", declared, err)
		}
		for _, target := range ids {
			size, _ := specialGlobalSize(target)
			for _, offset := range []int{0, 1, size - 1, size} {
				globals, err := parseGlobals(hookGlobalFixture(uint64(declared), num(65)+globalInitializer(uint64(offset), uint64(target))), table)
				if target > declared {
					if !errors.Is(err, ErrMalformed) {
						t.Fatalf("target %#x exceeds declaration %#x: %v", target, declared, err)
					}
					continue
				}
				if err != nil || len(globals) != 2 || len(globals[1].pointers) != 1 {
					t.Fatalf("declaration %#x target %#x offset %d: %v", declared, target, offset, err)
				}
				if got := globals[1].pointers[0]; got != (globalPointer{target: target, targetOffset: uint64(offset)}) {
					t.Fatalf("incorrect fixup: %+v", got)
				}
			}
		}
	}
	for _, target := range ids {
		size, _ := specialGlobalSize(target)
		for _, offset := range []uint64{uint64(size) + 1, math.MaxUint64} {
			if _, err := parseGlobals(hookGlobalFixture(globalMatchOffsets, num(65)+globalInitializer(offset, uint64(target))), table); !errors.Is(err, ErrMalformed) {
				t.Fatalf("oversized offset %d into %#x: %v", offset, target, err)
			}
		}
		if _, err := parseGlobals(hookGlobalFixture(0, num(65)+globalInitializer(0, uint64(target))), table); !errors.Is(err, ErrMalformed) {
			t.Fatalf("undeclared target %#x: %v", target, err)
		}
	}
	// A pointer to the detection-name table composes with any offset.
	for _, offset := range []uint64{0, 1, 1 << 20} {
		if _, err := parseGlobals(hookGlobalFixture(globalVirusNames, num(65)+globalInitializer(offset, globalVirusNames)), table); err != nil {
			t.Fatalf("detection-name pointer offset %d: %v", offset, err)
		}
	}
	for _, unsupported := range []uint64{1, 0x7fff, 0x8006, 0x100008000, math.MaxUint64} {
		if _, err := parseGlobals(hookGlobalFixture(unsupported), table); !errors.Is(err, ErrUnsupported) {
			t.Fatalf("unsupported declaration %#x: %v", unsupported, err)
		}
		if unsupported >= 0x8000 {
			if _, err := parseGlobals(hookGlobalFixture(globalMatchOffsets, num(65)+globalInitializer(0, unsupported)), table); !errors.Is(err, ErrUnsupported) {
				t.Fatalf("unsupported target %#x: %v", unsupported, err)
			}
		}
	}
}

func TestProgramNeedsRealMatchContext(t *testing.T) {
	table := parseTestTypes(t, typeEncoding(structType, 2, 65, 65))
	for _, tc := range []struct {
		target int
		want   bool
	}{
		{0, false}, {globalFileSize, false}, {globalKind, false},
		{globalMatchCounts, true}, {globalMatchOffsets, true},
	} {
		// Put the tested fixup after a normal file-size pointer so the
		// predicate must inspect every global, not only the first.
		globals, err := parseGlobals(hookGlobalFixture(globalMatchOffsets,
			num(65)+globalInitializer(0, globalFileSize), num(65)+globalInitializer(0, uint64(tc.target))), table)
		if err != nil {
			t.Fatal(err)
		}
		p := &Program{globals: globals}
		if got := p.NeedsMatchContext(); got != tc.want {
			t.Fatalf("target %#x needs match context %v; want %v", tc.target, got, tc.want)
		}
	}
	// The interpreter leaves struct globals zero, so pointers nested in a
	// struct initializer bind nothing and need no match context.
	globals, err := parseGlobals(hookGlobalFixture(globalMatchOffsets,
		num(69)+globalInitializer(0, globalMatchCounts, 0, globalMatchOffsets)), table)
	if err != nil || (&Program{globals: globals}).NeedsMatchContext() || len(globals[1].pointers) != 0 {
		t.Fatal("struct initializer bound match context", err)
	}
	globals, err = parseGlobals(hookGlobalFixture(globalMatchOffsets), table)
	if err != nil || (&Program{globals: globals}).NeedsMatchContext() {
		t.Fatal("unused declaration must not require context", err)
	}
	if (&Program{}).NeedsMatchContext() || (*Program)(nil).NeedsMatchContext() {
		t.Fatal("empty program requires context")
	}
}
