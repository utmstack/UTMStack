package signatures

import (
	"encoding/binary"
	"errors"
	"testing"
)

// headerReader is a logical program triggered by "AB" that reads the
// executable header (through pe_rawaddr, or by declaring a pointer to the PE
// header record) and returns without a detection.
func headerReader(target string, record bool) callbackFixture {
	f := callbackFixture{
		kind: 256, logical: "Header.Reader;Engine:51-255,Target:" + target + ";0;4142", result: 32,
		locals: []int{32}, blocks: [][]string{{callbackReturn(32, 0)}},
	}
	if record {
		f.types = []string{callbackType(4, 8, 8)}
		f.globalMax, f.globals = 0x8003, []string{callbackGlobal(65, 0, 0x8003)}
		return f
	}
	f.types = []string{callbackType(1, 2, 32, 32)}
	f.apiMax, f.apis = 15, []string{callbackAPI(15, 69, "pe_rawaddr")}
	f.blocks = [][]string{{callbackCall(0, 15, callbackConstant(0x1000, 32)), callbackReturn(32, 0)}}
	return f
}

// A program that reads the executable header can run after the reference
// rejects an unmapped PE entry point. Parsed ELF and Mach-O header data remain
// incomplete because the reference fills fields this engine leaves zero.
func TestHeaderReadingBytecodeNeedsAModeledHeader(t *testing.T) {
	for _, record := range []bool{false, true} {
		d := compileCallback(t, headerReader("1", record))
		name, err := d.Match(nil2ctx(), brokenPEWith("AB"))
		if name != "" || err != nil {
			t.Fatalf("record %v, rejected PE header: (%q, %v), want clean", record, name, err)
		}
		pe := testPE()
		copy(pe[600:], "AB")
		if name, err := d.Match(nil2ctx(), pe); name != "" || err != nil {
			t.Fatalf("record %v, parsed PE: (%q, %v), want clean", record, name, err)
		}
		if name, err := d.Match(nil2ctx(), []byte("MZ no PE header AB")); name != "" || err != nil {
			t.Fatalf("record %v, no PE header: (%q, %v), want clean", record, name, err)
		}

		d = compileCallback(t, headerReader("6", record))
		elf := elf64(binary.LittleEndian, 1, 0, nil, 13, 64)
		copy(elf[len(elf)-2:], "AB")
		res, err := d.MatchObject(nil2ctx(), elf, MatchContext{Type: TypeELF})
		if res.Name != "" || !errors.Is(err, ErrRuleIncomplete) {
			t.Fatalf("record %v, parsed ELF: (%q, %v), want incomplete", record, res.Name, err)
		}
	}
	// A program that does not read the header still decides the object.
	d := compileCallback(t, callbackFixture{
		kind: 256, logical: "Header.Ignorer;Engine:51-255,Target:1;0;4142", result: 32,
		locals: []int{32}, blocks: [][]string{{callbackReturn(32, 0)}},
	})
	if name, err := d.Match(nil2ctx(), brokenPEWith("AB")); name != "" || err != nil {
		t.Fatalf("header not read, unmodeled PE header: (%q, %v), want clean", name, err)
	}
}
