package signatures

import (
	"encoding/binary"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
)

func TestPESectionsAreTheReferenceRecords(t *testing.T) {
	pe := peWithResources(14, 0x0102)
	le := binary.LittleEndian
	// The second section's raw data runs past the end of the 2 KiB file and
	// its virtual size is zero.
	le.PutUint32(pe[416+8:], 0)
	le.PutUint32(pe[416+16:], 0x800)
	le.PutUint32(pe[416+36:], 0x40000040)
	m := parsePE(pe)
	if m == nil || !m.valid {
		t.Fatal("header not parsed")
	}
	info := m.bytecodePE()
	want := []bytecode.Section{
		{RVA: 0x1000, VSize: 0x1000, Raw: 0x200, RSize: 0x200, URVA: 0x1000, UVSize: 0x200, URaw: 0x200, URSize: 0x200},
		{RVA: 0x2000, VSize: 0x1000, Raw: 0x400, RSize: 0x400, Characteristics: 0x40000040, URVA: 0x2000, UVSize: 0, URaw: 0x400, URSize: 0x400},
	}
	if info.EntryPoint != 0x200 || info.HeaderSize != 0x1000 || len(info.Sections) != 2 {
		t.Fatalf("entry %#x headers %#x sections %d", info.EntryPoint, info.HeaderSize, len(info.Sections))
	}
	for i, s := range want {
		if info.Sections[i] != s {
			t.Errorf("section %d = %+v, want %+v", i, info.Sections[i], s)
		}
	}
	// A section starting past the end of the file keeps its record with no
	// raw data.
	le.PutUint32(pe[416+20:], 0x900)
	if s := parsePE(pe).bytecodePE().Sections[1]; s.RSize != 0 || s.URSize != 0 || s.Raw != 0x800 || s.URaw != 0x900 {
		t.Errorf("section past the end = %+v", s)
	}
}
