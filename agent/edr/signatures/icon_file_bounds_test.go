package signatures

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// ClamAV's icon parser reads an image from the mapped file, ignoring the
// resource data entry's declared image size. The v130 audit compares these
// inert PE fixtures with pinned ClamAV 1.5.4.
func TestIconImageFileBoundsReferenceModes(t *testing.T) {
	const flatHash = "1000000000000000000000000000000000000000000000000000000000000000000000040000080000000000000000000000000000040000080000000000"
	d := iconDatabase(t, "Win.Test.IconA:MS:IE:"+flatHash+"\n",
		"Icon.Live;Engine:51-255,Target:1,IconGroup1:MS;0;4d5a")
	flat := labIcon{w: 16, bpp: 24, rgba: make([][4]byte, 16*16)}
	cases := []struct {
		name   string
		icons  []labIcon
		change func(pe []byte, meta *peMeta, entryAt int)
		live   bool
	}{
		{"baseline", []labIcon{flat}, nil, true},
		{"declared-zero", []labIcon{flat}, func(pe []byte, _ *peMeta, at int) {
			binary.LittleEndian.PutUint32(pe[at+4:], 0)
		}, true},
		{"declared-one", []labIcon{flat}, func(pe []byte, _ *peMeta, at int) {
			binary.LittleEndian.PutUint32(pe[at+4:], 1)
		}, true},
		{"declared-huge", []labIcon{flat}, func(pe []byte, _ *peMeta, at int) {
			binary.LittleEndian.PutUint32(pe[at+4:], 0xffffffff)
		}, true},
		{"image-rva-unmapped", []labIcon{flat}, func(pe []byte, _ *peMeta, at int) {
			binary.LittleEndian.PutUint32(pe[at:], 0xffffffff)
		}, false},
		{"image-rva-last-eight", []labIcon{flat}, func(pe []byte, meta *peMeta, at int) {
			sec := meta.sections[0]
			binary.LittleEndian.PutUint32(pe[at:], sec.rva+uint32(len(pe)-8)-sec.raw)
		}, false},
		{"image-rva-header-only", []labIcon{flat}, func(pe []byte, meta *peMeta, at int) {
			sec := meta.sections[0]
			start := len(pe) - 40
			original, ok := meta.rawAddr(binary.LittleEndian.Uint32(pe[at:]))
			if !ok || original+40 > len(pe) {
				panic("test fixture icon header unavailable")
			}
			copy(pe[start:], pe[original:original+40])
			binary.LittleEndian.PutUint32(pe[at:], sec.rva+uint32(start)-sec.raw)
		}, false},
		{"first-unmapped-second-live", []labIcon{flat, flat}, func(pe []byte, _ *peMeta, at int) {
			binary.LittleEndian.PutUint32(pe[at:], 0xffffffff)
		}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pe := labProgram(tc.icons)
			meta := parsePE(pe)
			if meta == nil || len(meta.sections) != 1 {
				t.Fatal("PE resources missing")
			}
			entries, state := meta.resourcesByID(pe, 3, 1)
			if state != iconFound || len(entries) != 1 {
				t.Fatalf("icon resource lookup: state=%v entries=%d", state, len(entries))
			}
			at, ok := meta.rawAddr(entries[0])
			if !ok || at+16 > len(pe) {
				t.Fatal("icon data entry unavailable")
			}
			if tc.change != nil {
				tc.change(pe, meta, at)
			}
			want := ""
			if tc.live {
				want = "Icon.Live"
			}
			expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, want, "")
			if dir := os.Getenv("EDR_ICON_BOUNDS_FIXTURE_DIR"); dir != "" {
				if err := os.WriteFile(filepath.Join(dir, tc.name+".exe"), pe, 0o600); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
