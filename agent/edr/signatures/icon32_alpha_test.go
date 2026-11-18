package signatures

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestZeroAlphaIconUsesANDMaskOrOpaqueFallback(t *testing.T) {
	const width = 16
	ic := labIcon{w: width, bpp: 32, rgba: make([][4]byte, width*width), mask: make([]bool, width*width)}
	for y := 0; y < width; y++ {
		for x := 0; x < width; x++ {
			i := y*width + x
			ic.rgba[i] = [4]byte{byte(x * 15), byte(y * 13), byte((x*7 + y*11) & 255), 0}
			ic.mask[i] = (x+y)%5 == 0
		}
	}
	masked, skip, sure := parseIconImage(ic.bitmap())
	if !sure || skip {
		t.Fatalf("zero-alpha icon with an AND mask was not decided: skip=%v sure=%v", skip, sure)
	}
	base := ic.bitmap()
	withoutMask := base[:len(base)-width*((width+31)/32*4)]
	opaque, skip, sure := parseIconImage(withoutMask)
	if !sure || skip {
		t.Fatalf("zero-alpha icon without an AND mask was not decided: skip=%v sure=%v", skip, sure)
	}
	if masked == opaque {
		t.Fatal("the AND mask did not change the patterned icon hash")
	}
	// In the reference, a zero alpha channel makes the 32-bit color rows
	// behave like a 24-bit icon with the same RGB data and AND mask.
	equivalent := ic
	equivalent.bpp = 24
	want, skip, sure := parseIconImage(equivalent.bitmap())
	if !sure || skip || masked != want {
		t.Fatalf("masked zero-alpha icon differs from equivalent 24-bit color: masked=%+v want=%+v", masked, want)
	}
	equivalent.mask = nil
	wantOpaque, skip, sure := parseIconImage(equivalent.bitmap())
	if !sure || skip || opaque != wantOpaque {
		t.Fatalf("maskless zero-alpha icon differs from opaque 24-bit color: got=%+v want=%+v", opaque, wantOpaque)
	}
	pe := labProgram([]labIcon{ic})
	icons := labIcons(t, pe)
	if len(icons) != 1 || icons[0] != masked {
		t.Fatalf("PE resource scanner did not retain the zero-alpha icon: %+v", icons)
	}
	// Pinned ClamAV 1.5.4 reports this IDB metric string for the generated PE.
	const referenceHex = "109640a008ea000b716060022b0101289060633c090adf0a0bce060bc50a074a00006d00047304006306085e05045b02081d00002706002c000b36250834"
	reference, err := decodeIconHash(referenceHex)
	if err != nil {
		t.Fatal(err)
	}
	if masked.hash != reference {
		t.Fatalf("zero-alpha icon hash differs from ClamAV 1.5.4: got=%+v want=%+v", masked.hash, reference)
	}
	d := iconDatabase(t, "Probe.Icon:A:B:"+referenceHex+"\n",
		"Icon.ZeroAlpha;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a")
	expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, "Icon.ZeroAlpha", "")
	// Place a copy of the color rows at the end of the PE file, with no mask
	// bytes after them. The original resource group remains intact, while its
	// icon data entry points at the new maskless image.
	masklessPE := append(append([]byte(nil), pe...), withoutMask...)
	const sectionRaw = 0x200
	const sectionHeader = 0x138
	const iconDataEntry = sectionRaw + 128
	newRawSize := len(masklessPE) - sectionRaw
	binary.LittleEndian.PutUint32(masklessPE[sectionHeader+8:], uint32(newRawSize))
	binary.LittleEndian.PutUint32(masklessPE[sectionHeader+16:], uint32((newRawSize+0x1ff)&^0x1ff))
	binary.LittleEndian.PutUint32(masklessPE[iconDataEntry:], uint32(0x1000+len(pe)-sectionRaw))
	binary.LittleEndian.PutUint32(masklessPE[iconDataEntry+4:], uint32(len(withoutMask)))
	masklessIcons := labIcons(t, masklessPE)
	if len(masklessIcons) != 1 || masklessIcons[0] != opaque {
		t.Fatalf("PE resource scanner did not use the maskless opaque fallback: %+v", masklessIcons)
	}
	// The pinned reference logs "found a broken and stupid icon" and
	// forces opacity when its AND-mask read reaches the end of the file.
	const referenceOpaqueHex = "10b8f0a00ae5000b8ac0600328020237305063ed090ada0a0bbe060bb30a072000004a0004520400490a064406043f0a0b1900001b00061f0a0036250842"
	referenceOpaque, err := decodeIconHash(referenceOpaqueHex)
	if err != nil {
		t.Fatal(err)
	}
	if opaque.hash != referenceOpaque {
		t.Fatalf("maskless zero-alpha icon differs from ClamAV 1.5.4: got=%+v want=%+v", opaque.hash, referenceOpaque)
	}
	masklessRule := iconDatabase(t, "Probe.Opaque:A:B:"+referenceOpaqueHex+"\n",
		"Icon.ZeroAlphaNoMask;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a")
	expectOutcome(t, masklessRule, string(masklessPE), MatchContext{Type: TypeMSEXE}, "Icon.ZeroAlphaNoMask", "")
	if dir := os.Getenv("EDR_ICON_ALPHA_FIXTURE_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "zero-alpha-mask.exe"), pe, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "zero-alpha-no-mask.exe"), masklessPE, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}
