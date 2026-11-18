package bytecode

import (
	"context"
	"encoding/binary"
	"errors"
	"testing"
)

func TestPERecordCarriesTheFourFieldsTheReferenceSets(t *testing.T) {
	r := peRecord(&PEInfo{EntryPoint: 0x400, HeaderSize: 0x1000, Sections: make([]Section, 3)})
	if len(r) != 648 {
		t.Fatalf("record is %d bytes, want 648", len(r))
	}
	le := binary.LittleEndian
	if le.Uint32(r[0:]) != 0 || le.Uint32(r[4:]) != 0x400 || le.Uint16(r[8:]) != 3 || le.Uint32(r[644:]) != 0x1000 {
		t.Fatalf("fields %x %x %x %x", le.Uint32(r[0:]), le.Uint32(r[4:]), le.Uint16(r[8:]), le.Uint32(r[644:]))
	}
	for i, b := range r[10:644] {
		if b != 0 {
			t.Fatalf("byte %d is %#x, want 0", 10+i, b)
		}
	}
	if z := peRecord(nil); len(z) != 648 || le.Uint32(z[4:]) != 0 || le.Uint32(z[644:]) != 0 {
		t.Fatal("a non-PE object gets an all-zero record")
	}
}

// peDataProgram loads 32 bits at offset of the header record (__clambc_pedata).
func peDataProgram(offset uint64) memoryFixture {
	return memoryFixture{globalMax: 0x8003, globals: []string{memoryGlobal(67, offset, 0x8003)}, functions: []compatFunction{{locals: []int{32}, result: 32, blocks: [][]string{{
		compatInst(32, 0, 39, memoryGlobalRef(1)), compatRet(32, compatNum(0)),
	}}}}}
}

func TestPEDataGlobalReadsTheRecord(t *testing.T) {
	pe := &PEInfo{EntryPoint: 0x1234, HeaderSize: 0x400, Sections: make([]Section, 5)}
	for offset, want := range map[uint64]uint64{4: 0x1234, 8: 5, 644: 0x400, 36: 0} {
		p := memoryParse(t, peDataProgram(offset))
		matches := EmptyMatchContext()
		matches.PE = pe
		got, err := p.RunWithMatchContext(context.Background(), 0, nil, []byte("MZ"), matches, 100)
		if err != nil || got.Value != want {
			t.Fatalf("offset %d: %+v, %v; want %#x", offset, got, err, want)
		}
		got, err = p.RunWithMatchContext(context.Background(), 0, nil, []byte("MZ"), EmptyMatchContext(), 100)
		if err != nil || got.Value != 0 {
			t.Fatalf("offset %d without a PE: %+v, %v; want 0", offset, got, err)
		}
		if got, err = p.RunWithInput(context.Background(), 0, nil, []byte("MZ"), 100); err != nil || got.Value != 0 {
			t.Fatalf("offset %d without match context: %+v, %v; want 0", offset, got, err)
		}
	}
}

func TestVirusNamesPointerLoadsButFaultsWhenUsed(t *testing.T) {
	// The reference never registers the detection-name table, so a pointer
	// to it has a value but any access through it fails.
	for _, offset := range []uint64{0, 3, 1 << 20} {
		m := memoryFixture{globalMax: 0x8002, globals: []string{memoryGlobal(67, offset, 0x8002)}, functions: []compatFunction{{locals: []int{32}, result: 32, blocks: [][]string{{
			compatRet(32, compatConstant(7, 32)),
		}}}}}
		if got, err := memoryParse(t, m).RunWithInput(context.Background(), 0, nil, nil, 100); err != nil || got.Value != 7 {
			t.Fatalf("offset %d: program with the pointer does not run: %+v, %v", offset, got, err)
		}
		m.functions[0].blocks = [][]string{{compatInst(32, 0, 39, memoryGlobalRef(1)), compatRet(32, compatNum(0))}}
		if _, err := memoryParse(t, m).RunWithInput(context.Background(), 0, nil, nil, 100); !errors.Is(err, ErrRuntime) {
			t.Fatalf("offset %d: load through the pointer: %v; want ErrRuntime", offset, err)
		}
	}
}
