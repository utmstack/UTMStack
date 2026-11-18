package signatures

import (
	"encoding/hex"
	"testing"
)

func TestBinaryByteCompareUsesLastTriggerAndEndian(t *testing.T) {
	big := compileTest(t, "ldb", "Cmp.Big;Engine:100-255,Target:0;1;414243;0(>>3#ib2#>512)\n")
	for _, tc := range []struct{ data, want string }{
		{"nothing", ""},
		{"xxABC\x02\x01", "Cmp.Big"},
		{"xxABC\x01\xff", ""},
		// The reference reads from the last trigger, not the first.
		{"xABC\x02\x01yABC\x01\xff", ""},
		{"xABC\x01\xffyABC\x02\x01", "Cmp.Big"},
	} {
		expectOutcome(t, big, tc.data, anyType, tc.want, "")
	}
	little := compileTest(t, "ldb", "Cmp.Little;Engine:100-255,Target:0;1;414243;0(>>3#il2#=513)\n")
	expectOutcome(t, little, "xxABC\x01\x02", anyType, "Cmp.Little", "")
	expectOutcome(t, little, "xxABC\x02\x01", anyType, "", "")
	little4 := compileTest(t, "ldb", "Cmp.Little4;Engine:100-255,Target:0;1;414243;0(>>3#il4#=67305985)\n")
	expectOutcome(t, little4, "xxABC\x01\x02\x03\x04", anyType, "Cmp.Little4", "")
	big8 := compileTest(t, "ldb", "Cmp.Big8;Engine:100-255,Target:0;1;414243;0(>>3#ib8#<0)\n")
	expectOutcome(t, big8, "xxABC\xff\xff\xff\xff\xff\xff\xff\xff", anyType, "Cmp.Big8", "")
	before := compileTest(t, "ldb", "Cmp.Before;Engine:100-255,Target:0;1;414243;0(<<2#ib2#=257)\n")
	expectOutcome(t, before, "x\x01\x01ABC", anyType, "Cmp.Before", "")
	and := compileTest(t, "ldb", "Cmp.And;Engine:100-255,Target:0;1;414243;0(>>3#ib2#>512,<514)\n")
	expectOutcome(t, and, "xxABC\x02\x01", anyType, "Cmp.And", "")
	expectOutcome(t, and, "xxABC\x02\x03", anyType, "", "")
}

func TestByteCompareBoundaryAndUnsupportedForms(t *testing.T) {
	// ClamAV 1.5.4's implementation rejects a read beginning at offset 0.
	boundary := compileTest(t, "ldb", "Cmp.Zero;Engine:100-255,Target:0;1;414243;0(<<1#ib1#=88)\n")
	expectOutcome(t, boundary, "XABC", anyType, "", "")
	short := compileTest(t, "ldb", "Cmp.Short;Engine:100-255,Target:0;1;414243;0(>>3#ib2#=513)\n")
	expectOutcome(t, short, "xxABC\x02", anyType, "", "")
	unsupported := compileTest(t, "ldb", "Cmp.Unsupported;Engine:100-255,Target:0;1;414243;0(>>3#ib3#=2)\n")
	expectOutcome(t, unsupported, "xxABC02", anyType, "", UnknownByteCompare)
	zero := compileTest(t, "ldb", "Cmp.Same;Engine:100-255,Target:0;1;414243;0(0#ib1#=65)\n")
	expectOutcome(t, zero, "xxABCx", anyType, "Cmp.Same", "")
	zeroEmpty := compileTest(t, "ldb", "Cmp.EmptyOffset;Engine:100-255,Target:0;1;414243;0(#ib1#=65)\n")
	expectOutcome(t, zeroEmpty, "xxABCx", anyType, "Cmp.EmptyOffset", "")
	negativeRight := compileTest(t, "ldb", "Cmp.NegRight;Engine:100-255,Target:0;1;414243;0(>>-1#ib1#=120)\n")
	expectOutcome(t, negativeRight, "xxABCxx", anyType, "Cmp.NegRight", "")
	negativeLeft := compileTest(t, "ldb", "Cmp.NegLeft;Engine:100-255,Target:0;1;414243;0(<<-1#ib1#=66)\n")
	expectOutcome(t, negativeLeft, "xxABCxx", anyType, "Cmp.NegLeft", "")
}

func TestASCIIByteCompareMatchesPinnedReference(t *testing.T) {
	for _, tc := range []struct {
		name, sub, data, want string
	}{
		{"hex", ">>3#hb2#=2", "xxABC02", "Cmp.ASCII"},
		{"hex-prefix", ">>3#hb4#=42", "xxABC0x2a", "Cmp.ASCII"},
		{"hex-prefix-nonexact", ">>3#hb3#=2", "xxABC02x", "Cmp.ASCII"},
		{"hex-prefix-exact", ">>3#hbe3#=2", "xxABC02x", ""},
		{"hex-invalid-is-zero", ">>3#hb3#=0", "xxABCxyz", "Cmp.ASCII"},
		{"hex-whitespace", ">>3#hb4#=42", "xxABC  2a", "Cmp.ASCII"},
		{"hex-whitespace-exact", ">>3#hbe4#=42", "xxABC  2a", ""},
		{"decimal", ">>3#db2#=42", "xxABC42", "Cmp.ASCII"},
		{"decimal-prefix", ">>3#db3#=42", "xxABC42x", "Cmp.ASCII"},
		{"decimal-prefix-exact", ">>3#dbe3#=42", "xxABC42x", ""},
		{"decimal-signed", ">>3#db3#=-42", "xxABC-42", "Cmp.ASCII"},
		{"decimal-invalid-is-zero", ">>3#db3#=0", "xxABCxyz", "Cmp.ASCII"},
		{"auto-short", ">>3#ab2#=42", "xxABC42", ""},
		{"auto-decimal", ">>3#ab3#=42", "xxABC042", "Cmp.ASCII"},
		{"auto-hex", ">>3#ab4#=42", "xxABC0x2a", "Cmp.ASCII"},
		{"hex-little", ">>3#hl4#=42", "xxABC2a00", "Cmp.ASCII"},
		{"hex-little-prefix", ">>3#hl4#=42", "xxABC0x2a", "Cmp.ASCII"},
		{"hex-little-odd", ">>3#hl3#=42", "xxABCa02", ""},
		{"hex-little-reset", ">>3#hl3#=0", "xxABCxyz", "Cmp.ASCII"},
		{"auto-little-decimal", ">>3#al4#=42", "xxABC0042", ""},
		{"auto-little-hex", ">>3#al4#=42", "xxABC0x2a", "Cmp.ASCII"},
		{"hex-no-order", ">>3#h2#=2", "xxABC02", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := compileTest(t, "ldb", "Cmp.ASCII;Engine:100-255,Target:0;1;414243;0("+tc.sub+")\n")
			if got := d.Report().Unknown[UnknownByteCompare]; got != 0 {
				t.Fatalf("supported form is unknown: %v", d.Report().Unknown)
			}
			expectOutcome(t, d, tc.data, anyType, tc.want, "")
		})
	}
}

func TestBinaryByteCompareReportTracksOnlyUnsupportedForms(t *testing.T) {
	d := compileTest(t, "ldb", "Cmp.Binary;Engine:100-255,Target:0;1;414243;0(>>3#ib2#>512)\n"+
		"Cmp.Unsupported;Engine:100-255,Target:0;1;414243;0(>>3#ib3#>512)\n")
	if got := d.Report().Unknown[UnknownByteCompare]; got != 1 {
		t.Fatalf("unknown byte comparisons = %d, want one unsupported form", got)
	}
}

func TestPinnedDailyByteCompareRule(t *testing.T) {
	const name = "Multios.Exploit.CVE_2022_3602-9976476-1"
	d := compileTest(t, "ldb", name+";Engine:100-255,Target:0;1;0603551d1e0101??0482{2}3082{2}(a0|a1)82{2}3082{2}8182{2}786e2d2d;0(>>26#ib2#>512)\n")
	if got := d.Report().Unknown[UnknownByteCompare]; got != 0 {
		t.Fatalf("selected daily rule is still incomplete: %v", d.Report().Unknown)
	}
	for _, tc := range []struct{ field, want string }{{"0201", name}, {"01ff", ""}} {
		body, err := hex.DecodeString("0603551d1e0101000482000030820000a0820000308200008182" + tc.field + "786e2d2d")
		if err != nil {
			t.Fatal(err)
		}
		expectOutcome(t, d, "xx"+string(body)+"zz", anyType, tc.want, "")
	}
}
