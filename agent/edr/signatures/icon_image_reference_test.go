package signatures

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// These cases are matched against the pinned ClamAV 1.5.4 scanner in the
// adjacent v129 audit. The files are inert resource-only PE fixtures.
func TestIconImageHeaderReferenceModes(t *testing.T) {
	const flatHash = "1000000000000000000000000000000000000000000000000000000000000000000000040000080000000000000000000000000000040000080000000000"
	d := iconDatabase(t, "Win.Test.IconA:MS:IE:"+flatHash+"\n",
		"Icon.Live;Engine:51-255,Target:1,IconGroup1:MS;0;4d5a")
	base := labIcon{w: 16, bpp: 24, rgba: make([][4]byte, 16*16)}
	cases := []struct {
		name   string
		icon   labIcon
		change func(group, image []byte)
		live   bool
	}{
		{"baseline", base, nil, true},
		{"group-width-mismatch", base, func(g, _ []byte) { g[6] = 17 }, true},
		{"group-height-mismatch", base, func(g, _ []byte) { g[7] = 17 }, true},
		{"small-header", base, func(_, b []byte) { binary.LittleEndian.PutUint32(b, 39) }, false},
		{"small-width", base, func(_, b []byte) { binary.LittleEndian.PutUint32(b[4:], 8) }, false},
		{"large-height", base, func(_, b []byte) { binary.LittleEndian.PutUint32(b[8:], 1024) }, false},
		{"skewed-aspect", base, func(_, b []byte) { binary.LittleEndian.PutUint32(b[8:], 64) }, false},
		{"unsupported-depth", base, func(_, b []byte) { binary.LittleEndian.PutUint16(b[14:], 7) }, false},
		{"compression-one", base, func(_, b []byte) { binary.LittleEndian.PutUint32(b[16:], 1) }, true},
		{"colors-used-one", base, func(_, b []byte) { binary.LittleEndian.PutUint32(b[32:], 1) }, true},
		{"odd-height", base, func(_, b []byte) { binary.LittleEndian.PutUint32(b[8:], 33) }, true},
		{"png-signature", base, func(_, b []byte) { copy(b, pngSignature) }, false},
		{"extended-header", labIcon{w: 16, bpp: 24, headerExtra: 68, rgba: make([][4]byte, 16*16)}, nil, true},
		{"rectangular-16x18", labIcon{w: 16, h: 18, bpp: 24, rgba: make([][4]byte, 16*18)}, nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pe := labProgram([]labIcon{tc.icon})
			meta := parsePE(pe)
			if meta == nil {
				t.Fatal("PE resources missing")
			}
			var group []byte
			if !meta.resources(pe, 14, func(entry uint32) bool {
				group, _ = meta.resourceData(pe, entry)
				return true
			}) {
				t.Fatal("group resource missing")
			}
			entries, state := meta.resourcesByID(pe, 3, 1)
			if state != iconFound || len(entries) != 1 {
				t.Fatalf("icon resource lookup: %v, %d", state, len(entries))
			}
			image, ok := meta.resourceData(pe, entries[0])
			if !ok {
				t.Fatal("icon resource data missing")
			}
			if tc.change != nil {
				tc.change(group, image)
			}
			want := ""
			if tc.live {
				want = "Icon.Live"
			}
			expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, want, "")
			if dir := os.Getenv("EDR_ICON_IMAGE_FIXTURE_DIR"); dir != "" {
				if err := os.WriteFile(filepath.Join(dir, tc.name+".exe"), pe, 0o600); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestIconImageWideBitmapReferenceModes(t *testing.T) {
	for _, tc := range []struct {
		name  string
		width int
		hash  string
	}{
		{"wide-192", 192, "18000000000000000000000000000000000000000000000000000000000000000000000600000c00000000000000000000000000000600000c0000000000"},
		{"wide-256", 256, "2000000000000000000000000000000000000000000000000000000000000000000000080000100000000000000000000000000000080000100000000000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ic := labIcon{w: tc.width, bpp: 24, rgba: make([][4]byte, tc.width*tc.width)}
			icon, skip, sure := parseIconImage(ic.bitmap())
			if !sure || skip || icon.hash.width != tc.width/8 {
				t.Fatalf("wide icon parse: sure=%t skip=%t width=%d", sure, skip, icon.hash.width)
			}
			reference, err := decodeIconHash(tc.hash)
			if err != nil || icon.hash != reference {
				t.Fatalf("wide icon hash differs from ClamAV: err=%v got=%+v want=%+v", err, icon.hash, reference)
			}
			pe := labProgram([]labIcon{ic})
			meta := parsePE(pe)
			icons, ok := meta.programIcons(pe)
			if !ok || len(icons) != 1 || icons[0] != icon {
				t.Fatalf("wide icon resources: ok=%t count=%d", ok, len(icons))
			}
			if dir := os.Getenv("EDR_ICON_IMAGE_FIXTURE_DIR"); dir != "" {
				if err := os.WriteFile(filepath.Join(dir, tc.name+".exe"), pe, 0o600); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestIconImageRectangularColorHash(t *testing.T) {
	const width, height = 18, 16
	ic := labIcon{w: width, h: height, bpp: 24, rgba: make([][4]byte, width*height)}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			ic.rgba[y*width+x] = [4]byte{
				byte((x*13 + y*7) & 255),
				byte((x*3 + y*19) & 255),
				byte((x*23 + y*11) & 255), 255,
			}
		}
	}
	icon, skip, sure := parseIconImage(ic.bitmap())
	if !sure || skip || icon.hash.width != 16 {
		t.Fatalf("rectangular color icon parse: sure=%t skip=%t width=%d", sure, skip, icon.hash.width)
	}
	const referenceHex = "10cf30a0bc6a050bb5907003e800004ee01046940108ea0a0be40409e00a07360000710004a60400600309550a0a4607011800001d0a041e000427162659"
	reference, err := decodeIconHash(referenceHex)
	if err != nil || icon.hash != reference {
		t.Fatalf("rectangular color icon hash differs from ClamAV: err=%v got=%+v want=%+v", err, icon.hash, reference)
	}
	pe := labProgram([]labIcon{ic})
	meta := parsePE(pe)
	icons, ok := meta.programIcons(pe)
	if !ok || len(icons) != 1 || icons[0] != icon {
		t.Fatalf("rectangular color icon resources: ok=%t count=%d", ok, len(icons))
	}
	d := iconDatabase(t, "Win.Test.Color:A:B:"+referenceHex+"\n",
		"Icon.Color;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a")
	expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, "Icon.Color", "")
	if dir := os.Getenv("EDR_ICON_IMAGE_FIXTURE_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "rectangular-color.exe"), pe, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}
