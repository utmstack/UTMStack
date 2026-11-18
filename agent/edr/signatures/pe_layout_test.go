package signatures

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"
)

type peFixtureSection struct {
	virtualSize, virtualAddress, rawSize, rawPointer uint32
}

// peFixture builds a PE32 image with a 224-byte optional header and sixteen
// data directories. Only header fields the offset model reads are set.
type peFixture struct {
	size                                   int
	entry, sectionAlign, fileAlign, header uint32
	sections                               []peFixtureSection
}

func (f peFixture) build() []byte {
	b := make([]byte, f.size)
	o := binary.LittleEndian
	copy(b, "MZ")
	o.PutUint32(b[0x3c:], 0x80)
	copy(b[0x80:], "PE\x00\x00")
	o.PutUint16(b[0x84:], 0x14c)
	o.PutUint16(b[0x86:], uint16(len(f.sections)))
	o.PutUint16(b[0x94:], 224)
	opt := 0x98
	o.PutUint16(b[opt:], 0x10b)
	o.PutUint32(b[opt+16:], f.entry)
	o.PutUint32(b[opt+32:], f.sectionAlign)
	o.PutUint32(b[opt+36:], f.fileAlign)
	o.PutUint32(b[opt+56:], 0x10000)
	o.PutUint32(b[opt+60:], f.header)
	o.PutUint32(b[opt+92:], 16)
	for i, s := range f.sections {
		h := b[0x178+40*i:]
		copy(h, fmt.Sprintf(".s%d", i))
		o.PutUint32(h[8:], s.virtualSize)
		o.PutUint32(h[12:], s.virtualAddress)
		o.PutUint32(h[16:], s.rawSize)
		o.PutUint32(h[20:], s.rawPointer)
	}
	return b
}

const peLayoutMarker = "UTMSTACK-PE-LAYOUT-MARKER"

func peLayoutRules(t *testing.T, name, body string) *Database {
	t.Helper()
	d, err := Compile(map[string][]byte{name: []byte(body)})
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// sig-1: a virtual-only section (SizeOfRawData 0) is never file-backed, so its
// PointerToRawData cannot invalidate the header. EP, section and whole-file PE
// rules and .text section hashes must still apply.
func TestVirtualOnlySectionPointerKeepsPESignatures(t *testing.T) {
	marker := hex.EncodeToString([]byte(peLayoutMarker))
	for _, pointer := range []uint32{0, 0xFFFFF000} {
		data := peFixture{size: 0x400, entry: 0x1000, sectionAlign: 0x1000, fileAlign: 0x200, header: 0x200, sections: []peFixtureSection{
			{0x200, 0x1000, 0x200, 0x200},
			{0x100, 0x2000, 0, pointer},
		}}.build()
		copy(data[0x200:], peLayoutMarker)
		sum := md5.Sum(data[0x200:0x400])
		for name, body := range map[string]string{
			"ep.ndb":  "PE.EP:1:EP+0:" + marker,
			"any.ndb": "PE.Any:1:*:" + marker,
			"s0.ndb":  "PE.S0:1:S0+0:" + marker,
			"ep.ldb":  "PE.Logic;Engine:51-255,Target:1,NumberOfSections:2-2,EntryPoint:512-512;0;" + marker,
			"sec.mdb": fmt.Sprintf("512:%x:PE.Section", sum),
		} {
			got, err := peLayoutRules(t, name, body).Match(context.Background(), data)
			if got == "" || err != nil {
				t.Errorf("pointer %#x %s: got %q, %v; want a detection", pointer, name, got, err)
			}
		}
		// The last section is virtual-only: SL+n starts at its aligned pointer
		// (offset 0 or past the end of the file), not at the .text data.
		got, err := peLayoutRules(t, "sl.ndb", "PE.SL:1:SL+0:"+marker).Match(context.Background(), data)
		if got != "" || err != nil {
			t.Errorf("pointer %#x SL+0: got %q, %v; want a decided non-match", pointer, got, err)
		}
	}
}

// A raw section that starts beyond the end of the file is treated as having no
// file data, like the reference parser's non-strict mode; it does not make the
// rest of the header unusable.
func TestMissingRawSectionKeepsPESignatures(t *testing.T) {
	marker := hex.EncodeToString([]byte(peLayoutMarker))
	data := peFixture{size: 0x400, entry: 0x1000, sectionAlign: 0x1000, fileAlign: 0x200, header: 0x200, sections: []peFixtureSection{
		{0x200, 0x1000, 0x200, 0x200},
		{0x200, 0x2000, 0x200, 0x10000},
	}}.build()
	copy(data[0x200:], peLayoutMarker)
	if got, err := peLayoutRules(t, "ep.ndb", "PE.EP:1:EP+0:"+marker).Match(context.Background(), data); got != "PE.EP" || err != nil {
		t.Fatalf("got %q, %v", got, err)
	}
}

// sig-3: offsets and section hashes use aligned PointerToRawData and
// SizeOfRawData, the aligned header size, and the 0x200 fallback alignment.
func TestPESectionAlignment(t *testing.T) {
	marker := hex.EncodeToString([]byte(peLayoutMarker))
	aligned := peFixture{size: 0x400, entry: 0x1000, sectionAlign: 0x1000, fileAlign: 0x200, header: 0x200, sections: []peFixtureSection{{0x200, 0x1000, 0x200, 0x200}}}.build()
	copy(aligned[0x200:], peLayoutMarker)
	sum := md5.Sum(aligned[0x200:0x400]) // the reference hashes raw..raw+aligned size
	files := map[string]string{
		"ep.ndb":  "PE.EP:1:EP+0:" + marker,
		"s0.ndb":  "PE.S0:1:S0+0:" + marker,
		"se.ndb":  "PE.SE:1:SE0:" + marker,
		"sec.mdb": fmt.Sprintf("512:%x:PE.Section", sum),
	}
	for name, body := range files {
		d := peLayoutRules(t, name, body)
		for _, s := range []peFixtureSection{{0x200, 0x1000, 0x200, 0x200}, {0x200, 0x1000, 0x1ff, 0x201}} {
			data := peFixture{size: 0x400, entry: 0x1000, sectionAlign: 0x1000, fileAlign: 0x200, header: 0x200, sections: []peFixtureSection{s}}.build()
			copy(data[0x200:], peLayoutMarker)
			if got, err := d.Match(context.Background(), data); got == "" || err != nil {
				t.Errorf("%s PointerToRawData=%#x SizeOfRawData=%#x: got %q, %v; want a detection", name, s.rawPointer, s.rawSize, got, err)
			}
		}
	}
}

// A section pointer that is not a multiple of FileAlignment but is a multiple of
// 0x200 switches the whole image to 0x200 alignment, and a truncated section is
// clamped to the end of the file instead of invalidating the header.
func TestPEFileAlignmentFallbackAndTruncation(t *testing.T) {
	marker := hex.EncodeToString([]byte(peLayoutMarker))
	fallback := peFixture{size: 0x400, entry: 0x1000, sectionAlign: 0x1000, fileAlign: 0x1000, header: 0x200, sections: []peFixtureSection{{0x200, 0x1000, 0x1f0, 0x200}}}.build()
	copy(fallback[0x200:], peLayoutMarker)
	truncated := peFixture{size: 0x300, entry: 0x1000, sectionAlign: 0x1000, fileAlign: 0x200, header: 0x200, sections: []peFixtureSection{{0x400, 0x1000, 0x400, 0x200}}}.build()
	copy(truncated[0x200:], peLayoutMarker)
	for name, tc := range map[string]struct {
		data []byte
		size int
	}{"fallback": {fallback, 0x200}, "truncated": {truncated, 0x100}} {
		sum := md5.Sum(tc.data[0x200 : 0x200+tc.size])
		for file, body := range map[string]string{
			"ep.ndb":  "PE.EP:1:EP+0:" + marker,
			"sec.mdb": fmt.Sprintf("%d:%x:PE.Section", tc.size, sum),
		} {
			if got, err := peLayoutRules(t, file, body).Match(context.Background(), tc.data); got == "" || err != nil {
				t.Errorf("%s %s: got %q, %v; want a detection", name, file, got, err)
			}
		}
	}
}

// An entry point below the section-aligned header size maps to the same file
// offset, even when it is past the unaligned SizeOfHeaders.
func TestPEEntryPointInAlignedHeader(t *testing.T) {
	marker := hex.EncodeToString([]byte(peLayoutMarker))
	data := peFixture{size: 0x600, entry: 0x300, sectionAlign: 0x1000, fileAlign: 0x200, header: 0x200, sections: []peFixtureSection{{0x200, 0x1000, 0x200, 0x400}}}.build()
	copy(data[0x300:], peLayoutMarker)
	if got, err := peLayoutRules(t, "ep.ndb", "PE.EP:1:EP+0:"+marker).Match(context.Background(), data); got != "PE.EP" || err != nil {
		t.Fatalf("got %q, %v", got, err)
	}
}

// When the entry point cannot map to a file offset, the reference rejects the
// PE header. Header-independent raw rules still run.
func TestUnmappedPEEntryPointRejectsHeaderRules(t *testing.T) {
	marker := hex.EncodeToString([]byte(peLayoutMarker))
	// The entry point lies in a virtual-only section, so it has no file offset.
	data := peFixture{size: 0x400, entry: 0x2000, sectionAlign: 0x1000, fileAlign: 0x200, header: 0x200, sections: []peFixtureSection{
		{0x200, 0x1000, 0x200, 0x200},
		{0x100, 0x2000, 0, 0},
	}}.build()
	copy(data[0x200:], peLayoutMarker)
	sum := md5.Sum(data[0x200:0x400])
	for name, body := range map[string]string{
		"ep.ndb":  "PE.EP:1:EP+0:" + marker,
		"s0.ndb":  "PE.S0:1:S0+0:" + marker,
		"nos.ldb": "PE.Sections;Engine:51-255,Target:1,NumberOfSections:2-2;0;" + marker,
		"sec.mdb": fmt.Sprintf("512:%x:PE.Section", sum),
	} {
		got, err := peLayoutRules(t, name, body).Match(context.Background(), data)
		if got != "" || err != nil {
			t.Errorf("%s: got %q, %v; want clean", name, got, err)
		}
	}
	d, err := Compile(map[string][]byte{"a.ndb": []byte("PE.EP:1:EP+0:" + marker), "b.ndb": []byte("PE.Any:1:*:" + marker)})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := d.Match(context.Background(), data); got != "PE.Any" || err != nil {
		t.Fatalf("header-independent rule: got %q, %v", got, err)
	}
	// A rule already decided without the header is not made incomplete.
	if got, err := peLayoutRules(t, "and.ldb", "PE.And;Target:1;0&1;"+marker+";EP+0:"+marker).Match(context.Background(), []byte("MZ no header")); got != "" || err != nil {
		t.Fatalf("decided rule: got %q, %v", got, err)
	}
	// Target 1 applies only to MZ objects.
	if got, err := peLayoutRules(t, "ep.ndb", "PE.EP:1:EP+0:"+marker).Match(context.Background(), []byte("plain "+peLayoutMarker)); got != "" || err != nil {
		t.Fatalf("non-executable: got %q, %v", got, err)
	}
}
