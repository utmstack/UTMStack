package signatures

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"testing"
)

func TestBMPPlainFuzzyImageReference(t *testing.T) {
	for _, tc := range []struct {
		name string
		hash uint64
	}{
		{"core1", 0xed83a7d05aa98c17},
		{"core4", 0x952a752a752a772a},
		{"core8", 0x86d227588f171af5},
		{"core24", 0xf7ef51ea4ca04170},
		{"info2_gap", 0xd5aa12aac0ef90eb},
		{"info8_shortpal", 0x9f474ca59d78c059},
		{"info24_gap", 0xf7ef51ea4ca04170},
		{"v2_24", 0xf7ef51ea4ca04170},
		{"v3_8_top", 0x86d227588f171af5},
		{"v5_32_alpha", 0xf7ef51ea4ca04170},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile("testdata/bmp_plain_" + tc.name + ".bmp")
			if err != nil {
				t.Fatal(err)
			}
			ih := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if !ih.ok || ih.hash != tc.hash || ih.uncertain != 0 {
				t.Fatalf("plain BMP hash %+v, reference %016x", ih, tc.hash)
			}
			rule := compileTest(t, "ldb", fmt.Sprintf("BMP.Plain;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", tc.hash))
			expectOutcome(t, rule, string(data), MatchContext{Type: TypeGraphics}, "BMP.Plain", "")
			negative := compileTest(t, "ldb", fmt.Sprintf("BMP.Other;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", tc.hash^1<<40))
			expectOutcome(t, negative, string(data), MatchContext{Type: TypeGraphics}, "", "")
		})
	}
}

func TestBMPPlainMalformedRemainsIncomplete(t *testing.T) {
	original, err := os.ReadFile("testdata/bmp_plain_info2_gap.bmp")
	if err != nil {
		t.Fatal(err)
	}
	rule := compileTest(t, "ldb", "BMP.Plain;Engine:150-255,Target:0;0;fuzzy_img#d5aa12aac0ef90eb\n")
	for _, tc := range []struct {
		name string
		edit func([]byte) []byte
	}{
		{"truncated-pixels", func(b []byte) []byte { return b[:len(b)-1] }},
		{"palette-too-large", func(b []byte) []byte {
			binary.LittleEndian.PutUint32(b[46:50], 5)
			return b
		}},
		{"oversize-width", func(b []byte) []byte {
			binary.LittleEndian.PutUint32(b[18:22], 1<<20)
			return b
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := tc.edit(bytes.Clone(original))
			if ih := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40}); ih.ok {
				t.Fatalf("malformed BMP unexpectedly hashed: %+v", ih)
			}
			expectOutcome(t, rule, string(data), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
		})
	}
}
