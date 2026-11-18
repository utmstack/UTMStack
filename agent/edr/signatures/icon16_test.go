package signatures

import (
	"encoding/binary"
	"testing"
)

func TestSixteenBitIconUsesReferenceColorAndMaskConversion(t *testing.T) {
	const width = 16
	bitmap := func(depth int) []byte {
		stride := (width*depth + 31) / 32 * 4
		b := make([]byte, 40+width*stride+width*4)
		binary.LittleEndian.PutUint32(b, 40)
		binary.LittleEndian.PutUint32(b[4:], width)
		binary.LittleEndian.PutUint32(b[8:], 2*width)
		binary.LittleEndian.PutUint16(b[12:], 1)
		binary.LittleEndian.PutUint16(b[14:], uint16(depth))
		return b
	}
	b16, b24 := bitmap(16), bitmap(24)
	ic := labIcon{w: width, bpp: 16, raw16: make([]uint16, width*width), mask: make([]bool, width*width)}
	for y := 0; y < width; y++ {
		for x := 0; x < width; x++ {
			lo, hi := uint32((x*17+y*13)&255), uint32((x*9+y*23)&255)
			p16 := 40 + (width-1-y)*width*2 + 2*x
			b16[p16], b16[p16+1] = byte(lo), byte(hi)
			ic.raw16[y*width+x] = uint16(lo | hi<<8)
			blue := lo & 0x1f
			green := (lo >> 5) | ((hi & 3) << 3)
			red := hi & 0xfc
			c := ((red<<3 | red>>2) << 17) |
				((green<<3 | green>>2) << 11) | (blue<<3 | blue>>2)
			bl, g, r := int(c&255), int((c>>8)&255), int((c>>16)&255)
			if (x+y)%5 == 0 {
				ic.mask[y*width+x] = true
				mask := 40 + width*width*2 + (width-1-y)*4 + x/8
				b16[mask] |= 0x80 >> (x % 8)
				a := int(c >> 24)
				bl, g, r = 255-a+a*bl/255, 255-a+a*g/255, 255-a+a*r/255
			}
			p24 := 40 + (width-1-y)*width*3 + 3*x
			b24[p24], b24[p24+1], b24[p24+2] = byte(bl), byte(g), byte(r)
		}
	}
	got, skip, sure := parseIconImage(b16)
	if !sure || skip {
		t.Fatalf("16-bit icon was not analyzed: skip=%v sure=%v", skip, sure)
	}
	want, skip, sure := parseIconImage(b24)
	if !sure || skip {
		t.Fatalf("24-bit comparison icon was not analyzed: skip=%v sure=%v", skip, sure)
	}
	if got != want {
		t.Fatalf("16-bit color or AND-mask conversion differs from reference pixels: got=%+v want=%+v", got, want)
	}
	// Pinned ClamAV 1.5.4, with an IDB rule and a logical icon-group trigger,
	// printed this exact hash for the generated 16-bit PE.
	const referenceHex = "109710a0494607088d4020968f00026ca0a0a6d00704e90408e60007e50804b60000c50908c70404910a047f08087a020624050026000030090025182648"
	reference, err := decodeIconHash(referenceHex)
	if err != nil {
		t.Fatal(err)
	}
	if got.hash != reference {
		t.Fatalf("16-bit icon hash differs from ClamAV 1.5.4: got=%+v want=%+v", got.hash, reference)
	}
	pe := labProgram([]labIcon{ic})
	icons := labIcons(t, pe)
	if len(icons) != 1 || icons[0] != got {
		t.Fatalf("PE resource scanner did not decode its 16-bit icon: icons=%+v", icons)
	}
	d := iconDatabase(t, "Probe.Icon:A:B:"+referenceHex+"\n",
		"Icon.Probe;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a")
	expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, "Icon.Probe", "")
	miss := iconDatabase(t, "Probe.Icon:A:B:"+iconMetrics+"\n",
		"Icon.Probe;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a")
	expectOutcome(t, miss, string(pe), MatchContext{Type: TypeMSEXE}, "", "")
	if _, skip, sure := parseIconImage(b16[:len(b16)-1]); !sure || !skip {
		t.Fatal("truncated 16-bit AND mask was not skipped")
	}
}
